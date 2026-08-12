package dashboard

import (
	"math"
	"testing"
	"time"

	"github.com/gestorbuy/api/internal/orders"
)

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 0.001
}

func day(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestSummarize_ItemWithLinkedCost(t *testing.T) {
	orderList := []orders.Order{
		{
			TotalAmount: 100,
			DateCreated: day("2026-08-01"),
			Items: []orders.OrderItem{
				{ExternalItemID: "MLB1", UnitPrice: 100, Quantity: 1},
			},
		},
	}
	costLookup := func(itemID string) (float64, bool) {
		if itemID == "MLB1" {
			return 60, true // custo 60, vendeu por 100 -> lucro 40
		}
		return 0, false
	}

	got := Summarize(day("2026-08-01"), day("2026-08-01"), orderList, costLookup)

	if !almostEqual(got.Revenue, 100) {
		t.Errorf("Revenue = %.2f, esperado 100", got.Revenue)
	}
	if !almostEqual(got.GrossProfit, 40) {
		t.Errorf("GrossProfit = %.2f, esperado 40", got.GrossProfit)
	}
	if !almostEqual(got.MarginPercent, 40) {
		t.Errorf("MarginPercent = %.2f, esperado 40", got.MarginPercent)
	}
	if !almostEqual(got.UnmatchedRevenue, 0) {
		t.Errorf("UnmatchedRevenue = %.2f, esperado 0", got.UnmatchedRevenue)
	}
}

func TestSummarize_ItemWithoutLinkedCost(t *testing.T) {
	orderList := []orders.Order{
		{
			TotalAmount: 100,
			DateCreated: day("2026-08-01"),
			Items: []orders.OrderItem{
				{ExternalItemID: "MLB-sem-vinculo", UnitPrice: 100, Quantity: 1},
			},
		},
	}
	costLookup := func(itemID string) (float64, bool) { return 0, false }

	got := Summarize(day("2026-08-01"), day("2026-08-01"), orderList, costLookup)

	if !almostEqual(got.Revenue, 100) {
		t.Errorf("Revenue = %.2f, esperado 100", got.Revenue)
	}
	if !almostEqual(got.GrossProfit, 0) {
		t.Errorf("GrossProfit = %.2f, esperado 0 (sem custo vinculado)", got.GrossProfit)
	}
	if !almostEqual(got.UnmatchedRevenue, 100) {
		t.Errorf("UnmatchedRevenue = %.2f, esperado 100", got.UnmatchedRevenue)
	}
}

func TestSummarize_EmptyPeriod(t *testing.T) {
	got := Summarize(day("2026-08-01"), day("2026-08-31"), nil, func(string) (float64, bool) { return 0, false })

	if got.Revenue != 0 || got.GrossProfit != 0 || got.MarginPercent != 0 {
		t.Errorf("esperava tudo zerado num período sem pedido, veio %+v", got)
	}
	if len(got.Daily) != 0 {
		t.Errorf("esperava Daily vazio, veio %d entradas", len(got.Daily))
	}
}

func TestSummarize_GroupsByDayAndSortsChronologically(t *testing.T) {
	orderList := []orders.Order{
		{TotalAmount: 50, DateCreated: day("2026-08-02")},
		{TotalAmount: 30, DateCreated: day("2026-08-01")},
		{TotalAmount: 20, DateCreated: day("2026-08-01")},
	}

	got := Summarize(day("2026-08-01"), day("2026-08-02"), orderList, func(string) (float64, bool) { return 0, false })

	if len(got.Daily) != 2 {
		t.Fatalf("esperava 2 dias agrupados, veio %d", len(got.Daily))
	}
	if got.Daily[0].Date != "2026-08-01" || !almostEqual(got.Daily[0].Revenue, 50) {
		t.Errorf("dia 1 = %+v, esperado 2026-08-01 com 50", got.Daily[0])
	}
	if got.Daily[1].Date != "2026-08-02" || !almostEqual(got.Daily[1].Revenue, 50) {
		t.Errorf("dia 2 = %+v, esperado 2026-08-02 com 50", got.Daily[1])
	}
}

func TestSummarize_MixOfMatchedAndUnmatchedItems(t *testing.T) {
	orderList := []orders.Order{
		{
			TotalAmount: 250,
			DateCreated: day("2026-08-01"),
			Items: []orders.OrderItem{
				{ExternalItemID: "MLB-com-custo", UnitPrice: 150, Quantity: 1},
				{ExternalItemID: "MLB-sem-custo", UnitPrice: 100, Quantity: 1},
			},
		},
	}
	costLookup := func(itemID string) (float64, bool) {
		if itemID == "MLB-com-custo" {
			return 90, true
		}
		return 0, false
	}

	got := Summarize(day("2026-08-01"), day("2026-08-01"), orderList, costLookup)

	if !almostEqual(got.GrossProfit, 60) { // 150 - 90
		t.Errorf("GrossProfit = %.2f, esperado 60", got.GrossProfit)
	}
	if !almostEqual(got.UnmatchedRevenue, 100) {
		t.Errorf("UnmatchedRevenue = %.2f, esperado 100", got.UnmatchedRevenue)
	}
}
