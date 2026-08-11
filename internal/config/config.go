// Package config carrega a configuração da aplicação a partir de variáveis de
// ambiente. Em desenvolvimento, um arquivo .env (carregado pelo godotenv no
// main.go) preenche essas variáveis; em produção elas vêm do orquestrador
// (não commitamos segredos em arquivo nenhum).
package config

import (
	"fmt"
	"os"
	"time"
)

// Config agrega tudo que a aplicação precisa para subir.
type Config struct {
	HTTPPort    string
	MongoURI    string
	MongoDBName string
	JWTSecret   string
	JWTTTL      time.Duration
}

// Load lê as variáveis de ambiente e valida o mínimo necessário para o boot.
// Falha rápido (fail-fast) se algo crítico estiver faltando, em vez de deixar
// o servidor subir em um estado inconsistente.
func Load() (Config, error) {
	// PORT é a convenção usada por PaaS como Railway/Heroku, que a injeta
	// automaticamente no container — tem prioridade sobre HTTP_PORT, que
	// continua funcionando para desenvolvimento local via .env.
	cfg := Config{
		HTTPPort:    getEnv("PORT", getEnv("HTTP_PORT", "8080")),
		MongoURI:    getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDBName: getEnv("MONGO_DB_NAME", "gestorbuy"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		JWTTTL:      24 * time.Hour,
	}

	if cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("config: JWT_SECRET é obrigatório (defina no .env ou nas variáveis de ambiente)")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
