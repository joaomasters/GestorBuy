package product

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gestorbuy/api/internal/auth"
)

type Handler struct {
	svc *Service
	log *slog.Logger
}

func NewHandler(svc *Service, log *slog.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

// Routes registra as rotas de produto, todas atrás de authMiddleware
// (tipicamente auth.Service.Middleware — recebido como parâmetro pra não
// criar dependência direta do pacote product no pacote auth).
func (h *Handler) Routes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler) {
	mux.Handle("POST /products", authMiddleware(http.HandlerFunc(h.create)))
	mux.Handle("GET /products", authMiddleware(http.HandlerFunc(h.list)))
	mux.Handle("GET /products/{id}", authMiddleware(http.HandlerFunc(h.get)))
	mux.Handle("PATCH /products/{id}", authMiddleware(http.HandlerFunc(h.update)))
	mux.Handle("DELETE /products/{id}", authMiddleware(http.HandlerFunc(h.delete)))
	mux.Handle("POST /products/{id}/channels/{marketplace}", authMiddleware(http.HandlerFunc(h.linkChannel)))
}

type variationDTO struct {
	VariationSKU string            `json:"variation_sku"`
	Attributes   map[string]string `json:"attributes,omitempty"`
	StockTotal   int               `json:"stock_total"`
	CostPrice    float64           `json:"cost_price"`
	Price        float64           `json:"price"`
}

func toVariations(dtos []variationDTO) []Variation {
	out := make([]Variation, len(dtos))
	for i, d := range dtos {
		out[i] = Variation{
			VariationSKU: d.VariationSKU,
			Attributes:   d.Attributes,
			StockTotal:   d.StockTotal,
			CostPrice:    d.CostPrice,
			Price:        d.Price,
		}
	}
	return out
}

type channelDTO struct {
	ItemID       string  `json:"item_id,omitempty"`
	SyncStatus   string  `json:"sync_status,omitempty"`
	LastSyncedAt *string `json:"last_synced_at,omitempty"`
	LastError    string  `json:"last_error,omitempty"`
}

type productResponse struct {
	ID                 string                `json:"id"`
	SKUMaster          string                `json:"sku_master"`
	Title              string                `json:"title"`
	Brand              string                `json:"brand,omitempty"`
	CategoryNormalized string                `json:"category_normalized,omitempty"`
	Variations         []variationDTO        `json:"variations"`
	StockStrategy      string                `json:"stock_strategy"`
	Channels           map[string]channelDTO `json:"channels,omitempty"`
	CreatedAt          string                `json:"created_at"`
	UpdatedAt          string                `json:"updated_at"`
}

func toResponse(p Product) productResponse {
	variations := make([]variationDTO, len(p.Variations))
	for i, v := range p.Variations {
		variations[i] = variationDTO{
			VariationSKU: v.VariationSKU,
			Attributes:   v.Attributes,
			StockTotal:   v.StockTotal,
			CostPrice:    v.CostPrice,
			Price:        v.Price,
		}
	}

	var channels map[string]channelDTO
	if len(p.Channels) > 0 {
		channels = make(map[string]channelDTO, len(p.Channels))
		for marketplace, ch := range p.Channels {
			dto := channelDTO{ItemID: ch.ItemID, SyncStatus: ch.SyncStatus, LastError: ch.LastError}
			if ch.LastSyncedAt != nil {
				formatted := ch.LastSyncedAt.Format(timeFormat)
				dto.LastSyncedAt = &formatted
			}
			channels[marketplace] = dto
		}
	}

	return productResponse{
		ID:                 p.ID,
		SKUMaster:          p.SKUMaster,
		Title:              p.Title,
		Brand:              p.Brand,
		CategoryNormalized: p.CategoryNormalized,
		Variations:         variations,
		StockStrategy:      string(p.StockStrategy),
		Channels:           channels,
		CreatedAt:          p.CreatedAt.Format(timeFormat),
		UpdatedAt:          p.UpdatedAt.Format(timeFormat),
	}
}

const timeFormat = "2006-01-02T15:04:05Z07:00"

