package marketplace

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
	col *mongo.Collection
}

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{col: db.Collection("marketplace_credentials")}
}

// Upsert cria ou substitui a credencial do tenant pra esse marketplace —
// reconectar (ex.: depois de revogar acesso no próprio ML) simplesmente
// sobrescreve a anterior.
func (r *Repository) Upsert(ctx context.Context, c Credential) error {
	c.UpdatedAt = time.Now().UTC()
	filter := bson.D{{Key: "tenant_id", Value: c.TenantID}, {Key: "marketplace", Value: c.Marketplace}}

	// connected_at não pode estar em $set e $setOnInsert ao mesmo tempo —
	// só entra em $setOnInsert, pra não ser sobrescrito numa reconexão.
	_, err := r.col.UpdateOne(ctx, filter, bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "_id", Value: c.ID},
			{Key: "tenant_id", Value: c.TenantID},
			{Key: "marketplace", Value: c.Marketplace},
			{Key: "access_token_enc", Value: c.AccessTokenEnc},
			{Key: "refresh_token_enc", Value: c.RefreshTokenEnc},
			{Key: "expires_at", Value: c.ExpiresAt},
			{Key: "external_user_id", Value: c.ExternalUserID},
			{Key: "updated_at", Value: c.UpdatedAt},
		}},
		{Key: "$setOnInsert", Value: bson.D{{Key: "connected_at", Value: time.Now().UTC()}}},
	}, options.UpdateOne().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("marketplace: upsert credential: %w", err)
	}
	return nil
}

func (r *Repository) FindByTenant(ctx context.Context, tenantID, mktplace string) (*Credential, error) {
	var c Credential
	filter := bson.D{{Key: "tenant_id", Value: tenantID}, {Key: "marketplace", Value: mktplace}}
	err := r.col.FindOne(ctx, filter).Decode(&c)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("marketplace: find credential: %w", err)
	}
	return &c, nil
}

func (r *Repository) Delete(ctx context.Context, tenantID, mktplace string) (bool, error) {
	filter := bson.D{{Key: "tenant_id", Value: tenantID}, {Key: "marketplace", Value: mktplace}}
	result, err := r.col.DeleteOne(ctx, filter)
	if err != nil {
		return false, fmt.Errorf("marketplace: delete credential: %w", err)
	}
	return result.DeletedCount > 0, nil
}
