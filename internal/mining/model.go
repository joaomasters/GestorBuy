// Package mining implementa a mineração de anúncios estilo Avant Pro —
// preço, quantidade vendida e faturamento estimado de qualquer anúncio
// público do Mercado Livre (concorrente ou não). Diferente de
// internal/marketplace, aqui não tem OAuth nenhum: a API pública de itens
// do ML (`GET /items/{id}`) já expõe tudo que precisamos sem autenticação.
package mining

import "time"

const MarketplaceMercadoLivre = "mercadolivre"

// Item é o documento de identidade do anúncio rastreado — baixo volume de
// escrita, só muda quando título/permalink mudam. O histórico de
// preço/vendas de verdade fica em Snapshot (time series).
type Item struct {
	ID             string    `bson:"_id"`
	TenantID       string    `bson:"tenant_id"`
	Marketplace    string    `bson:"marketplace"`
	ExternalItemID string    `bson:"external_item_id"`
	Title          string    `bson:"title"`
	Permalink      string    `bson:"permalink,omitempty"`
	SellerID       int64     `bson:"seller_id,omitempty"`
	CategoryID     string    `bson:"category_id,omitempty"`
	CreatedAt      time.Time `bson:"created_at"`
	UpdatedAt      time.Time `bson:"updated_at"`
}

// Snapshot é um documento da time series collection `mined_snapshots` —
// mesmo padrão descrito na seção 1.3/2.2 do doc de arquitetura.
type Snapshot struct {
	TS                    time.Time    `bson:"ts"`
	Metadata              SnapshotMeta `bson:"metadata"`
	Price                 float64      `bson:"price"`
	SoldQuantity          int64        `bson:"sold_quantity"`
	AvailableQuantity     int64        `bson:"available_quantity"`
	EstimatedRevenueTotal float64      `bson:"estimated_revenue_total"`
}

type SnapshotMeta struct {
	TenantID       string `bson:"tenant_id"`
	Marketplace    string `bson:"marketplace"`
	ExternalItemID string `bson:"external_item_id"`
}
