package mining

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const publicAPIBaseURL = "https://api.mercadolibre.com"

type PublicItem struct {
	ID                string  `json:"id"`
	Title             string  `json:"title"`
	Price             float64 `json:"price"`
	SoldQuantity      int64   `json:"sold_quantity"`
	AvailableQuantity int64   `json:"available_quantity"`
	Permalink         string  `json:"permalink"`
	CategoryID        string  `json:"category_id"`
	SellerID          int64   `json:"seller_id"`
}

type MercadoLivreClient struct {
	httpClient *http.Client
}

func NewMercadoLivreClient() *MercadoLivreClient {
	return &MercadoLivreClient{httpClient: &http.Client{Timeout: 10 * time.Second}}
}

// FetchPublicItem busca os dados públicos de um anúncio — sem token, sem
// autenticação, é a mesma informação que qualquer visitante vê na página do
// produto. `sold_quantity` já vem pronto da API, poupando a gente de
// scraping ou de calcular delta manualmente pro total acumulado.
func (c *MercadoLivreClient) FetchPublicItem(ctx context.Context, itemID string) (*PublicItem, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, publicAPIBaseURL+"/items/"+itemID, nil)
	if err != nil {
		return nil, fmt.Errorf("mining: montar request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("mining: request de item público: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("mining: item %s não encontrado", itemID)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mining: API pública do ML retornou %d: %s", resp.StatusCode, string(body))
	}

	var item PublicItem
	if err := json.Unmarshal(body, &item); err != nil {
		return nil, fmt.Errorf("mining: decodificar item: %w", err)
	}
	return &item, nil
}
