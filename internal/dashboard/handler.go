package dashboard

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/gestorbuy/api/internal/auth"
)

type Handler struct {
	svc *Service
	log *slog.Logger
}

func NewHandler(svc *Service, log *slog.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

func (h *Handler) Routes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler) {
	mux.Handle("GET /dashboard/summary", authMiddleware(http.HandlerFunc(h.summary)))
}

const dateFormat = "2006-01-02"

func (h *Handler) summary(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "não autenticado")
		return
	}

	now := time.Now().UTC()
	from := now.AddDate(0, 0, -30)
	to := now

	if raw := r.URL.Query().Get("from"); raw != "" {
		parsed, err := time.Parse(dateFormat, raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "from inválido, use YYYY-MM-DD")
			return
		}
		from = parsed
	}
	if raw := r.URL.Query().Get("to"); raw != "" {
		parsed, err := time.Parse(dateFormat, raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "to inválido, use YYYY-MM-DD")
			return
		}
		// Inclui o dia inteiro do "to" — sem isso, pedidos feitos depois da
		// meia-noite desse dia ficariam de fora.
		to = parsed.Add(24*time.Hour - time.Second)
	}

	summary, err := h.svc.Summary(r.Context(), claims.TenantID, from, to)
	if err != nil {
		h.log.Error("dashboard.summary falhou", "error", err)
		writeError(w, http.StatusInternalServerError, "não foi possível calcular o resumo")
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
