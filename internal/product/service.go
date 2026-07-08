package product

import (
	"context"
	"errors"
	"strings"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 10
	MaxPageSize     = 100
	MaxSearchLength = 120

	DefaultSort  = SortFieldID
	DefaultOrder = SortOrderAsc

	SortFieldID        = "id"
	SortFieldSKU       = "sku"
	SortFieldName      = "name"
	SortFieldPrice     = "price"
	SortFieldCreatedAt = "created_at"
	SortFieldUpdatedAt = "updated_at"

	SortOrderAsc  = "asc"
	SortOrderDesc = "desc"

	maxProductSKULength         = 64
	maxProductNameLength        = 120
	maxProductDescriptionLength = 500
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) List(ctx context.Context, input ListProductsInput) (ListProductsResult, error) {
	normalizedInput, err := normalizeListProductsInput(input)
	if err != nil {
		return ListProductsResult{}, err
	}

	return s.repository.List(ctx, normalizedInput)
}

func (s *Service) Get(ctx context.Context, id string) (Product, error) {
	return s.repository.Get(ctx, strings.TrimSpace(id))
}

func (s *Service) GetBySKU(ctx context.Context, sku string) (Product, error) {
	return s.repository.GetBySKU(ctx, normalizeSKU(sku))
}

func (s *Service) SKUExists(ctx context.Context, sku string) (SKUExistsResult, error) {
	normalizedSKU := normalizeSKU(sku)

	_, err := s.repository.GetBySKU(ctx, normalizedSKU)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return SKUExistsResult{
				SKU:    normalizedSKU,
				Exists: false,
			}, nil
		}

		return SKUExistsResult{}, err
	}

	return SKUExistsResult{
		SKU:    normalizedSKU,
		Exists: true,
	}, nil
}

func (s *Service) Create(ctx context.Context, input CreateProductInput) (Product, error) {
	normalizedInput := CreateProductInput{
		SKU:         normalizeSKU(input.SKU),
		Name:        strings.TrimSpace(input.Name),
		Description: strings.TrimSpace(input.Description),
		Price:       input.Price,
	}

	if err := validateProductInput(
		normalizedInput.SKU,
		normalizedInput.Name,
		normalizedInput.Description,
		normalizedInput.Price,
	); err != nil {
		return Product{}, err
	}

	return s.repository.Create(ctx, normalizedInput)
}

func (s *Service) Update(ctx context.Context, id string, input UpdateProductInput) (Product, error) {
	normalizedInput := UpdateProductInput{
		SKU:         normalizeSKU(input.SKU),
		Name:        strings.TrimSpace(input.Name),
		Description: strings.TrimSpace(input.Description),
		Price:       input.Price,
	}

	if err := validateProductInput(
		normalizedInput.SKU,
		normalizedInput.Name,
		normalizedInput.Description,
		normalizedInput.Price,
	); err != nil {
		return Product{}, err
	}

	return s.repository.Update(ctx, strings.TrimSpace(id), normalizedInput)
}

func (s *Service) Patch(ctx context.Context, id string, input PatchProductInput) (Product, error) {
	normalizedInput, err := normalizePatchProductInput(input)
	if err != nil {
		return Product{}, err
	}

	trimmedID := strings.TrimSpace(id)

	existingProduct, err := s.repository.Get(ctx, trimmedID)
	if err != nil {
		return Product{}, err
	}

	updateInput := UpdateProductInput{
		SKU:         existingProduct.SKU,
		Name:        existingProduct.Name,
		Description: existingProduct.Description,
		Price:       existingProduct.Price,
	}

	if normalizedInput.SKU != nil {
		updateInput.SKU = *normalizedInput.SKU
	}

	if normalizedInput.Name != nil {
		updateInput.Name = *normalizedInput.Name
	}

	if normalizedInput.Description != nil {
		updateInput.Description = *normalizedInput.Description
	}

	if normalizedInput.Price != nil {
		updateInput.Price = *normalizedInput.Price
	}

	return s.Update(ctx, trimmedID, updateInput)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repository.Delete(ctx, strings.TrimSpace(id))
}

type FieldViolation struct {
	Field   string
	Message string
}

type ValidationError struct {
	Fields []FieldViolation
}

func (e ValidationError) Error() string {
	return "validation failed"
}