type createRequest struct {
	SKUMaster          string         `json:"sku_master"`
	Title              string         `json:"title"`
	Brand              string         `json:"brand"`
	CategoryNormalized string         `json:"category_normalized"`
	Variations         []variationDTO `json:"variations"`
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "não autenticado")
		return
	}

	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	p, err := h.svc.Create(r.Context(), claims.TenantID, CreateInput{
		SKUMaster:          req.SKUMaster,
		Title:              req.Title,
		Brand:              req.Brand,
		CategoryNormalized: req.CategoryNormalized,
		Variations:         toVariations(req.Variations),
	})
	if err != nil {
		h.handleServiceError(w, err, "create")
		return
	}

	writeJSON(w, http.StatusCreated, toResponse(*p))
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "não autenticado")
		return
	}

	p, err := h.svc.Get(r.Context(), claims.TenantID, r.PathValue("id"))
	if err != nil {
		h.handleServiceError(w, err, "get")
		return
	}
	writeJSON(w, http.StatusOK, toResponse(*p))
}

type listResponse struct {
	Items  []productResponse `json:"items"`
	Total  int64             `json:"total"`
	Limit  int64             `json:"limit"`
	Offset int64             `json:"offset"`
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "não autenticado")
		return
	}

	q := r.URL.Query()
	limit, _ := strconv.ParseInt(q.Get("limit"), 10, 64)
	offset, _ := strconv.ParseInt(q.Get("offset"), 10, 64)

	products, total, err := h.svc.List(r.Context(), claims.TenantID, ListInput{
		Limit:  limit,
		Offset: offset,
		Query:  q.Get("q"),
	})
	if err != nil {
		h.log.Error("product.list falhou", "error", err)
		writeError(w, http.StatusInternalServerError, "não foi possível listar produtos")
		return
	}

	items := make([]productResponse, len(products))
	for i, p := range products {
		items[i] = toResponse(p)
	}
	if limit <= 0 {
		limit = defaultListLimit
	}
	writeJSON(w, http.StatusOK, listResponse{Items: items, Total: total, Limit: limit, Offset: offset})
}

type updateRequest struct {
	Title              *string        `json:"title"`
	Brand              *string        `json:"brand"`
	CategoryNormalized *string        `json:"category_normalized"`
	Variations         []variationDTO `json:"variations,omitempty"`
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "não autenticado")
		return
	}

	var req updateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	in := UpdateInput{Title: req.Title, Brand: req.Brand, CategoryNormalized: req.CategoryNormalized}
	if req.Variations != nil {
		v := toVariations(req.Variations)
		in.Variations = &v
	}

	p, err := h.svc.Update(r.Context(), claims.TenantID, r.PathValue("id"), in)
	if err != nil {
		h.handleServiceError(w, err, "update")
		return
	}
	writeJSON(w, http.StatusOK, toResponse(*p))
}

type linkChannelRequest struct {
	ItemID string `json:"item_id"`
}

// linkChannel só registra o vínculo produto<->anúncio (ex.: item_id do
// Mercado Livre). O sync de verdade (empurrar preço/estoque) é feito pelo
// pacote internal/marketplace, registrado à parte em cmd/api/main.go — esse
// handler aqui não sabe nada sobre a API do marketplace, só grava o
// relacionamento.
func (h *Handler) linkChannel(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "não autenticado")
		return
	}

	var req linkChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	p, err := h.svc.LinkChannel(r.Context(), claims.TenantID, r.PathValue("id"), r.PathValue("marketplace"), req.ItemID)
	if err != nil {
		h.handleServiceError(w, err, "link_channel")
		return
	}
	writeJSON(w, http.StatusOK, toResponse(*p))
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "não autenticado")
		return
	}

	if err := h.svc.Delete(r.Context(), claims.TenantID, r.PathValue("id")); err != nil {
		h.handleServiceError(w, err, "delete")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleServiceError(w http.ResponseWriter, err error, op string) {
	switch {
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrSKUTaken),
		errors.Is(err, ErrSKUMasterRequired),
		errors.Is(err, ErrTitleRequired),
		errors.Is(err, ErrNoVariations),
		errors.Is(err, ErrVariationSKURequired),
		errors.Is(err, ErrDuplicateVariation),
		errors.Is(err, ErrMarketplaceRequired):
		writeError(w, http.StatusUnprocessableEntity, err.Error())
	default:
		h.log.Error("product."+op+" falhou", "error", err)
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
