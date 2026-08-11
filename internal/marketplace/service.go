package marketplace

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/gestorbuy/api/internal/marketplace/mercadolivre"
	"github.com/gestorbuy/api/internal/platform/crypto"
	"github.com/gestorbuy/api/internal/product"
)

var (
	ErrNotConfigured = errors.New("marketplace: integração não configurada no servidor (faltam ML_CLIENT_ID/ML_CLIENT_SECRET/ML_REDIRECT_URI)")
	ErrNotConnected  = errors.New("marketplace: tenant não conectou esse marketplace ainda")
	ErrInvalidState  = errors.New("marketplace: state inválido ou expirado")
)

// stateClaims carrega o tenant_id através do redirect OAuth — o callback do
// Mercado Livre chega como um GET público (o navegador do usuário, sem
// nenhum header nosso), então o tenant precisa viajar dentro do próprio
// `state` que a gente mandou na URL de autorização, assinado pra não ser
// forjável.
type stateClaims struct {
	TenantID string `json:"tenant_id"`
	jwt.RegisteredClaims
}

const statePurpose = "ml_oauth_state"
const stateTTL = 10 * time.Minute

type Service struct {
	repo       *Repository
	products   *product.Service
	ml         *mercadolivre.Client
	cipher     *crypto.Cipher
	stateKey   []byte
	configured bool
}

// New monta o serviço. clientID/clientSecret/redirectURI/encryptionKey
// vazios são aceitos (ambiente ainda sem o app do Mercado Livre criado) —
// o serviço sobe normalmente, só recusa os métodos que dependem disso com
// ErrNotConfigured, em vez de derrubar o boot inteiro da aplicação.
func New(repo *Repository, products *product.Service, jwtSecret, mlClientID, mlClientSecret, mlRedirectURI, encryptionKeyBase64 string) (*Service, error) {
	configured := mlClientID != "" && mlClientSecret != "" && mlRedirectURI != "" && encryptionKeyBase64 != ""

	svc := &Service{
		repo:       repo,
		products:   products,
		ml:         mercadolivre.NewClient(mlClientID, mlClientSecret, mlRedirectURI),
		stateKey:   []byte(jwtSecret),
		configured: configured,
	}

	if encryptionKeyBase64 != "" {
		cipher, err := crypto.New(encryptionKeyBase64)
		if err != nil {
			return nil, fmt.Errorf("marketplace: TOKEN_ENCRYPTION_KEY inválida: %w", err)
		}
		svc.cipher = cipher
	}

	return svc, nil
}

func (s *Service) ConnectURL(tenantID string) (string, error) {
	if !s.configured {
		return "", ErrNotConfigured
	}

	now := time.Now()
	claims := stateClaims{
		TenantID: tenantID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   statePurpose,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(stateTTL)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	state, err := token.SignedString(s.stateKey)
	if err != nil {
		return "", fmt.Errorf("marketplace: assinar state: %w", err)
	}

	return s.ml.AuthorizationURL(state), nil
}

func (s *Service) parseState(raw string) (string, error) {
	claims := &stateClaims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("marketplace: método de assinatura inesperado: %v", t.Header["alg"])
		}
		return s.stateKey, nil
	})
	if err != nil || !token.Valid || claims.Subject != statePurpose {
		return "", ErrInvalidState
	}
	return claims.TenantID, nil
}

// HandleCallback troca o code pelo par de tokens e persiste (cifrado). Devolve
// o tenant_id decodificado do state, pro handler saber pra onde redirecionar.
func (s *Service) HandleCallback(ctx context.Context, code, state string) (string, error) {
	if !s.configured {
		return "", ErrNotConfigured
	}

	tenantID, err := s.parseState(state)
	if err != nil {
		return "", err
	}

	tr, err := s.ml.ExchangeCode(ctx, code)
	if err != nil {
		return "", fmt.Errorf("marketplace: trocar code por token: %w", err)
	}

	accessTokenEnc, err := s.cipher.Encrypt(tr.AccessToken)
	if err != nil {
		return "", fmt.Errorf("marketplace: cifrar access_token: %w", err)
	}
	refreshTokenEnc, err := s.cipher.Encrypt(tr.RefreshToken)
	if err != nil {
		return "", fmt.Errorf("marketplace: cifrar refresh_token: %w", err)
	}

	cred := Credential{
		ID:              bson.NewObjectID().Hex(),
		TenantID:        tenantID,
		Marketplace:     MarketplaceMercadoLivre,
		AccessTokenEnc:  accessTokenEnc,
		RefreshTokenEnc: refreshTokenEnc,
		ExpiresAt:       time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second),
		ExternalUserID:  fmt.Sprintf("%d", tr.UserID),
	}
	if err := s.repo.Upsert(ctx, cred); err != nil {
		return "", err
	}

	return tenantID, nil
}

