package dashboard

import (
	"context"
	"sort"
	"time"

	"github.com/gestorbuy/api/internal/orders"
	"github.com/gestorbuy/api/internal/product"
)

const marketplaceMercadoLivre = "mercadolivre"

type Service struct {
	orders   *orders.Repository
	products *product.Repository
}

func NewService(ordersRepo *orders.Repository, productsRepo *product.Repository) *Service {
	return &Service{orders: ordersRepo, products: productsRepo}
}

// Summary busca os pedidos do período e agrega. A parte de I/O (buscar
// pedidos, buscar custo de cada item_id no catálogo) fica só aqui — a
// agregação em si é a função pura Summarize, testável sem Mongo.
func (s *Service) Summary(ctx context.Context, tenantID string, from, to time.Time) (Summary, error) {
	orderList, err := s.orders.ListInRange(ctx, tenantID, from, to)
	if err != nil {
		return Summary{}, err
	}

	// Cache de custo por item_id — evita repetir a mesma busca no catálogo
	// pra pedidos diferentes do mesmo produto dentro do período. nil
	// significa "já buscou, não achou vínculo" (bem diferente de "ainda não
	// buscou").
	costCache := map[string]*float64{}

	costLookup := func(externalItemID string) (float64, bool) {
		if externalItemID == "" {
			return 0, false
		}
		if cached, ok := costCache[externalItemID]; ok {
			if cached == nil {
				return 0, false
			}
			return *cached, true
		}

		p, err := s.products.FindByChannelItemID(ctx, tenantID, marketplaceMercadoLivre, externalItemID)
		if err != nil || p == nil || len(p.Variations) == 0 {
			costCache[externalItemID] = nil
			return 0, false
		}
		cost := p.Variations[0].CostPrice
		costCache[externalItemID] = &cost
		return cost, true
	}

	return Summarize(from, to, orderList, costLookup), nil
}

// Summarize agrega uma lista de pedidos já buscada. Recebe costLookup como
// função em vez de mapa pronto pra deixar o chamador decidir a estratégia
// de cache/I/O — aqui dentro é só matemática.
func Summarize(from, to time.Time, orderList []orders.Order, costLookup func(externalItemID string) (float64, bool)) Summary {
	dailyMap := make(map[string]float64)
	var revenue, grossProfit, unmatchedRevenue float64

	for _, o := range orderList {
		revenue += o.TotalAmount

		day := o.DateCreated.Format("2006-01-02")
		dailyMap[day] += o.TotalAmount

		for _, item := range o.Items {
			itemRevenue := item.UnitPrice * float64(item.Quantity)
			cost, found := costLookup(item.ExternalItemID)
			if !found {
				unmatchedRevenue += itemRevenue
				continue
			}
			grossProfit += (item.UnitPrice - cost) * float64(item.Quantity)
		}
	}

	var marginPercent float64
	if revenue != 0 {
		marginPercent = grossProfit / revenue * 100
	}

	days := make([]string, 0, len(dailyMap))
	for d := range dailyMap {
		days = append(days, d)
	}
	sort.Strings(days)

	daily := make([]DailyRevenue, len(days))
	for i, d := range days {
		daily[i] = DailyRevenue{Date: d, Revenue: dailyMap[d]}
	}

	return Summary{
		From:             from,
		To:               to,
		Revenue:          revenue,
		GrossProfit:      grossProfit,
		MarginPercent:    marginPercent,
		UnmatchedRevenue: unmatchedRevenue,
		Daily:            daily,
	}
}
