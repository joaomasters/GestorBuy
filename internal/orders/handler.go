package orders

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
	mux.Handle("POST /orders/sync", authMiddleware(http.HandlerFunc(h.sync)))
	mux.Handle("GET /orders", authMiddleware(http.HandlerFunc(h.list)))
}

func (h *Handler) sync(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "não autenticado")
		return
	}

	count, err := h.svc.SyncOrders(r.Context(), claims.TenantID)
	if err != nil {
		h.handleServiceError(w, err, "sync")
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"synced": count})
}

type orderItemDTO struct {
	Title     string  `json:"title"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}

type orderDTO struct {
	ID              string         `json:"id"`
	Marketplace     string         `json:"marketplace"`
	ExternalOrderID string         `json:"external_order_id"`
	Status          string         `json:"status"`
	TotalAmount     float64        `json:"total_amount"`
	BuyerNickname   string         `json:"buyer_nickname,omitempty"`
	Items           []orderItemDTO `json:"items"`
	DateCreated     string         `json:"date_created"`
}

const timeFormat = "2006-01-02T15:04:05Z07:00"

func toOrderDTO(o Order) orderDTO {
	items := make([]orderItemDTO, len(o.Items))
	for i, it := range o.Items {
		items[i] = orderItemDTO{Title: it.Title, Quantity: it.Quantity, UnitPrice: it.UnitPrice}
	}
	return orderDTO{
		ID:              o.ID,
		Marketplace:     o.Marketplace,
		ExternalOrderID: o.ExternalOrderID,
		Status:          o.Status,
		TotalAmount:     o.TotalAmount,
		BuyerNickname:   o.BuyerNickname,
		Items:           items,
		DateCreated:     o.DateCreated.Format(timeFormat),
	}
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "não autenticado")
		return
	}

	list, err := h.svc.List(r.Context(), claims.TenantID)
	if err != nil {
		h.log.Error("orders.list falhou", "error", err)
		writeError(w, http.StatusInternalServerError, "não foi possível listar pedidos")
		return
	}

	out := make([]orderDTO, len(list))
	for i, o := range list {
		out[i] = toOrderDTO(o)
	}
	writeJSON(w, http.StatusOK, map[string]any{"orders": out})
}

func (h *Handler) handleServiceError(w http.ResponseWriter, err error, op string) {
	switch {
	case errors.Is(err, ErrMarketplaceNotConnected):
		writeError(w, http.StatusConflict, err.Error())
	default:
		h.log.Error("orders."+op+" falhou", "error", err)
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
