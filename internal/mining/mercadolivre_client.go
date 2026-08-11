package mining

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ErrForbidden sinaliza um 403 da API do ML — na prática, hoje isso
// significa "esse item não pertence à conta autenticada" (ver nota abaixo).
// O Service traduz isso pra uma mensagem amigável (ErrNotOwnItem).
var ErrForbidden = errors.New("mining: acesso negado pela API do Mercado Livre")

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

// FetchItem busca os dados de um anúncio — título, preço, sold_quantity
// (unidades vendidas), etc.
//
// Descoberto durante a validação: o Mercado Livre bloqueia (403
// PolicyAgent) chamadas anônimas a este mesmo endpoint, mesmo vindas de
// infraestrutura "legítima" como o Railway — não é um bloqueio de IP de
// sandbox, é deliberado nos endpoints que uma ferramenta de mineração
// usaria. Por isso accessToken aqui não é mais opcional na prática: sem um
// token de uma conta conectada (ver internal/marketplace), a chamada falha.
// Mantido como parâmetro (não obrigatório na assinatura) só porque a
// requisição HTTP em si funciona igual com ou sem — quem decide se tem
// token disponível é o Service, não este client.
func (c *MercadoLivreClient) FetchItem(ctx context.Context, itemID, accessToken string) (*PublicItem, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, publicAPIBaseURL+"/items/"+itemID, nil)
	if err != nil {
		return nil, fmt.Errorf("mining: montar request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("mining: request de item público: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("mining: item %s não encontrado", itemID)
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, ErrForbidden
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mining: API do ML retornou %d: %s", resp.StatusCode, string(body))
	}

	var item PublicItem
	if err := json.Unmarshal(body, &item); err != nil {
		return nil, fmt.Errorf("mining: decodificar item: %w", err)
	}
	return &item, nil
}
