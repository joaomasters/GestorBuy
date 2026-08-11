package orders

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const apiBaseURL = "https://api.mercadolibre.com"

// RawOrder é o formato (parcial — só os campos que usamos) da resposta de
// GET /orders/search. Ler os próprios pedidos autenticado é um caso de uso
// sancionado pra apps de vendedor — diferente da leitura de item de
// terceiro (ver internal/mining), não temos indício de bloqueio aqui.
type RawOrder struct {
	ID          int64     `json:"id"`
	Status      string    `json:"status"`
	DateCreated time.Time `json:"date_created"`
	TotalAmount float64   `json:"total_amount"`
	Buyer       struct {
		Nickname string `json:"nickname"`
	} `json:"buyer"`
	OrderItems []struct {
		Item struct {
			Title string `json:"title"`
		} `json:"item"`
		Quantity  int     `json:"quantity"`
		UnitPrice float64 `json:"unit_price"`
	} `json:"order_items"`
}

type searchResponse struct {
	Results []RawOrder `json:"results"`
}

type MercadoLivreClient struct {
	httpClient *http.Client
}

func NewMercadoLivreClient() *MercadoLivreClient {
	return &MercadoLivreClient{httpClient: &http.Client{Timeout: 15 * time.Second}}
}

func (c *MercadoLivreClient) SearchOrders(ctx context.Context, accessToken, sellerID string) ([]RawOrder, error) {
	q := url.Values{"seller": {sellerID}, "sort": {"date_desc"}}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiBaseURL+"/orders/search?"+q.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("orders: montar request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("orders: request de busca: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("orders: API do ML retornou %d: %s", resp.StatusCode, string(body))
	}

	var parsed searchResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("orders: decodificar resposta: %w", err)
	}
	return parsed.Results, nil
}
