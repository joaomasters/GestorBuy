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

var ErrInvalidItemReference = errors.New("mining: não foi possível identificar o ID do anúncio nesse link/texto")

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

type Service struct {
	repo *Repository
	ml   *MercadoLivreClient
}

func NewService(repo *Repository, ml *MercadoLivreClient) *Service {
	return &Service{repo: repo, ml: ml}
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

	public, err := s.ml.FetchPublicItem(ctx, itemID)
	if err != nil {
		return nil, nil, err
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
