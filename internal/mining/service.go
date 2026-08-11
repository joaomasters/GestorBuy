package mining

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	ErrInvalidItemReference    = errors.New("mining: não foi possível identificar o ID do anúncio nesse link/texto")
	ErrMarketplaceNotConnected = errors.New("mining: conecte sua conta do Mercado Livre em Integrações antes de rastrear anúncios")

	// ErrNotOwnItem: o Mercado Livre nega leitura de itens de terceiros
	// mesmo pra apps autenticados — só libera acesso a anúncios da própria
	// conta conectada (ver comentário em mercadolivre_client.go). Rastrear
	// concorrente exigiria uma abordagem completamente diferente (extensão
	// de navegador rodando na sessão do usuário, não chamada de API
	// server-side) — fora de escopo por enquanto.
	ErrNotOwnItem = errors.New("mining: esse anúncio não pertence à conta do Mercado Livre conectada — a API do ML não libera leitura de itens de terceiros pra apps autenticados")
)

// itemIDPattern casa tanto o ID puro ("MLB1234567890") quanto o jeito que
// aparece nas URLs do Mercado Livre ("MLB-1234567890" com hífen).
var itemIDPattern = regexp.MustCompile(`(?i)MLB-?(\d+)`)

// ParseItemID normaliza qualquer coisa que o usuário colar (URL completa,
// ID com ou sem hífen) pro formato que a API do ML espera. Função pura,
// sem I/O — por isso é fácil de testar isoladamente.
func ParseItemID(urlOrID string) (string, error) {
	match := itemIDPattern.FindStringSubmatch(strings.TrimSpace(urlOrID))
	if match == nil {
		return "", ErrInvalidItemReference
	}
	return "MLB" + match[1], nil
}

// MarketplaceAccount dá acesso ao necessário de uma conta de marketplace
// conectada — token válido e o ID da conta no marketplace, pra confirmar
// que um item pertence a ela antes de tentar ler. Satisfeita por
// *internal/marketplace.Service (que já tem os dois métodos) — interface
// local em vez de importar o pacote concreto, pra não acoplar mining a
// marketplace além do necessário.
type MarketplaceAccount interface {
	EnsureValidToken(ctx context.Context, tenantID string) (string, error)
	ConnectedExternalUserID(ctx context.Context, tenantID string) (string, error)
}

type Service struct {
	repo    *Repository
	ml      *MercadoLivreClient
	account MarketplaceAccount
}

func NewService(repo *Repository, ml *MercadoLivreClient, account MarketplaceAccount) *Service {
	return &Service{repo: repo, ml: ml, account: account}
}

// TrackAndSnapshot resolve o ID (aceita link ou ID direto), busca o estado
// atual na API pública do ML, atualiza os metadados do item e grava um novo
// ponto no histórico (time series). Usada tanto pra rastrear um item novo
// quanto pra atualizar um que já existe — não distingue os dois casos, é
// sempre "olha de novo agora".
func (s *Service) TrackAndSnapshot(ctx context.Context, tenantID, urlOrID string) (*Item, *Snapshot, error) {
	itemID, err := ParseItemID(urlOrID)
	if err != nil {
		return nil, nil, err
	}

	accessToken, err := s.account.EnsureValidToken(ctx, tenantID)
	if err != nil {
		return nil, nil, ErrMarketplaceNotConnected
	}

	public, err := s.ml.FetchItem(ctx, itemID, accessToken)
	if err != nil {
		if errors.Is(err, ErrForbidden) {
			return nil, nil, ErrNotOwnItem
		}
		return nil, nil, err
	}

	// Defesa extra: a API do ML já barra item de terceiro com 403 (tratado
	// acima), mas confirmamos de novo aqui caso esse comportamento mude —
	// nunca gravamos snapshot de um item que não bateu com o dono da conta
	// conectada.
	connectedUserID, err := s.account.ConnectedExternalUserID(ctx, tenantID)
	if err != nil {
		return nil, nil, ErrMarketplaceNotConnected
	}
	if fmt.Sprintf("%d", public.SellerID) != connectedUserID {
		return nil, nil, ErrNotOwnItem
	}

	item, err := s.repo.UpsertItem(ctx, Item{
		ID:             bson.NewObjectID().Hex(),
		TenantID:       tenantID,
		Marketplace:    MarketplaceMercadoLivre,
		ExternalItemID: itemID,
		Title:          public.Title,
		Permalink:      public.Permalink,
		SellerID:       public.SellerID,
		CategoryID:     public.CategoryID,
	})
	if err != nil {
		return nil, nil, err
	}

	snap := Snapshot{
		TS: time.Now().UTC(),
		Metadata: SnapshotMeta{
			TenantID:       tenantID,
			Marketplace:    MarketplaceMercadoLivre,
			ExternalItemID: itemID,
		},
		Price:                 public.Price,
		SoldQuantity:          public.SoldQuantity,
		AvailableQuantity:     public.AvailableQuantity,
		EstimatedRevenueTotal: public.Price * float64(public.SoldQuantity),
	}
	if err := s.repo.InsertSnapshot(ctx, snap); err != nil {
		return nil, nil, err
	}

	return item, &snap, nil
}

// RefreshByID repete TrackAndSnapshot pro item já rastreado, sem o usuário
// precisar colar o link de novo.
func (s *Service) RefreshByID(ctx context.Context, tenantID, id string) (*Item, *Snapshot, error) {
	item, err := s.repo.FindItem(ctx, tenantID, id)
	if err != nil {
		return nil, nil, err
	}
	if item == nil {
		return nil, nil, fmt.Errorf("mining: item não encontrado")
	}
	return s.TrackAndSnapshot(ctx, tenantID, item.ExternalItemID)
}

type ItemWithSnapshot struct {
	Item     Item
	Snapshot *Snapshot
}

func (s *Service) List(ctx context.Context, tenantID string) ([]ItemWithSnapshot, error) {
	items, err := s.repo.ListItems(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	out := make([]ItemWithSnapshot, 0, len(items))
	for _, item := range items {
		snap, err := s.repo.LatestSnapshot(ctx, tenantID, item.Marketplace, item.ExternalItemID)
		if err != nil {
			return nil, err
		}
		out = append(out, ItemWithSnapshot{Item: item, Snapshot: snap})
	}
	return out, nil
}

func (s *Service) History(ctx context.Context, tenantID, id string) (*Item, []Snapshot, error) {
	item, err := s.repo.FindItem(ctx, tenantID, id)
	if err != nil {
		return nil, nil, err
	}
	if item == nil {
		return nil, nil, fmt.Errorf("mining: item não encontrado")
	}

	history, err := s.repo.History(ctx, tenantID, item.Marketplace, item.ExternalItemID)
	if err != nil {
		return nil, nil, err
	}
	return item, history, nil
}
