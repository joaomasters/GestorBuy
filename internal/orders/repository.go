package orders

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository struct {
	col *mongo.Collection
}

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{col: db.Collection("orders")}
}

// Upsert grava o pedido — idempotente por (tenant_id, marketplace,
// external_order_id), então sincronizar de novo os mesmos pedidos (ex.:
// status mudou) só atualiza, nunca duplica.
func (r *Repository) Upsert(ctx context.Context, o Order) error {
	o.SyncedAt = time.Now().UTC()
	filter := bson.D{
		{Key: "tenant_id", Value: o.TenantID},
		{Key: "marketplace", Value: o.Marketplace},
		{Key: "external_order_id", Value: o.ExternalOrderID},
	}
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "status", Value: o.Status},
			{Key: "total_amount", Value: o.TotalAmount},
			{Key: "buyer_nickname", Value: o.BuyerNickname},
			{Key: "items", Value: o.Items},
			{Key: "date_created", Value: o.DateCreated},
			{Key: "synced_at", Value: o.SyncedAt},
		}},
		{Key: "$setOnInsert", Value: bson.D{
			{Key: "_id", Value: o.ID},
			{Key: "tenant_id", Value: o.TenantID},
			{Key: "marketplace", Value: o.Marketplace},
			{Key: "external_order_id", Value: o.ExternalOrderID},
		}},
	}

	if _, err := r.col.UpdateOne(ctx, filter, update, options.UpdateOne().SetUpsert(true)); err != nil {
		return fmt.Errorf("orders: upsert: %w", err)
	}
	return nil
}

func (r *Repository) List(ctx context.Context, tenantID string) ([]Order, error) {
	cursor, err := r.col.Find(ctx,
		bson.D{{Key: "tenant_id", Value: tenantID}},
		options.Find().SetSort(bson.D{{Key: "date_created", Value: -1}}),
	)
	if err != nil {
		return nil, fmt.Errorf("orders: list: %w", err)
	}
	defer cursor.Close(ctx)

	orders := make([]Order, 0)
	if err := cursor.All(ctx, &orders); err != nil {
		return nil, fmt.Errorf("orders: decode list: %w", err)
	}
	return orders, nil
}

// ListInRange devolve os pedidos com date_created em [from, to] — usado
// pelo internal/dashboard pra agregar faturamento/lucro do período.
func (r *Repository) ListInRange(ctx context.Context, tenantID string, from, to time.Time) ([]Order, error) {
	filter := bson.D{
		{Key: "tenant_id", Value: tenantID},
		{Key: "date_created", Value: bson.D{{Key: "$gte", Value: from}, {Key: "$lte", Value: to}}},
	}
	cursor, err := r.col.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "date_created", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("orders: list in range: %w", err)
	}
	defer cursor.Close(ctx)

	orders := make([]Order, 0)
	if err := cursor.All(ctx, &orders); err != nil {
		return nil, fmt.Errorf("orders: decode list in range: %w", err)
	}
	return orders, nil
}
