// O navegador precisa ir pro Mercado Livre de verdade (é ele quem autoriza
// o app) — por isso esse Route Handler não devolve JSON, ele redireciona.
// Chama o backend Go (autenticado com o JWT do cookie) só pra montar a URL
// de autorização com o `state` assinado.
import { NextResponse } from "next/server";
import { cookies } from "next/headers";
import { SESSION_COOKIE } from "@/lib/session";

const API_URL = process.env.INTERNAL_API_URL;

export async function GET(request: Request) {
  if (!API_URL) {
    return NextResponse.json(
      { error: "INTERNAL_API_URL não configurado" },
      { status: 500 }
    );
  }

  const cookieStore = await cookies();
  const token = cookieStore.get(SESSION_COOKIE)?.value;
  if (!token) {
    return NextResponse.redirect(new URL("/login", request.url));
  }

  const res = await fetch(`${API_URL}/integrations/mercadolivre/connect`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });

  const data = await res
    .json()
    .catch(() => ({}) as { error?: string; auth_url?: string });

  if (!res.ok || !data.auth_url) {
    const redirectURL = new URL("/integrations", request.url);
    redirectURL.searchParams.set("error", data.error ?? "connect_failed");
    return NextResponse.redirect(redirectURL);
  }

  return NextResponse.redirect(data.auth_url);
}
