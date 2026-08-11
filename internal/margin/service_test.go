package margin

import (
	"math"
	"testing"
)

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 0.001
}

func TestCalculate_PositiveMargin(t *testing.T) {
	result := Calculate(CalculateInput{
		CostPrice:             32.50,
		SalePrice:             89.90,
		MarketplaceFeePercent: 12,
		ShippingCost:          8.00,
		TaxPercent:            6,
	})

	wantFee := 89.90 * 0.12
	wantTax := 89.90 * 0.06
	wantProfit := 89.90 - 32.50 - wantFee - wantTax - 8.00

	if !almostEqual(result.MarketplaceFeeAmount, wantFee) {
		t.Errorf("MarketplaceFeeAmount = %.4f, esperado %.4f", result.MarketplaceFeeAmount, wantFee)
	}
	if !almostEqual(result.TaxAmount, wantTax) {
		t.Errorf("TaxAmount = %.4f, esperado %.4f", result.TaxAmount, wantTax)
	}
	if !almostEqual(result.NetProfit, wantProfit) {
		t.Errorf("NetProfit = %.4f, esperado %.4f", result.NetProfit, wantProfit)
	}
	if result.NetProfit <= 0 {
		t.Errorf("esperava margem positiva, veio NetProfit=%.4f", result.NetProfit)
	}
	if result.BreakEvenPrice == nil {
		t.Fatal("esperava BreakEvenPrice calculado, veio nil")
	}

	// No break-even, o lucro líquido deveria ser ~0.
	atBreakEven := Calculate(CalculateInput{
		CostPrice:             32.50,
		SalePrice:             *result.BreakEvenPrice,
		MarketplaceFeePercent: 12,
		ShippingCost:          8.00,
		TaxPercent:            6,
	})
	if !almostEqual(atBreakEven.NetProfit, 0) {
		t.Errorf("lucro no break-even deveria ser ~0, veio %.4f", atBreakEven.NetProfit)
	}
}

func TestCalculate_NegativeMargin(t *testing.T) {
	result := Calculate(CalculateInput{
		CostPrice:             80,
		SalePrice:             89.90,
		MarketplaceFeePercent: 17,
		ShippingCost:          15,
		TaxPercent:            6,
	})

	if result.NetProfit >= 0 {
		t.Errorf("esperava lucro negativo, veio NetProfit=%.4f", result.NetProfit)
	}
	if result.MarginPercent >= 0 {
		t.Errorf("esperava margem negativa, veio MarginPercent=%.4f", result.MarginPercent)
	}
}

func TestCalculate_BreakEvenImpossibleWhenFeesExceed100Percent(t *testing.T) {
	result := Calculate(CalculateInput{
		CostPrice:             10,
		SalePrice:             50,
		MarketplaceFeePercent: 60,
		TaxPercent:            45,
	})

	if result.BreakEvenPrice != nil {
		t.Errorf("esperava BreakEvenPrice nil (taxa+imposto >= 100%%), veio %.4f", *result.BreakEvenPrice)
	}
}

func TestCalculate_ZeroSalePriceDoesNotPanic(t *testing.T) {
	result := Calculate(CalculateInput{CostPrice: 10, SalePrice: 0})
	if result.MarginPercent != 0 {
		t.Errorf("MarginPercent com sale_price=0 deveria ser 0, veio %.4f", result.MarginPercent)
	}
}
