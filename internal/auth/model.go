// Package auth implementa registro/login multi-tenant e o middleware que
// resolve tenant_id a partir do JWT — a espinha dorsal do isolamento entre
// sellers descrito na seção 2.3 do documento de arquitetura: todo repository
// do resto do sistema recebe o tenant_id do contexto da requisição, nunca de
// um parâmetro solto vindo do cliente.
package auth

import "time"

type Role string

const RoleAdmin Role = "admin"

// User é o documento raiz da coleção `users`. PasswordHash nunca é
// serializado para fora do processo (sem tag json, só bson).
type User struct {
	ID           string    `bson:"_id"`
	TenantID     string    `bson:"tenant_id"`
	Email        string    `bson:"email"`
	PasswordHash string    `bson:"password_hash"`
	Role         Role      `bson:"role"`
	CreatedAt    time.Time `bson:"created_at"`
}
