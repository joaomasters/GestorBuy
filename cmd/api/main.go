// Command api sobe o servidor HTTP do GESTORBUY. É o único ponto de
// composição (wiring) da aplicação: aqui e só aqui a config vira conexões
// reais e os módulos (tenant, auth, health, ...) são instanciados e ligados
// ao roteador.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/gestorbuy/api/internal/auth"
	"github.com/gestorbuy/api/internal/config"
	"github.com/gestorbuy/api/internal/health"
	"github.com/gestorbuy/api/internal/marketplace"
	"github.com/gestorbuy/api/internal/mining"
	"github.com/gestorbuy/api/internal/platform/httpserver"
	"github.com/gestorbuy/api/internal/platform/mongodb"
	"github.com/gestorbuy/api/internal/product"
	"github.com/gestorbuy/api/internal/tenant"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// .env é opcional (produção usa env vars do orquestrador) — erro de
	// arquivo ausente é silenciosamente ignorado.
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Error("configuração inválida", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	client, db, err := mongodb.Connect(ctx, cfg.MongoURI, cfg.MongoDBName)
	if err != nil {
		log.Error("falha ao conectar no MongoDB", "error", err)
		os.Exit(1)
	}
	defer func() { _ = client.Disconnect(context.Background()) }()

	if err := mongodb.EnsureIndexes(ctx, db); err != nil {
		log.Error("falha ao criar índices", "error", err)
		os.Exit(1)
	}

	// Wiring dos módulos: repository -> service -> handler, seguindo sempre
	// essa mesma ordem para os próximos módulos (products, orders, ...).
	tenantRepo := tenant.NewRepository(db)
	userRepo := auth.NewRepository(db)
	authSvc := auth.NewService(client, userRepo, tenantRepo, cfg.JWTSecret, cfg.JWTTTL)
	authHandler := auth.NewHandler(authSvc, log)
	healthHandler := health.NewHandler(client)
	productRepo := product.NewRepository(db)
	productSvc := product.NewService(productRepo)
	productHandler := product.NewHandler(productSvc, log)

	// Hub OAuth2 de marketplace — ML_CLIENT_ID/SECRET/REDIRECT_URI e
	// TOKEN_ENCRYPTION_KEY vazios são aceitos (app do Mercado Livre ainda
	// não criado); o serviço sobe normal e só recusa connect/sync com
	// ErrNotConfigured até essas variáveis existirem.
	marketplaceRepo := marketplace.NewRepository(db)
	marketplaceSvc, err := marketplace.New(marketplaceRepo, productSvc, cfg.JWTSecret, cfg.MLClientID, cfg.MLClientSecret, cfg.MLRedirectURI, cfg.TokenEncryptionKey)
	if err != nil {
		log.Error("configuração de marketplace inválida", "error", err)
		os.Exit(1)
	}
	marketplaceHandler := marketplace.NewHandler(marketplaceSvc, log, cfg.FrontendURL)

	miningRepo := mining.NewRepository(db)
	miningSvc := mining.NewService(miningRepo, mining.NewMercadoLivreClient())
	miningHandler := mining.NewHandler(miningSvc, log)

	mux := http.NewServeMux()
	healthHandler.Routes(mux)
	authHandler.Routes(mux)
	mux.Handle("GET /me", authSvc.Middleware(http.HandlerFunc(authHandler.Me)))
	productHandler.Routes(mux, authSvc.Middleware)
	marketplaceHandler.Routes(mux, authSvc.Middleware)
	marketplaceHandler.RoutesPublic(mux)
	miningHandler.Routes(mux, authSvc.Middleware)

	srv := httpserver.New(":"+cfg.HTTPPort, mux, log)

	go func() {
		log.Info("servidor HTTP no ar", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("servidor HTTP encerrou com erro", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	log.Info("encerrando servidor (sinal recebido)...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown não gracioso", "error", err)
	}
}
