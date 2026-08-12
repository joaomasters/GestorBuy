package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// EnsureIndexes cria (de forma idempotente — CreateOne é no-op se o índice já
// existe com as mesmas opções) os índices essenciais descritos na seção 2.4
// do documento de arquitetura. Roda no boot da aplicação; à medida que novos
// módulos (products, orders, mined_products...) forem implementados, seus
// índices entram aqui seguindo o mesmo padrão: tenant_id sempre como prefixo
// de índice composto em coleção multi-tenant.
func EnsureIndexes(ctx context.Context, db *mongo.Database) error {
	// Índice composto com tenant_id como prefixo: toda listagem/busca de
	// usuários de um tenant usa esse índice (padrão da seção 2.4).
	usersTenantIdx := mongo.IndexModel{
		Keys:    bson.D{{Key: "tenant_id", Value: 1}, {Key: "email", Value: 1}},
		Options: options.Index().SetName("tenant_id_email"),
	}
	// Unicidade de e-mail é GLOBAL (não por tenant): no MVP um usuário
	// pertence a exatamente um tenant e faz login só com e-mail+senha, sem
	// informar o tenant antecipadamente — login por e-mail ambíguo entre
	// tenants ficaria inseguro/incorreto. Suporte a um usuário em múltiplos
	// tenants (ex.: agência) é item de Fase 2, não deste MVP.
	usersEmailUniqueIdx := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("email_unique"),
	}
	if _, err := db.Collection("users").Indexes().CreateMany(ctx, []mongo.IndexModel{usersTenantIdx, usersEmailUniqueIdx}); err != nil {
		return fmt.Errorf("mongodb: índice de users: %w", err)
	}

	tenantsIdx := mongo.IndexModel{
		Keys:    bson.D{{Key: "slug", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("slug_unique"),
	}
	if _, err := db.Collection("tenants").Indexes().CreateOne(ctx, tenantsIdx); err != nil {
		return fmt.Errorf("mongodb: índice de tenants: %w", err)
	}

	if err := ensureProductIndexes(ctx, db); err != nil {
		return err
	}

	if err := ensureMarketplaceIndexes(ctx, db); err != nil {
		return err
	}

	if err := ensureMiningIndexesAndCollections(ctx, db); err != nil {
		return err
	}

	if err := ensureOrdersIndexes(ctx, db); err != nil {
		return err
	}

	return nil
}

func ensureOrdersIndexes(ctx context.Context, db *mongo.Database) error {
	uniqueIdx := mongo.IndexModel{
		Keys: bson.D{
			{Key: "tenant_id", Value: 1},
			{Key: "marketplace", Value: 1},
			{Key: "external_order_id", Value: 1},
		},
		Options: options.Index().SetUnique(true).SetName("tenant_id_marketplace_order_unique"),
	}
	// Suporta o range query de internal/dashboard (ListInRange) sem scan
	// completo da coleção por tenant.
	dateRangeIdx := mongo.IndexModel{
		Keys:    bson.D{{Key: "tenant_id", Value: 1}, {Key: "date_created", Value: -1}},
		Options: options.Index().SetName("tenant_id_date_created"),
	}
	if _, err := db.Collection("orders").Indexes().CreateMany(ctx, []mongo.IndexModel{uniqueIdx, dateRangeIdx}); err != nil {
		return fmt.Errorf("mongodb: índice de orders: %w", err)
	}
	return nil
}

func ensureMarketplaceIndexes(ctx context.Context, db *mongo.Database) error {
	credentialIdx := mongo.IndexModel{
		Keys:    bson.D{{Key: "tenant_id", Value: 1}, {Key: "marketplace", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("tenant_id_marketplace_unique"),
	}
	if _, err := db.Collection("marketplace_credentials").Indexes().CreateOne(ctx, credentialIdx); err != nil {
		return fmt.Errorf("mongodb: índice de marketplace_credentials: %w", err)
	}
	return nil
}

func ensureMiningIndexesAndCollections(ctx context.Context, db *mongo.Database) error {
	itemsIdx := mongo.IndexModel{
		Keys: bson.D{
			{Key: "tenant_id", Value: 1},
			{Key: "marketplace", Value: 1},
			{Key: "external_item_id", Value: 1},
		},
		Options: options.Index().SetUnique(true).SetName("tenant_id_marketplace_item_unique"),
	}
	if _, err := db.Collection("mined_items").Indexes().CreateOne(ctx, itemsIdx); err != nil {
		return fmt.Errorf("mongodb: índice de mined_items: %w", err)
	}

	if err := ensureTimeSeriesCollection(ctx, db, "mined_snapshots"); err != nil {
		return err
	}

	return nil
}

// ensureTimeSeriesCollection cria a coleção como Time Series (seção 1.3 do
// doc de arquitetura) se ela ainda não existir. CreateCollection não é
// idempotente feito CreateIndex — tentar criar de novo retorna erro
// NamespaceExists — então checamos a lista de coleções antes.
func ensureTimeSeriesCollection(ctx context.Context, db *mongo.Database, name string) error {
	names, err := db.ListCollectionNames(ctx, bson.D{{Key: "name", Value: name}})
	if err != nil {
		return fmt.Errorf("mongodb: listar coleções pra checar %s: %w", name, err)
	}
	if len(names) > 0 {
		return nil // já existe, nada a fazer
	}

	tsOpts := options.TimeSeries().
		SetTimeField("ts").
		SetMetaField("metadata").
		SetGranularity("hours")

	if err := db.CreateCollection(ctx, name, options.CreateCollection().SetTimeSeriesOptions(tsOpts)); err != nil {
		return fmt.Errorf("mongodb: criar time series collection %s: %w", name, err)
	}
	return nil
}

func ensureProductIndexes(ctx context.Context, db *mongo.Database) error {
	skuUniqueIdx := mongo.IndexModel{
		Keys:    bson.D{{Key: "tenant_id", Value: 1}, {Key: "sku_master", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("tenant_id_sku_master_unique"),
	}
	variationSKUIdx := mongo.IndexModel{
		Keys:    bson.D{{Key: "tenant_id", Value: 1}, {Key: "variations.variation_sku", Value: 1}},
		Options: options.Index().SetName("tenant_id_variation_sku"),
	}
	if _, err := db.Collection("products").Indexes().CreateMany(ctx, []mongo.IndexModel{skuUniqueIdx, variationSKUIdx}); err != nil {
		return fmt.Errorf("mongodb: índices de products: %w", err)
	}

	// Text index para busca de anúncios (seção 2.4 do doc de arquitetura).
	textIdx := mongo.IndexModel{
		Keys: bson.D{{Key: "title", Value: "text"}, {Key: "brand", Value: "text"}, {Key: "sku_master", Value: "text"}},
		Options: options.Index().
			SetName("products_text_search").
			SetWeights(bson.D{{Key: "title", Value: 10}, {Key: "brand", Value: 5}, {Key: "sku_master", Value: 3}}).
			SetDefaultLanguage("portuguese"),
	}
	if _, err := db.Collection("products").Indexes().CreateOne(ctx, textIdx); err != nil {
		return fmt.Errorf("mongodb: índice de texto de products: %w", err)
	}

	return nil
}
