// Injetado nas páginas de produto do Mercado Livre. Lê os dados que já
// estão renderizados na tela — mesma coisa que o usuário vê com os
// próprios olhos, na própria sessão autenticada dele.
//
// AVISO: os seletores abaixo são minha melhor estimativa da estrutura do
// site (não validados contra uma página real nesta sessão — sem acesso a
// navegador aqui). Bem provável que precisem de ajuste. Se a extração vier
// vazia ou errada, me manda um print da página + o que apareceu no popup
// que eu ajusto os seletores.
//
// Estratégia deliberada: várias tentativas em cascata (seletor específico
// → seletor mais genérico → regex sobre o texto da página), pra degradar
// melhor em vez de falhar tudo se o ML mudar uma classe CSS.

export type ExtractedItem = {
  externalItemId: string; // manda a URL inteira — o backend já sabe extrair o MLB... dela (mesma regex de internal/mining.ParseItemID)
  title: string;
  permalink: string;
  sellerNickname: string;
  price: number;
  soldQuantity: number;
};

function queryText(selectors: string[]): string | null {
  for (const selector of selectors) {
    const el = document.querySelector(selector);
    if (el?.textContent?.trim()) return el.textContent.trim();
  }
  return null;
}

function parseBRLNumber(raw: string): number {
  // "1.234,56" -> 1234.56 ; "1.234" (só a parte inteira, sem vírgula) -> 1234
  const cleaned = raw.replace(/[^\d.,]/g, "");
  if (cleaned.includes(",")) {
    return parseFloat(cleaned.replace(/\./g, "").replace(",", "."));
  }
  return parseFloat(cleaned.replace(/\./g, ""));
}

function extractTitle(): string {
  return (
    queryText([".ui-pdp-title", "h1.ui-pdp-title", "h1"]) ??
    document.title.replace(/\s*\|\s*MercadoLivre.*$/i, "").trim()
  );
}

function extractPrice(): number {
  // Preço "principal" da página — tenta o container mais específico
  // primeiro (evita pegar preço "de/por" riscado ou parcelamento).
  const candidates = [
    ".ui-pdp-price__second-line .andes-money-amount__fraction",
    ".ui-pdp-price__main-container .andes-money-amount__fraction",
    ".andes-money-amount__fraction",
  ];
  for (const selector of candidates) {
    const el = document.querySelector(selector);
    if (el?.textContent) {
      const value = parseBRLNumber(el.textContent);
      if (!Number.isNaN(value) && value > 0) return value;
    }
  }
  return 0;
}

function extractSoldQuantity(): number {
  // "vendidos" costuma aparecer perto do subtítulo/avaliações — tenta uma
  // área específica primeiro, depois cai pra procurar em todo o texto da
  // página (mais lento, mas mais resiliente a mudança de layout).
  const scoped = queryText([".ui-pdp-subtitle", ".ui-pdp-header__subtitle"]);
  const scopedMatch = scoped?.match(/([\d.,]+)\s*vendid/i);
  if (scopedMatch) return Math.round(parseBRLNumber(scopedMatch[1]));

  const bodyMatch = document.body.innerText.match(/([\d.,]+)\s*vendid/i);
  if (bodyMatch) return Math.round(parseBRLNumber(bodyMatch[1]));

  return 0;
}

function extractSellerNickname(): string {
  const scoped = queryText([
    ".ui-pdp-seller__link-trigger-button",
    ".ui-pdp-seller__header__title",
  ]);
  if (scoped) return scoped;

  const bodyMatch = document.body.innerText.match(/Vendido por\s+([^\n]+)/i);
  return bodyMatch ? bodyMatch[1].trim() : "";
}

function extractPermalink(): string {
  const canonical = document.querySelector('link[rel="canonical"]');
  return canonical?.getAttribute("href") ?? window.location.href;
}

function extractItem(): ExtractedItem {
  return {
    externalItemId: window.location.href,
    title: extractTitle(),
    permalink: extractPermalink(),
    sellerNickname: extractSellerNickname(),
    price: extractPrice(),
    soldQuantity: extractSoldQuantity(),
  };
}

chrome.runtime.onMessage.addListener((message, _sender, sendResponse) => {
  if (message?.type === "EXTRACT_ITEM") {
    sendResponse(extractItem());
  }
  return true;
});