type Status struct {
	Marketplace string `json:"marketplace"`
	Connected   bool   `json:"connected"`
	ConnectedAt string `json:"connected_at,omitempty"`
}

func (s *Service) Status(ctx context.Context, tenantID string) ([]Status, error) {
	cred, err := s.repo.FindByTenant(ctx, tenantID, MarketplaceMercadoLivre)
	if err != nil {
		return nil, err
	}
	if cred == nil {
		return []Status{{Marketplace: MarketplaceMercadoLivre, Connected: false}}, nil
	}
	return []Status{{
		Marketplace: MarketplaceMercadoLivre,
		Connected:   true,
		ConnectedAt: cred.ConnectedAt.Format(time.RFC3339),
	}}, nil
}

func (s *Service) Disconnect(ctx context.Context, tenantID string) error {
	deleted, err := s.repo.Delete(ctx, tenantID, MarketplaceMercadoLivre)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrNotConnected
	}
	return nil
}

// EnsureValidToken devolve um access_token utilizável, refrescando primeiro
// se estiver perto de expirar (janela de 5 min). Estratégia lazy —
// refresca sob demanda, não tem worker em background nesta entrega.
func (s *Service) EnsureValidToken(ctx context.Context, tenantID string) (string, error) {
	if !s.configured {
		return "", ErrNotConfigured
	}

	cred, err := s.repo.FindByTenant(ctx, tenantID, MarketplaceMercadoLivre)
	if err != nil {
		return "", err
	}
	if cred == nil {
		return "", ErrNotConnected
	}

	if time.Until(cred.ExpiresAt) > 5*time.Minute {
		return s.cipher.Decrypt(cred.AccessTokenEnc)
	}

	refreshToken, err := s.cipher.Decrypt(cred.RefreshTokenEnc)
	if err != nil {
		return "", err
	}
	tr, err := s.ml.RefreshToken(ctx, refreshToken)
	if err != nil {
		return "", fmt.Errorf("marketplace: refresh token: %w", err)
	}

	accessTokenEnc, err := s.cipher.Encrypt(tr.AccessToken)
	if err != nil {
		return "", err
	}
	refreshTokenEnc, err := s.cipher.Encrypt(tr.RefreshToken)
	if err != nil {
		return "", err
	}

	cred.AccessTokenEnc = accessTokenEnc
	cred.RefreshTokenEnc = refreshTokenEnc
	cred.ExpiresAt = time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second)
	if err := s.repo.Upsert(ctx, *cred); err != nil {
		return "", err
	}

	return tr.AccessToken, nil
}

// SyncProductChannel empurra preço/estoque do produto (variations[0] — ver
// nota de simplificação em internal/product/model.go) pro anúncio já
// vinculado no Mercado Livre.
func (s *Service) SyncProductChannel(ctx context.Context, tenantID, productID string) error {
	p, err := s.products.Get(ctx, tenantID, productID)
	if err != nil {
		return err
	}

	ch, linked := p.Channels[MarketplaceMercadoLivre]
	if !linked || ch.ItemID == "" {
		return errors.New("marketplace: produto não está vinculado a um anúncio do Mercado Livre")
	}
	if len(p.Variations) == 0 {
		return errors.New("marketplace: produto sem variações, nada pra sincronizar")
	}

	accessToken, err := s.EnsureValidToken(ctx, tenantID)
	if err != nil {
		s.recordSyncError(ctx, tenantID, productID, ch, err)
		return err
	}

	primary := p.Variations[0]
	if err := s.ml.UpdateItemPriceStock(ctx, accessToken, ch.ItemID, primary.Price, primary.StockTotal); err != nil {
		s.recordSyncError(ctx, tenantID, productID, ch, err)
		return err
	}

	now := time.Now().UTC()
	_, err = s.products.UpdateChannel(ctx, tenantID, productID, MarketplaceMercadoLivre, product.Channel{
		ItemID:       ch.ItemID,
		SyncStatus:   product.SyncStatusSynced,
		LastSyncedAt: &now,
	})
	return err
}

func (s *Service) recordSyncError(ctx context.Context, tenantID, productID string, ch product.Channel, syncErr error) {
	_, _ = s.products.UpdateChannel(ctx, tenantID, productID, MarketplaceMercadoLivre, product.Channel{
		ItemID:     ch.ItemID,
		SyncStatus: product.SyncStatusError,
		LastError:  syncErr.Error(),
	})
}
