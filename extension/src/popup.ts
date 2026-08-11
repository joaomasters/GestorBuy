import type { ExtractedItem } from "./content-script";

function $(id: string): HTMLElement {
  const el = document.getElementById(id);
  if (!el) throw new Error(`elemento #${id} não encontrado no popup.html`);
  return el;
}

const loginView = $("login-view");
const appView = $("app-view");
const emailInput = $("email") as HTMLInputElement;
const passwordInput = $("password") as HTMLInputElement;
const loginBtn = $("login-btn") as HTMLButtonElement;
const loginError = $("login-error");

const notProductPage = $("not-product-page");
const itemPreview = $("item-preview");
const saveBtn = $("save-btn") as HTMLButtonElement;
const saveMessage = $("save-message");
const logoutBtn = $("logout") as HTMLButtonElement;

let currentItem: ExtractedItem | null = null;

async function refreshView() {
  const { token } = await chrome.runtime.sendMessage({ type: "GET_TOKEN" });

  if (!token) {
    loginView.hidden = false;
    appView.hidden = true;
    return;
  }

  loginView.hidden = true;
  appView.hidden = false;
  await loadCurrentPageItem();
}

async function loadCurrentPageItem() {
  const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
  if (!tab?.id || !tab.url?.includes("mercadolivre.com.br")) {
    notProductPage.hidden = false;
    itemPreview.hidden = true;
    saveBtn.hidden = true;
    return;
  }

  try {
    const item: ExtractedItem = await chrome.tabs.sendMessage(tab.id, {
      type: "EXTRACT_ITEM",
    });
    currentItem = item;
    notProductPage.hidden = true;
    itemPreview.hidden = false;
    saveBtn.hidden = false;

    $("preview-title").textContent = item.title || "(não identificado)";
    $("preview-price").textContent = item.price
      ? item.price.toLocaleString("pt-BR", { style: "currency", currency: "BRL" })
      : "(não identificado)";
    $("preview-sold").textContent = item.soldQuantity
      ? String(item.soldQuantity)
      : "(não identificado)";
    $("preview-seller").textContent = item.sellerNickname || "(não identificado)";
  } catch {
    // Content script não respondeu (página ainda carregando, ou não é uma
    // página de produto de verdade) — trata como "não é página de produto".
    notProductPage.hidden = false;
    itemPreview.hidden = true;
    saveBtn.hidden = true;
  }
}

loginBtn.addEventListener("click", async () => {
  loginError.hidden = true;
  loginBtn.disabled = true;
  loginBtn.textContent = "Entrando...";

  const result = await chrome.runtime.sendMessage({
    type: "LOGIN",
    email: emailInput.value,
    password: passwordInput.value,
  });

  loginBtn.disabled = false;
  loginBtn.textContent = "Entrar";

  if (!result.ok) {
    loginError.textContent = result.error;
    loginError.hidden = false;
    return;
  }

  await refreshView();
});

saveBtn.addEventListener("click", async () => {
  if (!currentItem) return;

  saveBtn.disabled = true;
  saveBtn.textContent = "Salvando...";
  saveMessage.hidden = true;

  const result = await chrome.runtime.sendMessage({
    type: "SAVE_ITEM",
    item: {
      external_item_id: currentItem.externalItemId,
      title: currentItem.title,
      permalink: currentItem.permalink,
      seller_nickname: currentItem.sellerNickname,
      price: currentItem.price,
      sold_quantity: currentItem.soldQuantity,
      available_quantity: 0,
    },
  });

  saveBtn.disabled = false;
  saveBtn.textContent = "Salvar no GestorBuy";

  saveMessage.hidden = false;
  saveMessage.className = result.ok ? "success" : "error";
  saveMessage.textContent = result.ok
    ? "Salvo! Confira em /mining no GestorBuy."
    : (result.error ?? "Não foi possível salvar");
});

logoutBtn.addEventListener("click", async () => {
  await chrome.runtime.sendMessage({ type: "LOGOUT" });
  await refreshView();
});

refreshView();
