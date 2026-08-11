package margin

import (
	"encoding/json"
	"net/http"

	"github.com/gestorbuy/api/internal/auth"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Routes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler) {
	mux.Handle("POST /margin/calculate", authMiddleware(http.HandlerFunc(h.calculate)))
}

type calculateRequest struct {
	CostPrice             float64 `json:"cost_price"`
	SalePrice             float64 `json:"sale_price"`
	MarketplaceFeePercent float64 `json:"marketplace_fee_percent"`
	ShippingCost          float64 `json:"shipping_cost"`
	TaxPercent            float64 `json:"tax_percent"`
}

type calculateResponse struct {
	MarketplaceFeeAmount float64  `json:"marketplace_fee_amount"`
	TaxAmount            float64  `json:"tax_amount"`
	NetProfit            float64  `json:"net_profit"`
	MarginPercent        float64  `json:"margin_percent"`
	BreakEvenPrice       *float64 `json:"break_even_price,omitempty"`
}

// calculate exige autenticação só por consistência com o resto da API (fica
// registrado nos logs de acesso quem usou), não porque o cálculo em si
// dependa de tenant_id — é stateless, sem persistência.
func (h *Handler) calculate(w http.ResponseWriter, r *http.Request) {
	if _, ok := auth.ClaimsFromContext(r.Context()); !ok {
		writeError(w, http.StatusUnauthorized, "não autenticado")
		return
	}

	var req calculateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	result := Calculate(CalculateInput{
		CostPrice:             req.CostPrice,
		SalePrice:             req.SalePrice,
		MarketplaceFeePercent: req.MarketplaceFeePercent,
		ShippingCost:          req.ShippingCost,
		TaxPercent:            req.TaxPercent,
	})

	writeJSON(w, http.StatusOK, calculateResponse{
		MarketplaceFeeAmount: result.MarketplaceFeeAmount,
		TaxAmount:            result.TaxAmount,
		NetProfit:            result.NetProfit,
		MarginPercent:        result.MarginPercent,
		BreakEvenPrice:       result.BreakEvenPrice,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
