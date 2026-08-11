// Package product implementa o catálogo unificado de anúncios (seção 2.1 do
// documento de arquitetura). Nesta primeira entrega guarda só os dados
// "verdade" do seller — SKU, título, variações — sem o sub-documento
// `channels` (preço/estoque por marketplace), que só faz sentido a partir do
// Hub OAuth2. Adicionar esse campo depois é seguro: MongoDB não exige
// migração pra um campo novo num documento existente.
package product

import "time"

// StockStrategy reflete se o estoque é um pool único compartilhado entre
// canais ou alocado por canal — decisão que só importa de verdade quando
// existir mais de um canal de venda sincronizado (Hub OAuth2). Por enquanto
// todo produto nasce com o padrão.
type StockStrategy string

const StockStrategyShared StockStrategy = "shared_pool"

// Variation é uma combinação vendável (ex.: cor+tamanho) dentro do produto
// mestre. Embutida no documento do produto — sempre lida/escrita junto com
// ele, não há caso de uso pra buscar uma variação isolada do produto pai.
type Variation struct {
	VariationSKU string            `bson:"variation_sku"`
	Attributes   map[string]string `bson:"attributes,omitempty"`
	StockTotal   int               `bson:"stock_total"`
	CostPrice    float64           `bson:"cost_price"`
	Price        float64           `bson:"price"`
}

// Channel guarda o estado de sincronização do produto num marketplace
// específico. Chega agora que existe o primeiro consumidor de verdade
// (internal/marketplace, Hub OAuth2 do Mercado Livre) — até então ficava de
// fora de propósito, porque um campo especulativo sem consumidor é só
// complexidade grátis.
//
// Simplificação assumida nesta entrega: o preço/estoque sincronizado é
// sempre o de `Variations[0]` — mapeamento de canal por variação individual
// fica para quando o catálogo precisar de verdade disso.
type Channel struct {
	ItemID       string     `bson:"item_id,omitempty"`
	SyncStatus   string     `bson:"sync_status,omitempty"` // not_linked | linked | synced | error
	LastSyncedAt *time.Time `bson:"last_synced_at,omitempty"`
	LastError    string     `bson:"last_error,omitempty"`
}

const (
	SyncStatusLinked = "linked"
	SyncStatusSynced = "synced"
	SyncStatusError  = "error"
)

// Product é o documento raiz da coleção `products`.
type Product struct {
	ID                 string             `bson:"_id"`
	TenantID           string             `bson:"tenant_id"`
	SKUMaster          string             `bson:"sku_master"`
	Title              string             `bson:"title"`
	Brand              string             `bson:"brand,omitempty"`
	CategoryNormalized string             `bson:"category_normalized,omitempty"`
	Variations         []Variation        `bson:"variations"`
	StockStrategy      StockStrategy      `bson:"stock_strategy"`
	Channels           map[string]Channel `bson:"channels,omitempty"`
	CreatedAt          time.Time          `bson:"created_at"`
	UpdatedAt          time.Time          `bson:"updated_at"`
}
