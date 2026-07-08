package product

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

var errRepositoryFailure = errors.New("repository failure")

type fakeRepository struct {
	listFn   func(ctx context.Context) ([]Product, error)
	getFn    func(ctx context.Context, id string) (Product, error)
	createFn func(ctx context.Context, input CreateProductInput) (Product, error)
	updateFn func(ctx context.Context, id string, input UpdateProductInput) (Product, error)
	deleteFn func(ctx context.Context, id string) error

	listCalls   int
	getCalls    int
	createCalls int
	updateCalls int
	deleteCalls int
}

func (f *fakeRepository) List(ctx context.Context) ([]Product, error) {
	f.listCalls++

	if f.listFn != nil {
		return f.listFn(ctx)
	}

	return nil, nil
}

func (f *fakeRepository) Get(ctx context.Context, id string) (Product, error) {
	f.getCalls++

	if f.getFn != nil {
		return f.getFn(ctx, id)
	}

	return Product{}, nil
}

func (f *fakeRepository) Create(ctx context.Context, input CreateProductInput) (Product, error) {
	f.createCalls++

	if f.createFn != nil {
		return f.createFn(ctx, input)
	}

	return Product{}, nil
}

func (f *fakeRepository) Update(ctx context.Context, id string, input UpdateProductInput) (Product, error) {
	f.updateCalls++

	if f.updateFn != nil {
		return f.updateFn(ctx, id, input)
	}

	return Product{}, nil
}

func (f *fakeRepository) Delete(ctx context.Context, id string) error {
	f.deleteCalls++

	if f.deleteFn != nil {
		return f.deleteFn(ctx, id)
	}

	return nil
}