func normalizeListProductsInput(input ListProductsInput) (ListProductsInput, error) {
	if input.Page == 0 {
		input.Page = DefaultPage
	}

	if input.PageSize == 0 {
		input.PageSize = DefaultPageSize
	}

	input.Search = strings.TrimSpace(input.Search)
	input.Sort = strings.ToLower(strings.TrimSpace(input.Sort))
	input.Order = strings.ToLower(strings.TrimSpace(input.Order))

	if input.Sort == "" {
		input.Sort = DefaultSort
	}

	if input.Order == "" {
		input.Order = DefaultOrder
	}

	var fields []FieldViolation

	if input.Page < 1 {
		fields = append(fields, FieldViolation{
			Field:   "page",
			Message: "page must be greater than or equal to 1",
		})
	}

	if input.PageSize < 1 {
		fields = append(fields, FieldViolation{
			Field:   "page_size",
			Message: "page_size must be greater than or equal to 1",
		})
	}

	if input.PageSize > MaxPageSize {
		fields = append(fields, FieldViolation{
			Field:   "page_size",
			Message: "page_size must be less than or equal to 100",
		})
	}

	if input.MinPrice != nil && *input.MinPrice < 0 {
		fields = append(fields, FieldViolation{
			Field:   "min_price",
			Message: "min_price must be greater than or equal to zero",
		})
	}

	if input.MaxPrice != nil && *input.MaxPrice < 0 {
		fields = append(fields, FieldViolation{
			Field:   "max_price",
			Message: "max_price must be greater than or equal to zero",
		})
	}

	if input.MinPrice != nil &&
		input.MaxPrice != nil &&
		*input.MinPrice > *input.MaxPrice {
		fields = append(fields, FieldViolation{
			Field:   "price_range",
			Message: "min_price must be less than or equal to max_price",
		})
	}

	if input.CreatedFrom != nil &&
		input.CreatedTo != nil &&
		input.CreatedFrom.After(*input.CreatedTo) {
		fields = append(fields, FieldViolation{
			Field:   "created_range",
			Message: "created_from must be less than or equal to created_to",
		})
	}

	if len(input.Search) > MaxSearchLength {
		fields = append(fields, FieldViolation{
			Field:   "search",
			Message: "search must be less than or equal to 120 characters",
		})
	}

	if !IsSupportedSortField(input.Sort) {
		fields = append(fields, FieldViolation{
			Field:   "sort",
			Message: "sort must be one of: id, name, price, created_at, updated_at",
		})
	}

	if !IsSupportedSortOrder(input.Order) {
		fields = append(fields, FieldViolation{
			Field:   "order",
			Message: "order must be one of: asc, desc",
		})
	}

	if len(fields) > 0 {
		return ListProductsInput{}, ValidationError{
			Fields: fields,
		}
	}

	return input, nil
}

func normalizePatchProductInput(input PatchProductInput) (PatchProductInput, error) {
	var fields []FieldViolation

	normalizedInput := PatchProductInput{}
	hasAnyField := false

	if input.SKU != nil {
		hasAnyField = true

		value := normalizeSKU(*input.SKU)
		normalizedInput.SKU = &value

		if value == "" {
			fields = append(fields, FieldViolation{
				Field:   "sku",
				Message: "sku is required",
			})
		}

		if len(value) > maxProductSKULength {
			fields = append(fields, FieldViolation{
				Field:   "sku",
				Message: "sku must be less than or equal to 64 characters",
			})
		}
	}

	if input.Name != nil {
		hasAnyField = true

		value := strings.TrimSpace(*input.Name)
		normalizedInput.Name = &value

		if value == "" {
			fields = append(fields, FieldViolation{
				Field:   "name",
				Message: "name is required",
			})
		}

		if len(value) > maxProductNameLength {
			fields = append(fields, FieldViolation{
				Field:   "name",
				Message: "name must be less than or equal to 120 characters",
			})
		}
	}

	if input.Description != nil {
		hasAnyField = true

		value := strings.TrimSpace(*input.Description)
		normalizedInput.Description = &value

		if len(value) > maxProductDescriptionLength {
			fields = append(fields, FieldViolation{
				Field:   "description",
				Message: "description must be less than or equal to 500 characters",
			})
		}
	}

	if input.Price != nil {
		hasAnyField = true

		value := *input.Price
		normalizedInput.Price = &value

		if value < 0 {
			fields = append(fields, FieldViolation{
				Field:   "price",
				Message: "price must be greater than or equal to zero",
			})
		}
	}

	if !hasAnyField {
		fields = append(fields, FieldViolation{
			Field:   "body",
			Message: "at least one field must be provided",
		})
	}

	if len(fields) > 0 {
		return PatchProductInput{}, ValidationError{
			Fields: fields,
		}
	}

	return normalizedInput, nil
}

func IsSupportedSortField(value string) bool {
	switch value {
	case SortFieldID,
		SortFieldSKU,
		SortFieldName,
		SortFieldPrice,
		SortFieldCreatedAt,
		SortFieldUpdatedAt:
		return true
	default:
		return false
	}
}

func IsSupportedSortOrder(value string) bool {
	switch value {
	case SortOrderAsc, SortOrderDesc:
		return true
	default:
		return false
	}
}

func validateProductInput(sku string, name string, description string, price float64) error {
	var fields []FieldViolation

	if sku == "" {
		fields = append(fields, FieldViolation{
			Field:   "sku",
			Message: "sku is required",
		})
	}

	if len(sku) > maxProductSKULength {
		fields = append(fields, FieldViolation{
			Field:   "sku",
			Message: "sku must be less than or equal to 64 characters",
		})
	}

	if name == "" {
		fields = append(fields, FieldViolation{
			Field:   "name",
			Message: "name is required",
		})
	}

	if len(name) > maxProductNameLength {
		fields = append(fields, FieldViolation{
			Field:   "name",
			Message: "name must be less than or equal to 120 characters",
		})
	}

	if len(description) > maxProductDescriptionLength {
		fields = append(fields, FieldViolation{
			Field:   "description",
			Message: "description must be less than or equal to 500 characters",
		})
	}

	if price < 0 {
		fields = append(fields, FieldViolation{
			Field:   "price",
			Message: "price must be greater than or equal to zero",
		})
	}

	if len(fields) > 0 {
		return ValidationError{
			Fields: fields,
		}
	}

	return nil
}

func normalizeSKU(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}
