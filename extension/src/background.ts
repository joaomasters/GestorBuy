// Service worker (Manifest V3) — guarda o JWT e é o único lugar que fala
// com o backend do GestorBuy. A extensão faz login próprio (não participa
// do cookie httpOnly do web/) porque não tem como compartilhar sessão com
// o navegador do mesmo jeito que o app Next.js faz.
import { API_URL, STORAGE_KEY_TOKEN } from "./config";

type Message =
  | { type: "LOGIN"; email: string; password: string }
  | { type: "GET_TOKEN" }
  | { type: "LOGOUT" }
  | {
      type: "SAVE_ITEM";
      item: {
        external_item_id: string;
        title: string;
        permalink: string;
        seller_nickname: string;
        price: number;
        sold_quantity: number;
        available_quantity: number;
      };
    };

async function getToken(): Promise<string | null> {
  const stored = await chrome.storage.local.get(STORAGE_KEY_TOKEN);
  return stored[STORAGE_KEY_TOKEN] ?? null;
}

async function handleLogin(email: string, password: string) {
  const res = await fetch(`${API_URL}/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password }),
  });
  const data = await res.json().catch(() => ({}) as { error?: string; token?: string });

  if (!res.ok || !data.token) {
    return { ok: false, error: data.error ?? "Não foi possível entrar" };
  }

  await chrome.storage.local.set({ [STORAGE_KEY_TOKEN]: data.token });
  return { ok: true };
}

async function handleSaveItem(item: Extract<Message, { type: "SAVE_ITEM" }>["item"]) {
  const token = await getToken();
  if (!token) {
    return { ok: false, error: "Faça login na extensão primeiro" };
  }

  const res = await fetch(`${API_URL}/mining/items/ingest`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify(item),
  });
  const data = await res.json().catch(() => ({}) as { error?: string });

  if (!res.ok) {
    return { ok: false, error: data.error ?? `Erro ${res.status}` };
  }
  return { ok: true };
}

chrome.runtime.onMessage.addListener((message: Message, _sender, sendResponse) => {
  (async () => {
    switch (message.type) {
      case "LOGIN":
        sendResponse(await handleLogin(message.email, message.password));
        break;
      case "GET_TOKEN":
        sendResponse({ token: await getToken() });
        break;
      case "LOGOUT":
        await chrome.storage.local.remove(STORAGE_KEY_TOKEN);
        sendResponse({ ok: true });
        break;
      case "SAVE_ITEM":
        sendResponse(await handleSaveItem(message.item));
        break;
    }
  })();

  return true; // mantém o canal aberto pra resposta assíncrona
});
