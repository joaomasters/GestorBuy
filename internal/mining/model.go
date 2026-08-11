// Package mining implementa histórico de preço/vendas estilo Avant Pro —
// preço, quantidade vendida e faturamento estimado ao longo do tempo.
//
// Restrição descoberta em produção: a API do Mercado Livre bloqueia leitura
// de itens de terceiros mesmo pra apps autenticados (só libera acesso a
// anúncios da própria conta conectada) — e bloqueia chamada anônima de
// qualquer item (403 PolicyAgent). Por isso, diferente do desenho original,
// esse pacote usa o OAuth de internal/marketplace e só rastreia itens da
// conta conectada. Minerar concorrente de verdade (like o Avant Pro real)
// exigiria uma extensão de navegador rodando na sessão do usuário — fora de
// escopo aqui.

package mining

import "time"

const MarketplaceMercadoLivre = "mercadolivre"

// Item é o documento de identidade do anúncio rastreado — baixo volume de
// escrita, só muda quando título/permalink mudam. O histórico de
// preço/vendas de verdade fica em Snapshot (time series).
// Source diferencia como o item chegou até nós: "api" (TrackAndSnapshot,
// só funciona pra item da própria conta conectada — ver
// mercadolivre_client.go) ou "extension" (IngestSnapshot, dado já
// pré-minerado pela extensão de navegador, sem restrição de dono porque
// não é mais uma chamada de servidor).
const (
	SourceAPI       = "api"
	SourceExtension = "extension"
)

type Item struct {
	ID             string    `bson:"_id"`
	TenantID       string    `bson:"tenant_id"`
	Marketplace    string    `bson:"marketplace"`
	ExternalItemID string    `bson:"external_item_id"`
	Title          string    `bson:"title"`
	Permalink      string    `bson:"permalink,omitempty"`
	SellerID       int64     `bson:"seller_id,omitempty"`
	SellerNickname string    `bson:"seller_nickname,omitempty"`
	CategoryID     string    `bson:"category_id,omitempty"`
	Source         string    `bson:"source,omitempty"`
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
