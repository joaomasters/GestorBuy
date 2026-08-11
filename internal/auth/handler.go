package auth

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gestorbuy/api/internal/tenant"
)

type Handler struct {
	svc *Service
	log *slog.Logger
}

func NewHandler(svc *Service, log *slog.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

// Routes registra as rotas públicas de auth. Routes protegidas (que exigem
// JWT válido) são registradas separadamente pelo chamador, envolvidas em
// svc.Middleware — ver internal/platform/httpserver/server.go.
func (h *Handler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /auth/register", h.register)
	mux.HandleFunc("POST /auth/login", h.login)
}

// Me é um endpoint protegido mínimo que devolve as claims do token — prova
// de que o AuthMiddleware está resolvendo tenant_id/user_id corretamente a
// partir do JWT. Serve de modelo para os handlers reais dos próximos módulos
// (Gestor/Analytics), que vão ler ClaimsFromContext do mesmo jeito.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "não autenticado")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"tenant_id": claims.TenantID,
		"user_id":   claims.UserID,
		"role":      string(claims.Role),
	})
}

type registerRequest struct {
	TenantName string `json:"tenant_name"`
	TenantSlug string `json:"tenant_slug"`
	Email      string `json:"email"`
	Password   string `json:"password"`
}

type registerResponse struct {
	TenantID string `json:"tenant_id"`
	UserID   string `json:"user_id"`
	Token    string `json:"token"`
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}
	if req.TenantName == "" || req.TenantSlug == "" || req.Email == "" || len(req.Password) < 8 {
		writeError(w, http.StatusUnprocessableEntity, "tenant_name, tenant_slug, email e password (mín. 8 caracteres) são obrigatórios")
		return
	}

	result, err := h.svc.Register(r.Context(), RegisterInput{
		TenantName: req.TenantName,
		TenantSlug: req.TenantSlug,
		Email:      req.Email,
		Password:   req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrEmailTaken), errors.Is(err, tenant.ErrSlugTaken):
			writeError(w, http.StatusConflict, err.Error())
		default:
			h.log.Error("auth.register falhou", "error", err)
			writeError(w, http.StatusInternalServerError, "não foi possível concluir o registro")
		}
		return
	}

	writeJSON(w, http.StatusCreated, registerResponse{TenantID: result.TenantID, UserID: result.UserID, Token: result.Token})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	token, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "e-mail ou senha inválidos")
			return
		}
		h.log.Error("auth.login falhou", "error", err)
		writeError(w, http.StatusInternalServerError, "não foi possível concluir o login")
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{Token: token})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
