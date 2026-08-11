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

	// FrontendURL é pra onde o backend redireciona o navegador depois do
	// callback OAuth de um marketplace (ver internal/marketplace).
	FrontendURL string

	// TokenEncryptionKey cifra credenciais de marketplace antes de gravar
	// no Mongo (ver internal/platform/crypto). Opcional no boot: só é
	// exigida no momento em que alguém tenta conectar um marketplace —
	// não faz sentido travar a aplicação inteira por causa de uma
	// integração que ainda não existe.
	TokenEncryptionKey string

	// Credenciais do app OAuth do Mercado Livre. Também opcionais no boot
	// pelo mesmo motivo — ver internal/marketplace/mercadolivre.
	MLClientID     string
	MLClientSecret string
	MLRedirectURI  string
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

		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:3000"),

		TokenEncryptionKey: os.Getenv("TOKEN_ENCRYPTION_KEY"),

		MLClientID:     os.Getenv("ML_CLIENT_ID"),
		MLClientSecret: os.Getenv("ML_CLIENT_SECRET"),
		MLRedirectURI:  os.Getenv("ML_REDIRECT_URI"),
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
