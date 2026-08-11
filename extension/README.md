# GESTORBUY — Extensão de Mineração (Mercado Livre)

Extensão Chrome (Manifest V3) que lê preço, quantidade vendida e vendedor
direto da página de um anúncio do Mercado Livre — **na sua própria sessão
de navegação**, não via API do backend. É o que permite minerar anúncio de
concorrente: o servidor não consegue ler isso (o Mercado Livre bloqueia),
mas o seu navegador, logado normalmente, consegue — porque pra ele é só
você navegando.

## ⚠️ Estado atual: não testada contra o site real

Os seletores CSS em [src/content-script.ts](src/content-script.ts) (título,
preço, vendidos, vendedor) foram escritos com base no que eu sei da
estrutura do Mercado Livre, sem validação ao vivo — não tenho como abrir um
navegador daqui. **Bem provável que precisem de ajuste.** Depois de instalar
e testar (abaixo), se o popup mostrar "(não identificado)" em algum campo,
me manda um print da página do anúncio que eu ajusto o seletor.

## Build

```bash
cd extension
npm install
npm run build      # gera background.js, content-script.js, popup.js
# ou: npm run watch   (rebuilda automaticamente a cada mudança)
```

Testar contra o backend local em vez de produção: troca `API_URL` em
[src/config.ts](src/config.ts) pra `http://localhost:8080` antes de buildar.

## Instalar no Chrome (modo desenvolvedor)

1. `chrome://extensions`
2. Ativa **"Modo do desenvolvedor"** (canto superior direito)
3. **"Carregar sem compactação"** → seleciona a pasta `extension/` (a raiz,
   não `src/`)
4. O ícone do GestorBuy aparece na barra de extensões

## Usando

1. Clica no ícone → faz login com o e-mail/senha da sua conta GestorBuy
2. Navega até a página de um anúncio no `mercadolivre.com.br` (seu ou de
   concorrente — tanto faz, não tem restrição aqui)
3. Clica no ícone de novo → o popup mostra uma prévia do que foi extraído
4. **"Salvar no GestorBuy"** → grava o snapshot, aparece em `/mining` no
   painel web

## Depois de mudar o código

Rodar `npm run build` de novo e clicar no botão de recarregar (ícone de
setas circulares) no card da extensão em `chrome://extensions` — o Chrome
não recarrega automaticamente.
