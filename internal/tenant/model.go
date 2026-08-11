// Package tenant modela a entidade Tenant (o "seller" que assina o SaaS) —
// a raiz de isolamento de todo o multi-tenancy descrito na seção 2.3 do
// documento de arquitetura.
package tenant

import "time"

// PlanTier reflete o tier de assinatura, que no futuro decide inclusive se o
// tenant vive no cluster compartilhado ou em um cluster dedicado (silo).
type PlanTier string

const (
	PlanTierStandard   PlanTier = "standard"
	PlanTierPro        PlanTier = "pro"
	PlanTierEnterprise PlanTier = "enterprise"
)

// Tenant é o documento raiz da coleção `tenants`.
type Tenant struct {
	ID        string    `bson:"_id"`
	Name      string    `bson:"name"`
	Slug      string    `bson:"slug"`
	PlanTier  PlanTier  `bson:"plan_tier"`
	CreatedAt time.Time `bson:"created_at"`
}
