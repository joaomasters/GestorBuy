package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"

	"github.com/gestorbuy/api/internal/tenant"
)

var ErrInvalidCredentials = errors.New("auth: credenciais inválidas")

// Claims são os dados embutidos no JWT. TenantID é o campo que sustenta todo
// o isolamento multi-tenant descrito na seção 2.3 do doc de arquitetura:
// todo handler autenticado lê o tenant_id daqui, nunca de um parâmetro de
// URL/body vindo do cliente.
type Claims struct {
	TenantID string `json:"tenant_id"`
	UserID   string `json:"user_id"`
	Role     Role   `json:"role"`
	jwt.RegisteredClaims
}

type Service struct {
	client  *mongo.Client
	users   *Repository
	tenants *tenant.Repository
	jwtKey  []byte
	jwtTTL  time.Duration
}

func NewService(client *mongo.Client, users *Repository, tenants *tenant.Repository, jwtKey string, jwtTTL time.Duration) *Service {
	return &Service{client: client, users: users, tenants: tenants, jwtKey: []byte(jwtKey), jwtTTL: jwtTTL}
}

// RegisterInput agrupa os dados para criar um tenant novo com seu primeiro
// usuário administrador.
type RegisterInput struct {
	TenantName string
	TenantSlug string
	Email      string
	Password   string
}

// Register cria o tenant e o usuário admin numa única transação ACID
// (Sessions/Transactions — seção 3.2 do doc de arquitetura): ou os dois
// documentos existem, ou nenhum existe. Evita o estado inconsistente de um
// tenant "órfão" sem nenhum usuário capaz de acessá-lo.
func (s *Service) Register(ctx context.Context, in RegisterInput) (*RegisterResult, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("auth: hash de senha: %w", err)
	}

	tenantID := bson.NewObjectID().Hex()
	userID := bson.NewObjectID().Hex()
	now := time.Now().UTC()
	newUser := User{
		ID:           userID,
		TenantID:     tenantID,
		Email:        in.Email,
		PasswordHash: string(hash),
		Role:         RoleAdmin,
		CreatedAt:    now,
	}

	session, err := s.client.StartSession()
	if err != nil {
		return nil, fmt.Errorf("auth: start session: %w", err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sCtx context.Context) (any, error) {
		t := tenant.Tenant{
			ID:        tenantID,
			Name:      in.TenantName,
			Slug:      in.TenantSlug,
			PlanTier:  tenant.PlanTierStandard,
			CreatedAt: now,
		}
		if err := s.tenants.Create(sCtx, t); err != nil {
			return nil, err
		}
		if err := s.users.Create(sCtx, newUser); err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	token, err := s.issueToken(newUser)
	if err != nil {
		return nil, fmt.Errorf("auth: registro concluído, mas falhou ao emitir token: %w", err)
	}

	return &RegisterResult{TenantID: tenantID, UserID: userID, Token: token}, nil
}

// RegisterResult é o resultado mínimo de Register — evita expor os documentos
// internos completos (ex.: password_hash) para a camada HTTP.
type RegisterResult struct {
	TenantID string
	UserID   string
	Token    string
}

// Login valida e-mail+senha e emite um JWT contendo tenant_id, user_id e
// role.
func (s *Service) Login(ctx context.Context, email, password string) (string, error) {
	u, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if u == nil {
		return "", ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	return s.issueToken(*u)
}

func (s *Service) issueToken(u User) (string, error) {
	now := time.Now()
	claims := Claims{
		TenantID: u.TenantID,
		UserID:   u.ID,
		Role:     u.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.jwtTTL)),
			Subject:   u.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.jwtKey)
	if err != nil {
		return "", fmt.Errorf("auth: assinar token: %w", err)
	}
	return signed, nil
}

// ParseToken valida a assinatura/expiração e devolve as claims — usado pelo
// AuthMiddleware.
func (s *Service) ParseToken(raw string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("auth: método de assinatura inesperado: %v", t.Header["alg"])
		}
		return s.jwtKey, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidCredentials
	}
	return claims, nil
}
