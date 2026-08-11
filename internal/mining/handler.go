package mining

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

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
	mux.Handle("POST /mining/items", authMiddleware(http.HandlerFunc(h.track)))
	mux.Handle("GET /mining/items", authMiddleware(http.HandlerFunc(h.list)))
	mux.Handle("POST /mining/items/{id}/refresh", authMiddleware(http.HandlerFunc(h.refresh)))
	mux.Handle("GET /mining/items/{id}/history", authMiddleware(http.HandlerFunc(h.history)))
}

type snapshotDTO struct {
	TS                    string  `json:"ts"`
	Price                 float64 `json:"price"`
	SoldQuantity          int64   `json:"sold_quantity"`
	AvailableQuantity     int64   `json:"available_quantity"`
	EstimatedRevenueTotal float64 `json:"estimated_revenue_total"`
}

func toSnapshotDTO(s Snapshot) snapshotDTO {
	return snapshotDTO{
		TS:                    s.TS.Format(timeFormat),
		Price:                 s.Price,
		SoldQuantity:          s.SoldQuantity,
		AvailableQuantity:     s.AvailableQuantity,
		EstimatedRevenueTotal: s.EstimatedRevenueTotal,
	}
}

const timeFormat = "2006-01-02T15:04:05Z07:00"

type itemDTO struct {
	ID             string       `json:"id"`
	Marketplace    string       `json:"marketplace"`
	ExternalItemID string       `json:"external_item_id"`
	Title          string       `json:"title"`
	Permalink      string       `json:"permalink,omitempty"`
	CategoryID     string       `json:"category_id,omitempty"`
	LatestSnapshot *snapshotDTO `json:"latest_snapshot,omitempty"`
}

func toItemDTO(item Item, snap *Snapshot) itemDTO {
	dto := itemDTO{
		ID:             item.ID,
		Marketplace:    item.Marketplace,
		ExternalItemID: item.ExternalItemID,
		Title:          item.Title,
		Permalink:      item.Permalink,
		CategoryID:     item.CategoryID,
	}
	if snap != nil {
		s := toSnapshotDTO(*snap)
		dto.LatestSnapshot = &s
	}
	return dto
}

type trackRequest struct {
	URLOrID string `json:"url_or_id"`
}

func (h *Handler) track(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "não autenticado")
		return
	}

	var req trackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	item, snap, err := h.svc.TrackAndSnapshot(r.Context(), claims.TenantID, req.URLOrID)
	if err != nil {
		h.handleServiceError(w, err, "track")
		return
	}
	writeJSON(w, http.StatusCreated, toItemDTO(*item, snap))
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "não autenticado")
		return
	}

	items, err := h.svc.List(r.Context(), claims.TenantID)
	if err != nil {
		h.log.Error("mining.list falhou", "error", err)
		writeError(w, http.StatusInternalServerError, "não foi possível listar itens rastreados")
		return
	}

	out := make([]itemDTO, len(items))
	for i, it := range items {
		out[i] = toItemDTO(it.Item, it.Snapshot)
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out})
}

func (h *Handler) refresh(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "não autenticado")
		return
	}

	item, snap, err := h.svc.RefreshByID(r.Context(), claims.TenantID, r.PathValue("id"))
	if err != nil {
		h.handleServiceError(w, err, "refresh")
		return
	}
	writeJSON(w, http.StatusOK, toItemDTO(*item, snap))
}

func (h *Handler) history(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "não autenticado")
		return
	}

	item, history, err := h.svc.History(r.Context(), claims.TenantID, r.PathValue("id"))
	if err != nil {
		h.handleServiceError(w, err, "history")
		return
	}

	snapshots := make([]snapshotDTO, len(history))
	for i, s := range history {
		snapshots[i] = toSnapshotDTO(s)
	}
	writeJSON(w, http.StatusOK, map[string]any{"item": toItemDTO(*item, nil), "snapshots": snapshots})
}

func (h *Handler) handleServiceError(w http.ResponseWriter, err error, op string) {
	switch {
	case errors.Is(err, ErrInvalidItemReference):
		writeError(w, http.StatusUnprocessableEntity, err.Error())
	default:
		h.log.Error("mining."+op+" falhou", "error", err)
		writeError(w, http.StatusBadGateway, err.Error())
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
