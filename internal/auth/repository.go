package auth

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var ErrEmailTaken = errors.New("auth: e-mail já está em uso neste tenant")

type Repository struct {
	col *mongo.Collection
}

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{col: db.Collection("users")}
}

// Create insere o usuário. Espera-se que ctx carregue uma sessão de
// transação quando chamado junto com tenant.Repository.Create (ver
// Service.Register), garantindo que tenant e primeiro usuário admin nasçam
// atomicamente.
func (r *Repository) Create(ctx context.Context, u User) error {
	_, err := r.col.InsertOne(ctx, u)
	if mongo.IsDuplicateKeyError(err) {
		return ErrEmailTaken
	}
	if err != nil {
		return fmt.Errorf("auth: create user: %w", err)
	}
	return nil
}

// FindByEmail busca o usuário só por e-mail (globalmente único — ver
// internal/platform/mongodb/indexes.go). É o caminho usado no login, onde o
// cliente ainda não informa o tenant_id; o tenant_id do documento encontrado
// passa a ser a fonte de verdade para o restante da sessão autenticada.
func (r *Repository) FindByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	err := r.col.FindOne(ctx, bson.D{{Key: "email", Value: email}}).Decode(&u)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("auth: find by email: %w", err)
	}
	return &u, nil
}
