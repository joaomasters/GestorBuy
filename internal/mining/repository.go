package mining

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository struct {
	items     *mongo.Collection
	snapshots *mongo.Collection
}

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{
		items:     db.Collection("mined_items"),
		snapshots: db.Collection("mined_snapshots"),
	}
}

// UpsertItem cria ou atualiza os metadados do anúncio rastreado — chamado
// toda vez que um snapshot novo é tirado, então título/permalink nunca
// ficam desatualizados.
func (r *Repository) UpsertItem(ctx context.Context, item Item) (*Item, error) {
	now := time.Now().UTC()
	filter := bson.D{
		{Key: "tenant_id", Value: item.TenantID},
		{Key: "marketplace", Value: item.Marketplace},
		{Key: "external_item_id", Value: item.ExternalItemID},
	}
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "title", Value: item.Title},
			{Key: "permalink", Value: item.Permalink},
			{Key: "seller_id", Value: item.SellerID},
			{Key: "seller_nickname", Value: item.SellerNickname},
			{Key: "category_id", Value: item.CategoryID},
			{Key: "source", Value: item.Source},
			{Key: "updated_at", Value: now},
		}},
		{Key: "$setOnInsert", Value: bson.D{
			{Key: "_id", Value: item.ID},
			{Key: "tenant_id", Value: item.TenantID},
			{Key: "marketplace", Value: item.Marketplace},
			{Key: "external_item_id", Value: item.ExternalItemID},
			{Key: "created_at", Value: now},
		}},
	}

	_, err := r.items.UpdateOne(ctx, filter, update, options.UpdateOne().SetUpsert(true))
	if err != nil {
		return nil, fmt.Errorf("mining: upsert item: %w", err)
	}

	var saved Item
	if err := r.items.FindOne(ctx, filter).Decode(&saved); err != nil {
		return nil, fmt.Errorf("mining: reler item após upsert: %w", err)
	}
	return &saved, nil
}

func (r *Repository) FindItem(ctx context.Context, tenantID, id string) (*Item, error) {
	var item Item
	filter := bson.D{{Key: "_id", Value: id}, {Key: "tenant_id", Value: tenantID}}
	err := r.items.FindOne(ctx, filter).Decode(&item)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("mining: find item: %w", err)
	}
	return &item, nil
}

func (r *Repository) ListItems(ctx context.Context, tenantID string) ([]Item, error) {
	cursor, err := r.items.Find(ctx,
		bson.D{{Key: "tenant_id", Value: tenantID}},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}),
	)
	if err != nil {
		return nil, fmt.Errorf("mining: list items: %w", err)
	}
	defer cursor.Close(ctx)

	items := make([]Item, 0)
	if err := cursor.All(ctx, &items); err != nil {
		return nil, fmt.Errorf("mining: decode items: %w", err)
	}
	return items, nil
}

func (r *Repository) InsertSnapshot(ctx context.Context, snap Snapshot) error {
	if _, err := r.snapshots.InsertOne(ctx, snap); err != nil {
		return fmt.Errorf("mining: insert snapshot: %w", err)
	}
	return nil
}

// LatestSnapshot devolve o snapshot mais recente do item — usado pra
// mostrar preço/vendidos/faturamento atual na listagem sem precisar puxar
// o histórico inteiro.
func (r *Repository) LatestSnapshot(ctx context.Context, tenantID, marketplace, externalItemID string) (*Snapshot, error) {
	filter := bson.D{
		{Key: "metadata.tenant_id", Value: tenantID},
		{Key: "metadata.marketplace", Value: marketplace},
		{Key: "metadata.external_item_id", Value: externalItemID},
	}
	opts := options.FindOne().SetSort(bson.D{{Key: "ts", Value: -1}})

	var snap Snapshot
	err := r.snapshots.FindOne(ctx, filter, opts).Decode(&snap)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("mining: latest snapshot: %w", err)
	}
	return &snap, nil
}

func (r *Repository) History(ctx context.Context, tenantID, marketplace, externalItemID string) ([]Snapshot, error) {
	filter := bson.D{
		{Key: "metadata.tenant_id", Value: tenantID},
		{Key: "metadata.marketplace", Value: marketplace},
		{Key: "metadata.external_item_id", Value: externalItemID},
	}
	opts := options.Find().SetSort(bson.D{{Key: "ts", Value: 1}})

	cursor, err := r.snapshots.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("mining: history: %w", err)
	}
	defer cursor.Close(ctx)

	snapshots := make([]Snapshot, 0)
	if err := cursor.All(ctx, &snapshots); err != nil {
		return nil, fmt.Errorf("mining: decode history: %w", err)
	}
	return snapshots, nil
}
