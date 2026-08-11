package tenant

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// ErrSlugTaken é retornado quando o slug do tenant já existe (índice único
// de `slug` — ver internal/platform/mongodb/indexes.go).
var ErrSlugTaken = errors.New("tenant: slug já está em uso")

// Repository encapsula o acesso à coleção `tenants`. É o único lugar do
// código que sabe o nome da coleção e a forma do documento no Mongo.
type Repository struct {
	col *mongo.Collection
}

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{col: db.Collection("tenants")}
}

// Create insere um novo tenant dentro da sessão/transação fornecida pelo
// chamador (auth.Service.Register cria tenant+usuário atomicamente — seção
// 3.2 do doc de arquitetura). ctx deve carregar a sessão via
// mongo.WithSession quando chamado dentro de uma transação.
func (r *Repository) Create(ctx context.Context, t Tenant) error {
	_, err := r.col.InsertOne(ctx, t)
	if mongo.IsDuplicateKeyError(err) {
		return ErrSlugTaken
	}
	if err != nil {
		return fmt.Errorf("tenant: create: %w", err)
	}
	return nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (*Tenant, error) {
	var t Tenant
	err := r.col.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&t)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("tenant: find by id: %w", err)
	}
	return &t, nil
}
