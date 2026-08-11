# GESTORBUY — Frontend

Painel web do GESTORBUY (Next.js 16 + TypeScript + Tailwind). Consome a API Go
em [../README.md](../README.md) via padrão **BFF** (Backend for Frontend):
o navegador só fala com este app — o JWT da API fica num cookie `httpOnly`
setado pelos Route Handlers de `app/api/auth/*`, nunca exposto a JavaScript
do cliente.

## Rodando localmente

```bash
cp .env.local.example .env.local
# ajuste INTERNAL_API_URL se a API não estiver em localhost:8080

npm install
npm run dev
```

Abra [http://localhost:3000](http://localhost:3000) — a API Go
([../README.md](../README.md)) precisa estar rodando em paralelo.

## Estrutura

```
app/
  api/auth/          Route Handlers: proxy pro /auth/* da API Go + cookie httpOnly
  login/, register/  Telas públicas
  products/          Área autenticada (protegida por proxy.ts)
    actions.ts         Server Actions (create/update/delete produto)
    product-form.tsx   Formulário compartilhado entre criar/editar
lib/
  api.ts             apiFetch() — único ponto de chamada à API Go, injeta o JWT do cookie
  session.ts         Nome/TTL do cookie de sessão
proxy.ts             Protege /products/* (convenção do Next.js 16, ex-middleware.ts)
```

## Deploy

Produção na **Vercel** (projeto `jgp-tech/web`), conectado ao GitHub
(`joaomasters/GestorBuy`, branch `main`, **Root Directory = `web`** — o
repositório é um monorepo com a API Go na raiz).

Variável de ambiente em produção: `INTERNAL_API_URL` apontando para o
serviço da API no Railway. Configurada direto no dashboard/CLI da Vercel,
nunca commitada.

```bash
npm run build   # valida build + typecheck antes de qualquer deploy manual
vercel --prod   # deploy manual (o normal é o auto-deploy no push pra main)
```
