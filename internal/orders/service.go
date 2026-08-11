package orders

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var ErrMarketplaceNotConnected = errors.New("orders: conecte sua conta do Mercado Livre em Integrações antes de sincronizar pedidos")

// MarketplaceAccount espelha a mesma interface local usada em
// internal/mining.Service — satisfeita por *marketplace.Service, sem
// import direto entre os pacotes.
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

// SyncOrders busca os pedidos atuais da conta conectada e grava/atualiza
// localmente. Devolve quantos pedidos foram processados.
func (s *Service) SyncOrders(ctx context.Context, tenantID string) (int, error) {
	accessToken, err := s.account.EnsureValidToken(ctx, tenantID)
	if err != nil {
		return 0, ErrMarketplaceNotConnected
	}
	sellerID, err := s.account.ConnectedExternalUserID(ctx, tenantID)
	if err != nil {
		return 0, ErrMarketplaceNotConnected
	}

	rawOrders, err := s.ml.SearchOrders(ctx, accessToken, sellerID)
	if err != nil {
		return 0, err
	}

	for _, ro := range rawOrders {
		items := make([]OrderItem, len(ro.OrderItems))
		for i, oi := range ro.OrderItems {
			items[i] = OrderItem{Title: oi.Item.Title, Quantity: oi.Quantity, UnitPrice: oi.UnitPrice}
		}

		order := Order{
			ID:              bson.NewObjectID().Hex(),
			TenantID:        tenantID,
			Marketplace:     MarketplaceMercadoLivre,
			ExternalOrderID: fmt.Sprintf("%d", ro.ID),
			Status:          ro.Status,
			TotalAmount:     ro.TotalAmount,
			BuyerNickname:   ro.Buyer.Nickname,
			Items:           items,
			DateCreated:     ro.DateCreated,
		}
		if err := s.repo.Upsert(ctx, order); err != nil {
			return 0, err
		}
	}

	return len(rawOrders), nil
}

func (s *Service) List(ctx context.Context, tenantID string) ([]Order, error) {
	return s.repo.List(ctx, tenantID)
}
