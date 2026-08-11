// Package health expõe o endpoint público de verificação de saúde do
// serviço, checando a conectividade real com o MongoDB (não só "o processo
// está de pé").
package health

import (
	"encoding/json"
	"net/http"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Handler struct {
	client *mongo.Client
}

func NewHandler(client *mongo.Client) *Handler {
	return &Handler{client: client}
}

func (h *Handler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.check)
}

type response struct {
	Status string `json:"status"`
	Mongo  string `json:"mongo"`
}

func (h *Handler) check(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if err := h.client.Ping(r.Context(), nil); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(response{Status: "unhealthy", Mongo: "unreachable"})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response{Status: "ok", Mongo: "ok"})
}
