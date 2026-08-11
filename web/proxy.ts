// Nota: no Next.js 16 o arquivo de proxy/middleware se chama `proxy.ts`
// (renomeado de `middleware.ts`) — ver AGENTS.md gerado pelo create-next-app.
//
// Só verifica a PRESENÇA do cookie de sessão, não decodifica o JWT (isso é
// trabalho da API Go). Se o token estiver expirado, a chamada a apiFetch()
// retorna 401 e a página trata redirecionando pra /login — o proxy sozinho
// não é a única linha de defesa (Server Actions também verificam a sessão,
// e a autorização de verdade é sempre validada pelo backend).
import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";
import { SESSION_COOKIE } from "@/lib/session";

export function proxy(request: NextRequest) {
  if (!request.cookies.has(SESSION_COOKIE)) {
    return NextResponse.redirect(new URL("/login", request.url));
  }
}

export const config = {
  matcher: [
    "/products/:path*",
    "/integrations/:path*",
    "/mining/:path*",
    "/orders/:path*",
    "/margin/:path*",
    "/extension/:path*",
  ],
};
