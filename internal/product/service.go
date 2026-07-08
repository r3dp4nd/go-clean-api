package product

import (
	"context"
	"strings"
)

const (
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

func (s *Service) List(ctx context.Context) ([]Product, error) {
	return s.repository.List(ctx)
}

func (s *Service) Get(ctx context.Context, id string) (Product, error) {
	return s.repository.Get(ctx, strings.TrimSpace(id))
}

func (s *Service) Create(ctx context.Context, input CreateProductInput) (Product, error) {
	normalizedInput := CreateProductInput{
		Name:        strings.TrimSpace(input.Name),
		Description: strings.TrimSpace(input.Description),
		Price:       input.Price,
	}

	if err := validateProductInput(
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
		Name:        strings.TrimSpace(input.Name),
		Description: strings.TrimSpace(input.Description),
		Price:       input.Price,
	}

	if err := validateProductInput(
		normalizedInput.Name,
		normalizedInput.Description,
		normalizedInput.Price,
	); err != nil {
		return Product{}, err
	}

	return s.repository.Update(ctx, strings.TrimSpace(id), normalizedInput)
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

func validateProductInput(name string, description string, price float64) error {
	var fields []FieldViolation

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
