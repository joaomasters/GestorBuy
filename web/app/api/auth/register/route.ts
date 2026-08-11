// Mesmo padrão de app/api/auth/login/route.ts, mas pro POST /auth/register
// (cria tenant + usuário admin atomicamente no backend — ver
// internal/auth/service.go).
import { NextResponse } from "next/server";
import { SESSION_COOKIE, SESSION_MAX_AGE_SECONDS } from "@/lib/session";

const API_URL = process.env.INTERNAL_API_URL;

export async function POST(request: Request) {
  if (!API_URL) {
    return NextResponse.json(
      { error: "INTERNAL_API_URL não configurado" },
      { status: 500 }
    );
  }

  const body = await request.json().catch(() => null);
  if (!body) {
    return NextResponse.json(
      { error: "corpo da requisição inválido" },
      { status: 400 }
    );
  }

  const backendRes = await fetch(`${API_URL}/auth/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
    cache: "no-store",
  });

  const data = await backendRes
    .json()
    .catch(() => ({}) as { error?: string; token?: string });

  if (!backendRes.ok || !data.token) {
    return NextResponse.json(
      { error: data.error ?? "não foi possível concluir o registro" },
      { status: backendRes.status || 500 }
    );
  }

  const response = NextResponse.json({ ok: true });
  response.cookies.set(SESSION_COOKIE, data.token, {
    httpOnly: true,
    secure: process.env.NODE_ENV === "production",
    sameSite: "lax",
    path: "/",
    maxAge: SESSION_MAX_AGE_SECONDS,
  });
  return response;
}
