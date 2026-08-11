// Package margin implementa a calculadora de margem real (estilo Avant
// Pro) — avalia taxa de marketplace, frete e imposto antes de publicar ou
// ajustar o preço de um produto, pra garantir margem positiva.
//
// Deliberadamente sem persistência nem dependência de nenhuma API externa:
// é cálculo puro a partir do que o usuário informa. Depois da descoberta de
// que o Mercado Livre bloqueia vários endpoints públicos/autenticados (ver
// internal/mining), não arriscamos depender de mais uma chamada externa
// pra uma funcionalidade que não precisa disso.
package margin

type CalculateInput struct {
	CostPrice             float64
	SalePrice             float64
	MarketplaceFeePercent float64
	ShippingCost          float64
	TaxPercent            float64
}

type CalculateResult struct {
	MarketplaceFeeAmount float64
	TaxAmount            float64
	NetProfit            float64
	MarginPercent        float64
	// BreakEvenPrice é nil quando taxa+imposto somam 100% ou mais — não
	// existe preço de venda que cubra os custos nesse cenário.
	BreakEvenPrice *float64
}

// Calculate é uma função pura — sem I/O, sem estado — por isso testável
// exaustivamente sem mocks nem banco.
func Calculate(in CalculateInput) CalculateResult {
	feeAmount := in.SalePrice * in.MarketplaceFeePercent / 100
	taxAmount := in.SalePrice * in.TaxPercent / 100
	netProfit := in.SalePrice - in.CostPrice - feeAmount - taxAmount - in.ShippingCost

	var marginPercent float64
	if in.SalePrice != 0 {
		marginPercent = netProfit / in.SalePrice * 100
	}

	result := CalculateResult{
		MarketplaceFeeAmount: feeAmount,
		TaxAmount:            taxAmount,
		NetProfit:            netProfit,
		MarginPercent:        marginPercent,
	}

	// break_even: sale_price * (1 - (fee%+tax%)/100) = cost_price + shipping_cost
	denominator := 1 - (in.MarketplaceFeePercent+in.TaxPercent)/100
	if denominator > 0 {
		breakEven := (in.CostPrice + in.ShippingCost) / denominator
		result.BreakEvenPrice = &breakEven
	}

	return result
}
