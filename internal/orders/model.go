// Package orders implementa a gestão de pedidos (Gestor Seller) — puxa os
// pedidos da própria conta conectada no Mercado Livre. Sincronização é sob
// demanda ("sincronizar agora"), não um worker/webhook em background nesta
// entrega — mesmo padrão pull-based já usado em internal/mining.
package orders

import "time"

const MarketplaceMercadoLivre = "mercadolivre"

type OrderItem struct {
	// ExternalItemID é o item_id do Mercado Livre — usado por
	// internal/dashboard pra achar o produto do catálogo vinculado (via
	// product.Channels) e saber o custo, calculando lucro bruto de verdade
	// em vez de só somar receita. Pode vir vazio pra pedidos sincronizados
	// antes desse campo existir — trate como "sem vínculo conhecido".
	ExternalItemID string  `bson:"external_item_id,omitempty"`
	Title          string  `bson:"title"`
	Quantity       int     `bson:"quantity"`
	UnitPrice      float64 `bson:"unit_price"`
}

type Order struct {
	ID              string      `bson:"_id"`
	TenantID        string      `bson:"tenant_id"`
	Marketplace     string      `bson:"marketplace"`
	ExternalOrderID string      `bson:"external_order_id"`
	Status          string      `bson:"status"`
	TotalAmount     float64     `bson:"total_amount"`
	BuyerNickname   string      `bson:"buyer_nickname,omitempty"`
	Items           []OrderItem `bson:"items"`
	DateCreated     time.Time   `bson:"date_created"`
	SyncedAt        time.Time   `bson:"synced_at"`
}
