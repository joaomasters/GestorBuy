package auth

import (
	"context"
	"net/http"
	"strings"
)

type ctxKey int

const claimsCtxKey ctxKey = iota

// Middleware extrai e valida o JWT do header Authorization e injeta as
// Claims (incluindo tenant_id) no contexto da requisição. Toda rota
// protegida deve passar por aqui antes de chegar em qualquer handler que
// acesse dados de um tenant — é o único ponto de entrada do tenant_id no
// resto do sistema (seção 2.3 do doc de arquitetura: nunca confiar em
// tenant_id vindo de query param/body).
func (s *Service) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		raw, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || raw == "" {
			writeError(w, http.StatusUnauthorized, "token de autenticação ausente")
			return
		}

		claims, err := s.ParseToken(raw)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "token de autenticação inválido ou expirado")
			return
		}

		ctx := context.WithValue(r.Context(), claimsCtxKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ClaimsFromContext recupera as claims injetadas pelo Middleware. Handlers
// downstream chamam isso para obter o tenant_id da requisição autenticada —
// nunca leem tenant_id de outro lugar.
func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(claimsCtxKey).(*Claims)
	return claims, ok
}
