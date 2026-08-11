// Package marketplace implementa o Hub de integração com marketplaces —
// hoje só o Mercado Livre (internal/marketplace/mercadolivre), OAuth2 +
// sincronização de preço/estoque. Shopee, Amazon, Magalu, Shein entram
// depois seguindo o mesmo desenho: um sub-pacote por marketplace com o
// client REST, mais uma entrada na Credential.Marketplace.
package marketplace

import "time"

// MarketplaceMercadoLivre é usado tanto como valor de Credential.Marketplace
// quanto como chave em product.Product.Channels — precisa bater
// exatamente com o literal usado nas rotas HTTP (ver handler.go:
// "/integrations/mercadolivre/...", "/products/{id}/channels/mercadolivre/sync").
const MarketplaceMercadoLivre = "mercadolivre"

// Credential guarda os tokens OAuth de um tenant pra um marketplace. Os
// campos *Enc são sempre ciphertext (ver internal/platform/crypto) — nunca
// gravamos token em texto puro no banco.
type Credential struct {
	ID              string    `bson:"_id"`
	TenantID        string    `bson:"tenant_id"`
	Marketplace     string    `bson:"marketplace"`
	AccessTokenEnc  string    `bson:"access_token_enc"`
	RefreshTokenEnc string    `bson:"refresh_token_enc"`
	ExpiresAt       time.Time `bson:"expires_at"`
	ExternalUserID  string    `bson:"external_user_id,omitempty"`
	ConnectedAt     time.Time `bson:"connected_at"`
	UpdatedAt       time.Time `bson:"updated_at"`
}
