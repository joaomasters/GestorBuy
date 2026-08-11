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

	return nil
}
