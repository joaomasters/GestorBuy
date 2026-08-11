// apiFetch é o único ponto de contato entre o Next.js e a API Go — sempre
// chamado do servidor (Server Components, Server Actions, Route Handlers),
// nunca do navegador. Injeta o JWT lido do cookie httpOnly, então o token
// nunca fica exposto a JavaScript do cliente (padrão BFF combinado com
// proxy.ts e os Route Handlers de app/api/auth/*).
import { cookies } from "next/headers";
import { SESSION_COOKIE } from "./session";

const API_URL = process.env.INTERNAL_API_URL;

export class ApiError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

export async function apiFetch<T = unknown>(
  path: string,
  init: RequestInit = {}
): Promise<T> {
  if (!API_URL) {
    throw new Error(
      "INTERNAL_API_URL não configurado — defina em .env.local (ver .env.local.example)"
    );
  }

  const cookieStore = await cookies();
  const token = cookieStore.get(SESSION_COOKIE)?.value;

  const headers = new Headers(init.headers);
  headers.set("Content-Type", "application/json");
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  const res = await fetch(`${API_URL}${path}`, {
    ...init,
    headers,
    cache: "no-store",
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({}) as { error?: string });
    throw new ApiError(res.status, body.error ?? `Erro ${res.status}`);
  }

  if (res.status === 204) {
    return undefined as T;
  }
  return res.json() as Promise<T>;
}
