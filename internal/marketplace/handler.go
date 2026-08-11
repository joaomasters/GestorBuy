package marketplace

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gestorbuy/api/internal/auth"
)

type Handler struct {
	svc         *Service
	log         *slog.Logger
	frontendURL string
}

func NewHandler(svc *Service, log *slog.Logger, frontendURL string) *Handler {
	return &Handler{svc: svc, log: log, frontendURL: frontendURL}
}

// Routes registra as rotas protegidas. O callback OAuth é registrado à
// parte por RoutesPublic — o Mercado Livre chama ele direto (redirect do
// navegador do usuário), sem o cookie/JWT da nossa sessão.
func (h *Handler) Routes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler) {
	mux.Handle("GET /integrations", authMiddleware(http.HandlerFunc(h.status)))
	mux.Handle("GET /integrations/mercadolivre/connect", authMiddleware(http.HandlerFunc(h.connect)))
	mux.Handle("DELETE /integrations/mercadolivre", authMiddleware(http.HandlerFunc(h.disconnect)))
	mux.Handle("POST /products/{id}/channels/mercadolivre/sync", authMiddleware(http.HandlerFunc(h.syncProductChannel)))
}

func (h *Handler) RoutesPublic(mux *http.ServeMux) {
	mux.HandleFunc("GET /integrations/mercadolivre/callback", h.callback)
}

func (h *Handler) connect(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "não autenticado")
		return
	}

	authURL, err := h.svc.ConnectURL(claims.TenantID)
	if err != nil {
		h.handleServiceError(w, err, "connect")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"auth_url": authURL})
}

// callback é chamado diretamente pelo Mercado Livre (redirect do navegador
// do usuário) — endpoint público, autenticação é o próprio `state` assinado.
func (h *Handler) callback(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	code := q.Get("code")
	state := q.Get("state")

	frontendURL := h.frontendURL

	if code == "" || state == "" {
		http.Redirect(w, r, frontendURL+"/integrations?error=missing_code_or_state", http.StatusFound)
		return
	}

	if _, err := h.svc.HandleCallback(r.Context(), code, state); err != nil {
		h.log.Error("marketplace.callback falhou", "error", err)
		http.Redirect(w, r, frontendURL+"/integrations?error=connect_failed", http.StatusFound)
		return
	}

	http.Redirect(w, r, frontendURL+"/integrations?connected=1", http.StatusFound)
}

func (h *Handler) status(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "não autenticado")
		return
	}

	statuses, err := h.svc.Status(r.Context(), claims.TenantID)
	if err != nil {
		h.log.Error("marketplace.status falhou", "error", err)
		writeError(w, http.StatusInternalServerError, "não foi possível consultar integrações")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"integrations": statuses})
}

func (h *Handler) disconnect(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "não autenticado")
		return
	}

	if err := h.svc.Disconnect(r.Context(), claims.TenantID); err != nil {
		h.handleServiceError(w, err, "disconnect")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) syncProductChannel(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "não autenticado")
		return
	}

	if err := h.svc.SyncProductChannel(r.Context(), claims.TenantID, r.PathValue("id")); err != nil {
		h.handleServiceError(w, err, "sync_product_channel")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "synced"})
}

func (h *Handler) handleServiceError(w http.ResponseWriter, err error, op string) {
	switch {
	case errors.Is(err, ErrNotConfigured):
		writeError(w, http.StatusServiceUnavailable, err.Error())
	case errors.Is(err, ErrNotConnected), errors.Is(err, ErrInvalidState):
		writeError(w, http.StatusConflict, err.Error())
	default:
		h.log.Error("marketplace."+op+" falhou", "error", err)
		writeError(w, http.StatusInternalServerError, "não foi possível processar a requisição")
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
