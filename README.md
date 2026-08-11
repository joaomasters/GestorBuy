# GESTORBUY — API

Backend do GESTORBUY: SaaS B2B de gestão centralizada de múltiplos marketplaces
+ inteligência de vendas. Contexto completo de arquitetura em
[docs/arquitetura-saas-marketplaces.md](docs/arquitetura-saas-marketplaces.md).

Este repositório contém a **Fase 0**: fundação multi-tenant (MongoDB + auth
JWT) sobre a qual o Hub de OAuth2 dos marketplaces e os módulos Gestor/Analytics
serão construídos nas próximas fases.

## Stack

- **Go 1.26+**
- **MongoDB** (driver oficial `mongo-driver/v2`) — local via Docker em dev, Atlas em produção
- Sem framework HTTP externo — `net/http` (stdlib, roteamento por método+padrão desde a 1.22)

## Rodando localmente

```bash
cp .env.example .env
# ajuste JWT_SECRET no .env antes de seguir

docker compose up -d mongodb   # sobe o MongoDB local
go run ./cmd/api               # sobe a API em :8080
```

> Se tiver `make` instalado, `make dev` faz as duas coisas de uma vez.

### Testar os endpoints

```bash
# 1. Cria um tenant + usuário admin, retorna JWT
curl -s -X POST localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"tenant_name":"Loja Exemplo","tenant_slug":"loja-exemplo","email":"admin@lojaexemplo.com","password":"senha-forte-123"}'

# 2. Login (mesmas credenciais)
curl -s -X POST localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@lojaexemplo.com","password":"senha-forte-123"}'

# 3. Rota protegida — prova que o middleware resolve tenant_id a partir do JWT
curl -s localhost:8080/me -H "Authorization: Bearer <token retornado acima>"

# 4. Health check
curl -s localhost:8080/health
```

## Testes e verificação

```bash
go vet ./...
go build ./...
go test ./...
```

CI (`.github/workflows/ci.yml`) roda os três em todo push/PR para `main`.

## Estrutura

```
cmd/api/            entrypoint — único lugar que faz o wiring dos módulos
internal/config/     leitura de env vars
internal/platform/   infraestrutura compartilhada (mongodb, httpserver)
internal/tenant/     entidade raiz de multi-tenancy
internal/auth/       registro/login, JWT, middleware de autenticação
internal/health/     healthcheck
```

Todo módulo de domínio novo (products, orders, mined_products...) segue o
mesmo padrão: `model.go` → `repository.go` → `service.go` → `handler.go`, e
toda coleção multi-tenant usa `tenant_id` como primeiro campo de índice
composto e como filtro obrigatório em toda query — nunca confiar em
`tenant_id` vindo de parâmetro de URL/body, sempre ler de
`auth.ClaimsFromContext`.

## Deploy

Produção roda em **Railway** (container Docker, ver [Dockerfile](Dockerfile)) conectado a um cluster **MongoDB Atlas** — não a um Mongo standalone: o fluxo de `POST /auth/register` usa transações (`session.WithTransaction`), que exigem um replica set, e Atlas sempre é um replica set (mesmo no tier gratuito M0).

Variáveis de ambiente esperadas em produção (configuradas direto no Railway, nunca commitadas):

| Variável | Descrição |
|---|---|
| `MONGO_URI` | Connection string do Atlas (`mongodb+srv://...`) |
| `MONGO_DB_NAME` | `gestorbuy` |
| `JWT_SECRET` | Segredo forte, exclusivo de produção — nunca reaproveitar o placeholder do `.env.example` |
| `PORT` | Injetada automaticamente pelo Railway — não precisa configurar |

Health check do Railway aponta para `GET /health`.

## Próximas entregas

- Hub OAuth2 com Mercado Livre e Shopee
- Criptografia de campo (CSFLE/Queryable Encryption) para credenciais de marketplace
- Módulo `products` (catálogo unificado, seção 2.1 do doc de arquitetura)
- Deploy em MongoDB Atlas real
