// Package mercadolivre é um wrapper fino sobre a API REST do Mercado
// Livre — sem SDK, `net/http` puro (mesma filosofia de dependência mínima
// do resto do projeto). Cobre só o que o Hub precisa nesta entrega: trocar
// code por token, refresh, e atualizar preço/estoque de um item.
package mercadolivre

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	authBaseURL = "https://auth.mercadolivre.com.br/authorization"
	apiBaseURL  = "https://api.mercadolibre.com"
)

type Client struct {
	clientID     string
	clientSecret string
	redirectURI  string
	httpClient   *http.Client
}

func NewClient(clientID, clientSecret, redirectURI string) *Client {
	return &Client{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
		httpClient:   &http.Client{Timeout: 15 * time.Second},
	}
}

// AuthorizationURL monta a URL pra onde o navegador do usuário precisa ir
// pra autorizar o app. `state` é opaco pro Mercado Livre — devolvido como
// está no callback, é onde carregamos o tenant_id assinado (ver
// internal/marketplace.Service.ConnectURL).
func (c *Client) AuthorizationURL(state string) string {
	q := url.Values{
		"response_type": {"code"},
		"client_id":     {c.clientID},
		"redirect_uri":  {c.redirectURI},
		"state":         {state},
	}
	return authBaseURL + "?" + q.Encode()
}

// TokenResponse é o corpo de resposta comum tanto do exchange de code
// quanto do refresh — o Mercado Livre usa o mesmo shape pros dois.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"` // segundos
	RefreshToken string `json:"refresh_token"`
	UserID       int64  `json:"user_id"`
}

func (c *Client) ExchangeCode(ctx context.Context, code string) (*TokenResponse, error) {
	return c.requestToken(ctx, url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {c.clientID},
		"client_secret": {c.clientSecret},
		"code":          {code},
		"redirect_uri":  {c.redirectURI},
	})
}

func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	return c.requestToken(ctx, url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {c.clientID},
		"client_secret": {c.clientSecret},
		"refresh_token": {refreshToken},
	})
}

func (c *Client) requestToken(ctx context.Context, form url.Values) (*TokenResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiBaseURL+"/oauth/token", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("mercadolivre: montar request de token: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("mercadolivre: request de token: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mercadolivre: token endpoint retornou %d: %s", resp.StatusCode, string(body))
	}

	var tr TokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return nil, fmt.Errorf("mercadolivre: decodificar resposta de token: %w", err)
	}
	return &tr, nil
}

// UpdateItemPriceStock empurra preço/estoque pro anúncio já existente no
// Mercado Livre. Exige um access_token válido — quem garante isso é
// internal/marketplace.Service.EnsureValidToken antes de chamar aqui.
func (c *Client) UpdateItemPriceStock(ctx context.Context, accessToken, itemID string, price float64, stock int) error {
	payload, err := json.Marshal(map[string]any{
		"price":              price,
		"available_quantity": stock,
	})
	if err != nil {
		return fmt.Errorf("mercadolivre: montar payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, apiBaseURL+"/items/"+itemID, strings.NewReader(string(payload)))
	if err != nil {
		return fmt.Errorf("mercadolivre: montar request de update: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("mercadolivre: request de update: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("mercadolivre: update de item retornou %d: %s", resp.StatusCode, string(body))
	}
	return nil
}
