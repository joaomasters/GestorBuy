// Package dashboard agrega pedidos (internal/orders) e custo do catálogo
// (internal/product) numa visão só — faturamento, lucro bruto, margem e
// receita diária. Não tem repository/coleção própria: só orquestra os
// outros dois pacotes, mesmo padrão que internal/marketplace já usa com
// internal/product.
package dashboard

import "time"

type DailyRevenue struct {
	Date    string  `json:"date"` // YYYY-MM-DD
	Revenue float64 `json:"revenue"`
}

// Summary é o resultado agregado de um período. "Líquido do Marketplace"
// (repasse líquido do marketplace, estilo UpSeller) fica de fora — exigiria
// a API de Faturamento/Billing do ML, que não temos conectada ainda.
type Summary struct {
	From          time.Time `json:"from"`
	To            time.Time `json:"to"`
	Revenue       float64   `json:"revenue"`
	GrossProfit   float64   `json:"gross_profit"`
	MarginPercent float64   `json:"margin_percent"`
	// UnmatchedRevenue é a fatia (aproximada, por linha de item) da receita
	// que não entrou no cálculo de lucro porque o item do pedido não está
	// vinculado a um produto do catálogo com custo cadastrado — não é uma
	// subtração exata de Revenue (que soma o total do pedido, incluindo
	// frete/desconto), é só um sinal pra UI avisar "parte disso é estimativa
	// incompleta".
	UnmatchedRevenue float64        `json:"unmatched_revenue"`
	Daily            []DailyRevenue `json:"daily"`
}
