package product

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	ErrSKUMasterRequired    = errors.New("product: sku_master é obrigatório")
	ErrTitleRequired        = errors.New("product: title é obrigatório")
	ErrNoVariations         = errors.New("product: produto precisa de pelo menos 1 variação")
	ErrDuplicateVariation   = errors.New("product: variation_sku duplicado dentro do produto")
	ErrVariationSKURequired = errors.New("product: toda variação precisa de variation_sku")
	ErrNotFound             = errors.New("product: produto não encontrado")
)

const (
	defaultListLimit = 20
	maxListLimit     = 100
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CreateInput chega já desserializado do handler — validação de negócio
// (distinta da validação de shape JSON) acontece toda aqui, antes de
// qualquer escrita no banco.
type CreateInput struct {
	SKUMaster          string
	Title              string
	Brand              string
	CategoryNormalized string
	Variations         []Variation
}

func validateVariations(variations []Variation) error {
	if len(variations) == 0 {
		return ErrNoVariations
	}
	seen := make(map[string]struct{}, len(variations))
	for _, v := range variations {
		if v.VariationSKU == "" {
			return ErrVariationSKURequired
		}
		if _, dup := seen[v.VariationSKU]; dup {
			return fmt.Errorf("%w: %s", ErrDuplicateVariation, v.VariationSKU)
		}
		seen[v.VariationSKU] = struct{}{}
	}
	return nil
}

func (s *Service) Create(ctx context.Context, tenantID string, in CreateInput) (*Product, error) {
	if in.SKUMaster == "" {
		return nil, ErrSKUMasterRequired
	}
	if in.Title == "" {
		return nil, ErrTitleRequired
	}
	if err := validateVariations(in.Variations); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	p := Product{
		ID:                 bson.NewObjectID().Hex(),
		TenantID:           tenantID,
		SKUMaster:          in.SKUMaster,
		Title:              in.Title,
		Brand:              in.Brand,
		CategoryNormalized: in.CategoryNormalized,
		Variations:         in.Variations,
		StockStrategy:      StockStrategyShared,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Service) Get(ctx context.Context, tenantID, id string) (*Product, error) {
	p, err := s.repo.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrNotFound
	}
	return p, nil
}

type ListInput struct {
	Limit  int64
	Offset int64
	Query  string
}

func (s *Service) List(ctx context.Context, tenantID string, in ListInput) ([]Product, int64, error) {
	limit := in.Limit
	if limit <= 0 {
		limit = defaultListLimit
	}
	if limit > maxListLimit {
		limit = maxListLimit
	}
	offset := in.Offset
	if offset < 0 {
		offset = 0
	}

	return s.repo.List(ctx, tenantID, ListOptions{Limit: limit, Offset: offset, Query: in.Query})
}

// UpdateInput usa ponteiros pra distinguir "campo não enviado" de "campo
// enviado como vazio" — só o que vier preenchido entra no $set.
type UpdateInput struct {
	Title              *string
	Brand              *string
	CategoryNormalized *string
	Variations         *[]Variation
}

func (s *Service) Update(ctx context.Context, tenantID, id string, in UpdateInput) (*Product, error) {
	set := bson.M{}
	if in.Title != nil {
		if *in.Title == "" {
			return nil, ErrTitleRequired
		}
		set["title"] = *in.Title
	}
	if in.Brand != nil {
		set["brand"] = *in.Brand
	}
	if in.CategoryNormalized != nil {
		set["category_normalized"] = *in.CategoryNormalized
	}
	if in.Variations != nil {
		if err := validateVariations(*in.Variations); err != nil {
			return nil, err
		}
		set["variations"] = *in.Variations
	}

	p, err := s.repo.Update(ctx, tenantID, id, set)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrNotFound
	}
	return p, nil
}

func (s *Service) Delete(ctx context.Context, tenantID, id string) error {
	deleted, err := s.repo.Delete(ctx, tenantID, id)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrNotFound
	}
	return nil
}