func TestServiceListProducts(t *testing.T) {
	expectedProducts := []Product{
		{
			ID:    "1",
			Name:  "Laptop",
			Price: 3500,
		},
		{
			ID:    "2",
			Name:  "Mouse",
			Price: 120,
		},
	}

	repository := &fakeRepository{
		listFn: func(ctx context.Context) ([]Product, error) {
			return expectedProducts, nil
		},
	}

	service := NewService(repository)

	items, err := service.List(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repository.listCalls != 1 {
		t.Fatalf("expected List to be called once, got %d", repository.listCalls)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 products, got %d", len(items))
	}

	if items[0].ID != "1" {
		t.Fatalf("expected first product ID %q, got %q", "1", items[0].ID)
	}
}

func TestServiceListProductsRepositoryError(t *testing.T) {
	repository := &fakeRepository{
		listFn: func(ctx context.Context) ([]Product, error) {
			return nil, errRepositoryFailure
		},
	}

	service := NewService(repository)

	_, err := service.List(context.Background())
	if !errors.Is(err, errRepositoryFailure) {
		t.Fatalf("expected repository error, got %v", err)
	}
}

func TestServiceGetProductTrimsID(t *testing.T) {
	repository := &fakeRepository{
		getFn: func(ctx context.Context, id string) (Product, error) {
			if id != "123" {
				t.Fatalf("expected trimmed id %q, got %q", "123", id)
			}

			return Product{
				ID:    id,
				Name:  "Laptop",
				Price: 3500,
			}, nil
		},
	}

	service := NewService(repository)

	item, err := service.Get(context.Background(), "  123  ")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repository.getCalls != 1 {
		t.Fatalf("expected Get to be called once, got %d", repository.getCalls)
	}

	if item.ID != "123" {
		t.Fatalf("expected ID %q, got %q", "123", item.ID)
	}
}

func TestServiceGetProductNotFound(t *testing.T) {
	repository := &fakeRepository{
		getFn: func(ctx context.Context, id string) (Product, error) {
			return Product{}, ErrNotFound
		},
	}

	service := NewService(repository)

	_, err := service.Get(context.Background(), "999")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestServiceCreateProductNormalizesInput(t *testing.T) {
	now := time.Now().UTC()

	repository := &fakeRepository{
		createFn: func(ctx context.Context, input CreateProductInput) (Product, error) {
			if input.Name != "Laptop" {
				t.Fatalf("expected trimmed name %q, got %q", "Laptop", input.Name)
			}

			if input.Description != "Laptop para desarrollo backend" {
				t.Fatalf("expected trimmed description %q, got %q", "Laptop para desarrollo backend", input.Description)
			}

			if input.Price != 3500 {
				t.Fatalf("expected price %v, got %v", 3500.0, input.Price)
			}

			return Product{
				ID:          "1",
				Name:        input.Name,
				Description: input.Description,
				Price:       input.Price,
				CreatedAt:   now,
				UpdatedAt:   now,
			}, nil
		},
	}

	service := NewService(repository)

	item, err := service.Create(context.Background(), CreateProductInput{
		Name:        "  Laptop  ",
		Description: "  Laptop para desarrollo backend  ",
		Price:       3500,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repository.createCalls != 1 {
		t.Fatalf("expected Create to be called once, got %d", repository.createCalls)
	}

	if item.Name != "Laptop" {
		t.Fatalf("expected product name %q, got %q", "Laptop", item.Name)
	}
}

func TestServiceCreateProductValidationErrorDoesNotCallRepository(t *testing.T) {
	repository := &fakeRepository{
		createFn: func(ctx context.Context, input CreateProductInput) (Product, error) {
			t.Fatal("repository Create should not be called when validation fails")
			return Product{}, nil
		},
	}

	service := NewService(repository)

	_, err := service.Create(context.Background(), CreateProductInput{
		Name:        "",
		Description: "Producto inválido",
		Price:       -10,
	})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	var validationErr ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}

	if repository.createCalls != 0 {
		t.Fatalf("expected Create not to be called, got %d calls", repository.createCalls)
	}

	if len(validationErr.Fields) != 2 {
		t.Fatalf("expected 2 field violations, got %d", len(validationErr.Fields))
	}

	expectedFields := map[string]string{
		"name":  "name is required",
		"price": "price must be greater than or equal to zero",
	}

	for _, field := range validationErr.Fields {
		expectedMessage, ok := expectedFields[field.Field]
		if !ok {
			t.Fatalf("unexpected field violation: %s", field.Field)
		}

		if field.Message != expectedMessage {
			t.Fatalf("expected message %q for field %q, got %q", expectedMessage, field.Field, field.Message)
		}
	}
}

func TestServiceCreateProductRepositoryError(t *testing.T) {
	repository := &fakeRepository{
		createFn: func(ctx context.Context, input CreateProductInput) (Product, error) {
			return Product{}, errRepositoryFailure
		},
	}

	service := NewService(repository)

	_, err := service.Create(context.Background(), CreateProductInput{
		Name:        "Laptop",
		Description: "Laptop para desarrollo backend",
		Price:       3500,
	})
	if !errors.Is(err, errRepositoryFailure) {
		t.Fatalf("expected repository error, got %v", err)
	}
}

func TestServiceUpdateProductNormalizesInputAndID(t *testing.T) {
	now := time.Now().UTC()

	repository := &fakeRepository{
		updateFn: func(ctx context.Context, id string, input UpdateProductInput) (Product, error) {
			if id != "123" {
				t.Fatalf("expected trimmed id %q, got %q", "123", id)
			}

			if input.Name != "Laptop Pro" {
				t.Fatalf("expected trimmed name %q, got %q", "Laptop Pro", input.Name)
			}

			if input.Description != "Laptop para Go" {
				t.Fatalf("expected trimmed description %q, got %q", "Laptop para Go", input.Description)
			}

			return Product{
				ID:          id,
				Name:        input.Name,
				Description: input.Description,
				Price:       input.Price,
				CreatedAt:   now,
				UpdatedAt:   now,
			}, nil
		},
	}

	service := NewService(repository)

	item, err := service.Update(context.Background(), "  123  ", UpdateProductInput{
		Name:        "  Laptop Pro  ",
		Description: "  Laptop para Go  ",
		Price:       4200,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repository.updateCalls != 1 {
		t.Fatalf("expected Update to be called once, got %d", repository.updateCalls)
	}

	if item.ID != "123" {
		t.Fatalf("expected product ID %q, got %q", "123", item.ID)
	}
}

func TestServiceUpdateProductValidationErrorDoesNotCallRepository(t *testing.T) {
	repository := &fakeRepository{
		updateFn: func(ctx context.Context, id string, input UpdateProductInput) (Product, error) {
			t.Fatal("repository Update should not be called when validation fails")
			return Product{}, nil
		},
	}

	service := NewService(repository)

	_, err := service.Update(context.Background(), "123", UpdateProductInput{
		Name:        "",
		Description: "Producto inválido",
		Price:       -1,
	})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	var validationErr ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}

	if repository.updateCalls != 0 {
		t.Fatalf("expected Update not to be called, got %d calls", repository.updateCalls)
	}
}

func TestServiceUpdateProductNotFound(t *testing.T) {
	repository := &fakeRepository{
		updateFn: func(ctx context.Context, id string, input UpdateProductInput) (Product, error) {
			return Product{}, ErrNotFound
		},
	}

	service := NewService(repository)

	_, err := service.Update(context.Background(), "999", UpdateProductInput{
		Name:        "Laptop",
		Description: "Laptop para backend",
		Price:       3500,
	})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestServiceDeleteProductTrimsID(t *testing.T) {
	repository := &fakeRepository{
		deleteFn: func(ctx context.Context, id string) error {
			if id != "123" {
				t.Fatalf("expected trimmed id %q, got %q", "123", id)
			}

			return nil
		},
	}

	service := NewService(repository)

	if err := service.Delete(context.Background(), "  123  "); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repository.deleteCalls != 1 {
		t.Fatalf("expected Delete to be called once, got %d", repository.deleteCalls)
	}
}

func TestServiceDeleteProductNotFound(t *testing.T) {
	repository := &fakeRepository{
		deleteFn: func(ctx context.Context, id string) error {
			return ErrNotFound
		},
	}

	service := NewService(repository)

	err := service.Delete(context.Background(), "999")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestServiceCreateProductMaxLengthValidation(t *testing.T) {
	repository := &fakeRepository{
		createFn: func(ctx context.Context, input CreateProductInput) (Product, error) {
			t.Fatal("repository Create should not be called when validation fails")
			return Product{}, nil
		},
	}

	service := NewService(repository)

	longName := strings.Repeat("a", maxProductNameLength+1)
	longDescription := strings.Repeat("b", maxProductDescriptionLength+1)

	_, err := service.Create(context.Background(), CreateProductInput{
		Name:        longName,
		Description: longDescription,
		Price:       100,
	})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	var validationErr ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}

	if repository.createCalls != 0 {
		t.Fatalf("expected Create not to be called, got %d calls", repository.createCalls)
	}

	if len(validationErr.Fields) != 2 {
		t.Fatalf("expected 2 field violations, got %d", len(validationErr.Fields))
	}
}
