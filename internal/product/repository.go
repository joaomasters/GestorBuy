package product

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var ErrSKUTaken = errors.New("product: sku_master já está em uso neste tenant")

type Repository struct {
	col *mongo.Collection
}

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{col: db.Collection("products")}
}

func (r *Repository) Create(ctx context.Context, p Product) error {
	_, err := r.col.InsertOne(ctx, p)
	if mongo.IsDuplicateKeyError(err) {
		return ErrSKUTaken
	}
	if err != nil {
		return fmt.Errorf("product: create: %w", err)
	}
	return nil
}

// FindByID busca sempre escopado por tenant_id — mesmo padrão de isolamento
// usado em todo o resto do sistema (ver internal/tenant/repository.go).
// FindByChannelItemID acha o produto do catálogo vinculado a um item_id de
// marketplace (via product.Channels) — usado por internal/dashboard pra
// achar o custo de um item de pedido e calcular lucro bruto de verdade.
func (r *Repository) FindByChannelItemID(ctx context.Context, tenantID, marketplace, itemID string) (*Product, error) {
	var p Product
	filter := bson.D{
		{Key: "tenant_id", Value: tenantID},
		{Key: "channels." + marketplace + ".item_id", Value: itemID},
	}
	err := r.col.FindOne(ctx, filter).Decode(&p)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("product: find by channel item id: %w", err)
	}
	return &p, nil
}

func (r *Repository) FindByID(ctx context.Context, tenantID, id string) (*Product, error) {
	var p Product
	filter := bson.D{{Key: "_id", Value: id}, {Key: "tenant_id", Value: tenantID}}
	err := r.col.FindOne(ctx, filter).Decode(&p)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("product: find by id: %w", err)
	}
	return &p, nil
}

// ListOptions controla paginação e busca textual. Limit é sempre aplicado
// (com teto), nunca uma listagem sem limite.
type ListOptions struct {
	Limit  int64
	Offset int64
	Query  string // usa o text index (título/marca/sku_master) quando preenchido
}

func (r *Repository) List(ctx context.Context, tenantID string, opts ListOptions) ([]Product, int64, error) {
	filter := bson.D{{Key: "tenant_id", Value: tenantID}}
	if opts.Query != "" {
		filter = append(filter, bson.E{Key: "$text", Value: bson.D{{Key: "$search", Value: opts.Query}}})
	}

	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("product: count: %w", err)
	}

	findOpts := options.Find().
		SetLimit(opts.Limit).
		SetSkip(opts.Offset).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.col.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, 0, fmt.Errorf("product: list: %w", err)
	}
	defer cursor.Close(ctx)

	products := make([]Product, 0, opts.Limit)
	if err := cursor.All(ctx, &products); err != nil {
		return nil, 0, fmt.Errorf("product: decode list: %w", err)
	}

	return products, total, nil
}

// Update aplica um $set parcial, sempre escopado por tenant_id. O service é
// quem monta `set` a partir da validação — o repository não decide regra de
// negócio, só persiste.
func (r *Repository) Update(ctx context.Context, tenantID, id string, set bson.M) (*Product, error) {
	set["updated_at"] = time.Now().UTC()
	filter := bson.D{{Key: "_id", Value: id}, {Key: "tenant_id", Value: tenantID}}
	update := bson.D{{Key: "$set", Value: set}}

	result := r.col.FindOneAndUpdate(ctx, filter, update, options.FindOneAndUpdate().SetReturnDocument(options.After))
	var p Product
	err := result.Decode(&p)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if mongo.IsDuplicateKeyError(err) {
		return nil, ErrSKUTaken
	}
	if err != nil {
		return nil, fmt.Errorf("product: update: %w", err)
	}
	return &p, nil
}

func (r *Repository) Delete(ctx context.Context, tenantID, id string) (bool, error) {
	filter := bson.D{{Key: "_id", Value: id}, {Key: "tenant_id", Value: tenantID}}
	result, err := r.col.DeleteOne(ctx, filter)
	if err != nil {
		return false, fmt.Errorf("product: delete: %w", err)
	}
	return result.DeletedCount > 0, nil
}
