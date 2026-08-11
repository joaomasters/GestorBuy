package product

import (
	"context"
	"errors"
	"testing"
)

// Testes das validações de negócio do service — não tocam no banco (repo
// fica nil de propósito; qualquer caminho que chegasse a usá-lo faria o
// teste panicar, o que também serve como sinal de que a validação não
// barrou a entrada inválida antes da hora).

func TestCreate_RequiresSKUMaster(t *testing.T) {
	svc := NewService(nil)
	_, err := svc.Create(context.Background(), "tenant-1", CreateInput{
		Title:      "Produto",
		Variations: []Variation{{VariationSKU: "SKU-1"}},
	})
	if !errors.Is(err, ErrSKUMasterRequired) {
		t.Fatalf("esperava ErrSKUMasterRequired, veio %v", err)
	}
}

func TestCreate_RequiresTitle(t *testing.T) {
	svc := NewService(nil)
	_, err := svc.Create(context.Background(), "tenant-1", CreateInput{
		SKUMaster:  "SKU-MASTER",
		Variations: []Variation{{VariationSKU: "SKU-1"}},
	})
	if !errors.Is(err, ErrTitleRequired) {
		t.Fatalf("esperava ErrTitleRequired, veio %v", err)
	}
}

func TestCreate_RequiresAtLeastOneVariation(t *testing.T) {
	svc := NewService(nil)
	_, err := svc.Create(context.Background(), "tenant-1", CreateInput{
		SKUMaster: "SKU-MASTER",
		Title:     "Produto",
	})
	if !errors.Is(err, ErrNoVariations) {
		t.Fatalf("esperava ErrNoVariations, veio %v", err)
	}
}

func TestCreate_RejectsVariationWithoutSKU(t *testing.T) {
	svc := NewService(nil)
	_, err := svc.Create(context.Background(), "tenant-1", CreateInput{
		SKUMaster:  "SKU-MASTER",
		Title:      "Produto",
		Variations: []Variation{{VariationSKU: ""}},
	})
	if !errors.Is(err, ErrVariationSKURequired) {
		t.Fatalf("esperava ErrVariationSKURequired, veio %v", err)
	}
}

func TestCreate_RejectsDuplicateVariationSKU(t *testing.T) {
	svc := NewService(nil)
	_, err := svc.Create(context.Background(), "tenant-1", CreateInput{
		SKUMaster: "SKU-MASTER",
		Title:     "Produto",
		Variations: []Variation{
			{VariationSKU: "SKU-1"},
			{VariationSKU: "SKU-1"},
		},
	})
	if !errors.Is(err, ErrDuplicateVariation) {
		t.Fatalf("esperava ErrDuplicateVariation, veio %v", err)
	}
}

func TestUpdate_RejectsEmptyTitle(t *testing.T) {
	svc := NewService(nil)
	empty := ""
	_, err := svc.Update(context.Background(), "tenant-1", "product-1", UpdateInput{Title: &empty})
	if !errors.Is(err, ErrTitleRequired) {
		t.Fatalf("esperava ErrTitleRequired, veio %v", err)
	}
}

func TestUpdate_RejectsInvalidVariations(t *testing.T) {
	svc := NewService(nil)
	variations := []Variation{{VariationSKU: "SKU-1"}, {VariationSKU: "SKU-1"}}
	_, err := svc.Update(context.Background(), "tenant-1", "product-1", UpdateInput{Variations: &variations})
	if !errors.Is(err, ErrDuplicateVariation) {
		t.Fatalf("esperava ErrDuplicateVariation, veio %v", err)
	}
}
