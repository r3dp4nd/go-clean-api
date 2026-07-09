package product

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/r3dp4nd/go-clean-api/internal/audit"
)

var errRepositoryFailure = errors.New("repository failure")

type fakeRepository struct {
	listFn        func(ctx context.Context, input ListProductsInput) (ListProductsResult, error)
	listDeletedFn func(ctx context.Context, input ListProductsInput) (ListProductsResult, error)
	getFn         func(ctx context.Context, id string) (Product, error)
	getDeletedFn  func(ctx context.Context, id string) (Product, error)
	getBySKUFn    func(ctx context.Context, sku string) (Product, error)
	createFn      func(ctx context.Context, input CreateProductInput) (Product, error)
	updateFn      func(ctx context.Context, id string, input UpdateProductInput) (Product, error)
	deleteFn      func(ctx context.Context, id string) error
	restoreFn     func(ctx context.Context, id string) (Product, error)
	hardDeleteFn  func(ctx context.Context, id string) error

	listCalls        int
	listDeletedCalls int
	getCalls         int
	getDeletedCalls  int
	getBySKUCalls    int
	createCalls      int
	updateCalls      int
	deleteCalls      int
	restoreCalls     int
	hardDeleteCalls  int
}

type fakeAuditRecorder struct {
	events []audit.Event
}

func (f *fakeAuditRecorder) Record(ctx context.Context, event audit.Event) error {
	f.events = append(f.events, event)
	return nil
}

func (f *fakeRepository) List(ctx context.Context, input ListProductsInput) (ListProductsResult, error) {
	f.listCalls++

	if f.listFn != nil {
		return f.listFn(ctx, input)
	}

	return ListProductsResult{}, nil
}

func (f *fakeRepository) ListDeleted(ctx context.Context, input ListProductsInput) (ListProductsResult, error) {
	f.listDeletedCalls++

	if f.listDeletedFn != nil {
		return f.listDeletedFn(ctx, input)
	}

	return ListProductsResult{}, nil
}

func (f *fakeRepository) Get(ctx context.Context, id string) (Product, error) {
	f.getCalls++

	if f.getFn != nil {
		return f.getFn(ctx, id)
	}

	return Product{}, nil
}

func (f *fakeRepository) GetDeleted(ctx context.Context, id string) (Product, error) {
	f.getDeletedCalls++

	if f.getDeletedFn != nil {
		return f.getDeletedFn(ctx, id)
	}

	return Product{}, nil
}

func (f *fakeRepository) GetBySKU(ctx context.Context, sku string) (Product, error) {
	f.getBySKUCalls++

	if f.getBySKUFn != nil {
		return f.getBySKUFn(ctx, sku)
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

func (f *fakeRepository) Restore(ctx context.Context, id string) (Product, error) {
	f.restoreCalls++

	if f.restoreFn != nil {
		return f.restoreFn(ctx, id)
	}

	return Product{}, nil
}

func (f *fakeRepository) HardDelete(ctx context.Context, id string) error {
	f.hardDeleteCalls++

	if f.hardDeleteFn != nil {
		return f.hardDeleteFn(ctx, id)
	}

	return nil
}

func TestServiceListProducts(t *testing.T) {
	expectedResult := ListProductsResult{
		Items: []Product{
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
		},
		Total:      2,
		Page:       1,
		PageSize:   10,
		TotalPages: 1,
		Search:     "lap",
		Sort:       SortFieldName,
		Order:      SortOrderDesc,
	}

	repository := &fakeRepository{
		listFn: func(ctx context.Context, input ListProductsInput) (ListProductsResult, error) {
			if input.Page != 1 {
				t.Fatalf("expected page %d, got %d", 1, input.Page)
			}

			if input.PageSize != 10 {
				t.Fatalf("expected page size %d, got %d", 10, input.PageSize)
			}

			if input.Search != "lap" {
				t.Fatalf("expected search %q, got %q", "lap", input.Search)
			}

			if input.Sort != SortFieldName {
				t.Fatalf("expected sort %q, got %q", SortFieldName, input.Sort)
			}

			if input.Order != SortOrderDesc {
				t.Fatalf("expected order %q, got %q", SortOrderDesc, input.Order)
			}

			return expectedResult, nil
		},
	}

	service := NewService(repository)

	result, err := service.List(context.Background(), ListProductsInput{
		Page:     1,
		PageSize: 10,
		Search:   "  lap  ",
		Sort:     " NAME ",
		Order:    " DESC ",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repository.listCalls != 1 {
		t.Fatalf("expected List to be called once, got %d", repository.listCalls)
	}

	if len(result.Items) != 2 {
		t.Fatalf("expected 2 products, got %d", len(result.Items))
	}

	if result.Total != 2 {
		t.Fatalf("expected total %d, got %d", 2, result.Total)
	}

	if result.Search != "lap" {
		t.Fatalf("expected search %q, got %q", "lap", result.Search)
	}

	if result.Sort != SortFieldName {
		t.Fatalf("expected sort %q, got %q", SortFieldName, result.Sort)
	}

	if result.Order != SortOrderDesc {
		t.Fatalf("expected order %q, got %q", SortOrderDesc, result.Order)
	}
}

func TestServiceListProductsRepositoryError(t *testing.T) {
	repository := &fakeRepository{
		listFn: func(ctx context.Context, input ListProductsInput) (ListProductsResult, error) {
			return ListProductsResult{}, errRepositoryFailure
		},
	}

	service := NewService(repository)

	_, err := service.List(context.Background(), ListProductsInput{
		Page:     1,
		PageSize: 10,
	})
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
			if input.SKU != "LAPTOP-001" {
				t.Fatalf("expected normalized sku %q, got %q", "LAPTOP-001", input.SKU)
			}

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
				SKU:         input.SKU,
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
		SKU:         "  laptop-001  ",
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

	if item.SKU != "LAPTOP-001" {
		t.Fatalf("expected product sku %q, got %q", "LAPTOP-001", item.SKU)
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
		SKU:         "LAPTOP-001",
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
		SKU:         "LAPTOP-001",
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

			if input.SKU != "LAPTOP-PRO-001" {
				t.Fatalf("expected normalized sku %q, got %q", "LAPTOP-PRO-001", input.SKU)
			}

			if input.Name != "Laptop Pro" {
				t.Fatalf("expected trimmed name %q, got %q", "Laptop Pro", input.Name)
			}

			if input.Description != "Laptop para Go" {
				t.Fatalf("expected trimmed description %q, got %q", "Laptop para Go", input.Description)
			}

			return Product{
				ID:          id,
				SKU:         input.SKU,
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
		SKU:         "  laptop-pro-001  ",
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

	if item.SKU != "LAPTOP-PRO-001" {
		t.Fatalf("expected product SKU %q, got %q", "LAPTOP-PRO-001", item.SKU)
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
		SKU:         "LAPTOP-001",
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
		SKU:         "LAPTOP-001",
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
		SKU:         "LAPTOP-001",
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

func TestServiceListProductsValidationErrorDoesNotCallRepository(t *testing.T) {
	repository := &fakeRepository{
		listFn: func(ctx context.Context, input ListProductsInput) (ListProductsResult, error) {
			t.Fatal("repository List should not be called when validation fails")
			return ListProductsResult{}, nil
		},
	}

	service := NewService(repository)

	_, err := service.List(context.Background(), ListProductsInput{
		Page:     -1,
		PageSize: MaxPageSize + 1,
	})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	var validationErr ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}

	if repository.listCalls != 0 {
		t.Fatalf("expected List not to be called, got %d calls", repository.listCalls)
	}

	if len(validationErr.Fields) != 2 {
		t.Fatalf("expected 2 validation fields, got %d", len(validationErr.Fields))
	}
}

func TestServiceListProductsSearchTooLongDoesNotCallRepository(t *testing.T) {
	repository := &fakeRepository{
		listFn: func(ctx context.Context, input ListProductsInput) (ListProductsResult, error) {
			t.Fatal("repository List should not be called when search validation fails")
			return ListProductsResult{}, nil
		},
	}

	service := NewService(repository)

	_, err := service.List(context.Background(), ListProductsInput{
		Page:     1,
		PageSize: 10,
		Search:   strings.Repeat("a", MaxSearchLength+1),
	})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	var validationErr ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}

	if repository.listCalls != 0 {
		t.Fatalf("expected List not to be called, got %d calls", repository.listCalls)
	}

	if len(validationErr.Fields) != 1 {
		t.Fatalf("expected 1 validation field, got %d", len(validationErr.Fields))
	}

	if validationErr.Fields[0].Field != "search" {
		t.Fatalf("expected field %q, got %q", "search", validationErr.Fields[0].Field)
	}
}

func TestServiceListProductsInvalidSortAndOrderDoesNotCallRepository(t *testing.T) {
	repository := &fakeRepository{
		listFn: func(ctx context.Context, input ListProductsInput) (ListProductsResult, error) {
			t.Fatal("repository List should not be called when sort/order validation fails")
			return ListProductsResult{}, nil
		},
	}

	service := NewService(repository)

	_, err := service.List(context.Background(), ListProductsInput{
		Page:     1,
		PageSize: 10,
		Sort:     "unknown",
		Order:    "random",
	})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	var validationErr ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}

	if repository.listCalls != 0 {
		t.Fatalf("expected List not to be called, got %d calls", repository.listCalls)
	}

	if len(validationErr.Fields) != 2 {
		t.Fatalf("expected 2 validation fields, got %d", len(validationErr.Fields))
	}
}

func TestServiceCreateProductSKURequiredDoesNotCallRepository(t *testing.T) {
	repository := &fakeRepository{
		createFn: func(ctx context.Context, input CreateProductInput) (Product, error) {
			t.Fatal("repository Create should not be called when sku validation fails")
			return Product{}, nil
		},
	}

	service := NewService(repository)

	_, err := service.Create(context.Background(), CreateProductInput{
		SKU:         "",
		Name:        "Laptop",
		Description: "Laptop para backend",
		Price:       3500,
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

	if len(validationErr.Fields) != 1 {
		t.Fatalf("expected 1 field violation, got %d", len(validationErr.Fields))
	}

	if validationErr.Fields[0].Field != "sku" {
		t.Fatalf("expected field %q, got %q", "sku", validationErr.Fields[0].Field)
	}
}

func TestServiceGetProductBySKUNormalizesSKU(t *testing.T) {
	repository := &fakeRepository{
		getBySKUFn: func(ctx context.Context, sku string) (Product, error) {
			if sku != "LAPTOP-001" {
				t.Fatalf("expected normalized sku %q, got %q", "LAPTOP-001", sku)
			}

			return Product{
				ID:    "1",
				SKU:   sku,
				Name:  "Laptop",
				Price: 3500,
			}, nil
		},
	}

	service := NewService(repository)

	item, err := service.GetBySKU(context.Background(), "  laptop-001  ")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repository.getBySKUCalls != 1 {
		t.Fatalf("expected GetBySKU to be called once, got %d", repository.getBySKUCalls)
	}

	if item.SKU != "LAPTOP-001" {
		t.Fatalf("expected SKU %q, got %q", "LAPTOP-001", item.SKU)
	}
}

func TestServiceGetProductBySKUNotFound(t *testing.T) {
	repository := &fakeRepository{
		getBySKUFn: func(ctx context.Context, sku string) (Product, error) {
			return Product{}, ErrNotFound
		},
	}

	service := NewService(repository)

	_, err := service.GetBySKU(context.Background(), "missing-sku")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestServiceListProductsWithPriceRange(t *testing.T) {
	minPrice := 1000.0
	maxPrice := 4000.0

	expectedResult := ListProductsResult{
		Items: []Product{
			{
				ID:    "1",
				SKU:   "LAPTOP-001",
				Name:  "Laptop",
				Price: 3500,
			},
		},
		Total:      1,
		Page:       1,
		PageSize:   10,
		TotalPages: 1,
		Sort:       SortFieldPrice,
		Order:      SortOrderDesc,
		MinPrice:   &minPrice,
		MaxPrice:   &maxPrice,
	}

	repository := &fakeRepository{
		listFn: func(ctx context.Context, input ListProductsInput) (ListProductsResult, error) {
			if input.MinPrice == nil || *input.MinPrice != minPrice {
				t.Fatalf("expected min price %v, got %v", minPrice, input.MinPrice)
			}

			if input.MaxPrice == nil || *input.MaxPrice != maxPrice {
				t.Fatalf("expected max price %v, got %v", maxPrice, input.MaxPrice)
			}

			return expectedResult, nil
		},
	}

	service := NewService(repository)

	result, err := service.List(context.Background(), ListProductsInput{
		Page:     1,
		PageSize: 10,
		Sort:     SortFieldPrice,
		Order:    SortOrderDesc,
		MinPrice: &minPrice,
		MaxPrice: &maxPrice,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repository.listCalls != 1 {
		t.Fatalf("expected List to be called once, got %d", repository.listCalls)
	}

	if result.MinPrice == nil || *result.MinPrice != minPrice {
		t.Fatalf("expected result min price %v, got %v", minPrice, result.MinPrice)
	}

	if result.MaxPrice == nil || *result.MaxPrice != maxPrice {
		t.Fatalf("expected result max price %v, got %v", maxPrice, result.MaxPrice)
	}
}

func TestServiceListProductsInvalidPriceRangeDoesNotCallRepository(t *testing.T) {
	minPrice := 5000.0
	maxPrice := 1000.0

	repository := &fakeRepository{
		listFn: func(ctx context.Context, input ListProductsInput) (ListProductsResult, error) {
			t.Fatal("repository List should not be called when price range validation fails")
			return ListProductsResult{}, nil
		},
	}

	service := NewService(repository)

	_, err := service.List(context.Background(), ListProductsInput{
		Page:     1,
		PageSize: 10,
		MinPrice: &minPrice,
		MaxPrice: &maxPrice,
	})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	var validationErr ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}

	if repository.listCalls != 0 {
		t.Fatalf("expected List not to be called, got %d calls", repository.listCalls)
	}

	if len(validationErr.Fields) != 1 {
		t.Fatalf("expected 1 validation field, got %d", len(validationErr.Fields))
	}

	if validationErr.Fields[0].Field != "price_range" {
		t.Fatalf("expected field %q, got %q", "price_range", validationErr.Fields[0].Field)
	}
}

func TestServiceListProductsWithCreatedRange(t *testing.T) {
	createdFrom := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	createdTo := time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC)

	expectedResult := ListProductsResult{
		Items: []Product{
			{
				ID:        "1",
				SKU:       "LAPTOP-001",
				Name:      "Laptop",
				Price:     3500,
				CreatedAt: createdFrom.Add(24 * time.Hour),
			},
		},
		Total:       1,
		Page:        1,
		PageSize:    10,
		TotalPages:  1,
		Sort:        SortFieldCreatedAt,
		Order:       SortOrderDesc,
		CreatedFrom: &createdFrom,
		CreatedTo:   &createdTo,
	}

	repository := &fakeRepository{
		listFn: func(ctx context.Context, input ListProductsInput) (ListProductsResult, error) {
			if input.CreatedFrom == nil || !input.CreatedFrom.Equal(createdFrom) {
				t.Fatalf("expected created from %v, got %v", createdFrom, input.CreatedFrom)
			}

			if input.CreatedTo == nil || !input.CreatedTo.Equal(createdTo) {
				t.Fatalf("expected created to %v, got %v", createdTo, input.CreatedTo)
			}

			return expectedResult, nil
		},
	}

	service := NewService(repository)

	result, err := service.List(context.Background(), ListProductsInput{
		Page:        1,
		PageSize:    10,
		Sort:        SortFieldCreatedAt,
		Order:       SortOrderDesc,
		CreatedFrom: &createdFrom,
		CreatedTo:   &createdTo,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repository.listCalls != 1 {
		t.Fatalf("expected List to be called once, got %d", repository.listCalls)
	}

	if result.CreatedFrom == nil || !result.CreatedFrom.Equal(createdFrom) {
		t.Fatalf("expected result created from %v, got %v", createdFrom, result.CreatedFrom)
	}

	if result.CreatedTo == nil || !result.CreatedTo.Equal(createdTo) {
		t.Fatalf("expected result created to %v, got %v", createdTo, result.CreatedTo)
	}
}

func TestServiceListProductsInvalidCreatedRangeDoesNotCallRepository(t *testing.T) {
	createdFrom := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	createdTo := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	repository := &fakeRepository{
		listFn: func(ctx context.Context, input ListProductsInput) (ListProductsResult, error) {
			t.Fatal("repository List should not be called when created range validation fails")
			return ListProductsResult{}, nil
		},
	}

	service := NewService(repository)

	_, err := service.List(context.Background(), ListProductsInput{
		Page:        1,
		PageSize:    10,
		CreatedFrom: &createdFrom,
		CreatedTo:   &createdTo,
	})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	var validationErr ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}

	if repository.listCalls != 0 {
		t.Fatalf("expected List not to be called, got %d calls", repository.listCalls)
	}

	if len(validationErr.Fields) != 1 {
		t.Fatalf("expected 1 validation field, got %d", len(validationErr.Fields))
	}

	if validationErr.Fields[0].Field != "created_range" {
		t.Fatalf("expected field %q, got %q", "created_range", validationErr.Fields[0].Field)
	}
}

func TestServiceSKUExistsReturnsTrue(t *testing.T) {
	repository := &fakeRepository{
		getBySKUFn: func(ctx context.Context, sku string) (Product, error) {
			if sku != "LAPTOP-001" {
				t.Fatalf("expected normalized sku %q, got %q", "LAPTOP-001", sku)
			}

			return Product{
				ID:    "1",
				SKU:   sku,
				Name:  "Laptop",
				Price: 3500,
			}, nil
		},
	}

	service := NewService(repository)

	result, err := service.SKUExists(context.Background(), "  laptop-001  ")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repository.getBySKUCalls != 1 {
		t.Fatalf("expected GetBySKU to be called once, got %d", repository.getBySKUCalls)
	}

	if result.SKU != "LAPTOP-001" {
		t.Fatalf("expected sku %q, got %q", "LAPTOP-001", result.SKU)
	}

	if !result.Exists {
		t.Fatal("expected exists to be true")
	}
}

func TestServiceSKUExistsReturnsFalseWhenNotFound(t *testing.T) {
	repository := &fakeRepository{
		getBySKUFn: func(ctx context.Context, sku string) (Product, error) {
			return Product{}, ErrNotFound
		},
	}

	service := NewService(repository)

	result, err := service.SKUExists(context.Background(), "missing-sku")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repository.getBySKUCalls != 1 {
		t.Fatalf("expected GetBySKU to be called once, got %d", repository.getBySKUCalls)
	}

	if result.SKU != "MISSING-SKU" {
		t.Fatalf("expected sku %q, got %q", "MISSING-SKU", result.SKU)
	}

	if result.Exists {
		t.Fatal("expected exists to be false")
	}
}

func TestServiceSKUExistsRepositoryError(t *testing.T) {
	repository := &fakeRepository{
		getBySKUFn: func(ctx context.Context, sku string) (Product, error) {
			return Product{}, errRepositoryFailure
		},
	}

	service := NewService(repository)

	_, err := service.SKUExists(context.Background(), "laptop-001")
	if !errors.Is(err, errRepositoryFailure) {
		t.Fatalf("expected repository error, got %v", err)
	}
}

func TestServicePatchProductMergesExistingProduct(t *testing.T) {
	now := time.Now().UTC()

	newName := "Laptop Pro"
	newPrice := 4200.0

	repository := &fakeRepository{
		getFn: func(ctx context.Context, id string) (Product, error) {
			if id != "123" {
				t.Fatalf("expected id %q, got %q", "123", id)
			}

			return Product{
				ID:          id,
				SKU:         "LAPTOP-001",
				Name:        "Laptop",
				Description: "Laptop básica",
				Price:       3500,
				CreatedAt:   now,
				UpdatedAt:   now,
			}, nil
		},
		updateFn: func(ctx context.Context, id string, input UpdateProductInput) (Product, error) {
			if id != "123" {
				t.Fatalf("expected id %q, got %q", "123", id)
			}

			if input.SKU != "LAPTOP-001" {
				t.Fatalf("expected preserved sku %q, got %q", "LAPTOP-001", input.SKU)
			}

			if input.Name != newName {
				t.Fatalf("expected patched name %q, got %q", newName, input.Name)
			}

			if input.Description != "Laptop básica" {
				t.Fatalf("expected preserved description %q, got %q", "Laptop básica", input.Description)
			}

			if input.Price != newPrice {
				t.Fatalf("expected patched price %v, got %v", newPrice, input.Price)
			}

			return Product{
				ID:          id,
				SKU:         input.SKU,
				Name:        input.Name,
				Description: input.Description,
				Price:       input.Price,
				CreatedAt:   now,
				UpdatedAt:   time.Now().UTC(),
			}, nil
		},
	}

	service := NewService(repository)

	item, err := service.Patch(context.Background(), "  123  ", PatchProductInput{
		Name:  &newName,
		Price: &newPrice,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repository.getCalls != 1 {
		t.Fatalf("expected Get to be called once, got %d", repository.getCalls)
	}

	if repository.updateCalls != 1 {
		t.Fatalf("expected Update to be called once, got %d", repository.updateCalls)
	}

	if item.Name != newName {
		t.Fatalf("expected name %q, got %q", newName, item.Name)
	}

	if item.Price != newPrice {
		t.Fatalf("expected price %v, got %v", newPrice, item.Price)
	}
}

func TestServicePatchProductNormalizesSKU(t *testing.T) {
	now := time.Now().UTC()

	newSKU := "  laptop-pro-001  "

	repository := &fakeRepository{
		getFn: func(ctx context.Context, id string) (Product, error) {
			return Product{
				ID:          id,
				SKU:         "LAPTOP-001",
				Name:        "Laptop",
				Description: "Laptop básica",
				Price:       3500,
				CreatedAt:   now,
				UpdatedAt:   now,
			}, nil
		},
		updateFn: func(ctx context.Context, id string, input UpdateProductInput) (Product, error) {
			if input.SKU != "LAPTOP-PRO-001" {
				t.Fatalf("expected normalized sku %q, got %q", "LAPTOP-PRO-001", input.SKU)
			}

			return Product{
				ID:          id,
				SKU:         input.SKU,
				Name:        input.Name,
				Description: input.Description,
				Price:       input.Price,
				CreatedAt:   now,
				UpdatedAt:   time.Now().UTC(),
			}, nil
		},
	}

	service := NewService(repository)

	item, err := service.Patch(context.Background(), "123", PatchProductInput{
		SKU: &newSKU,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if item.SKU != "LAPTOP-PRO-001" {
		t.Fatalf("expected sku %q, got %q", "LAPTOP-PRO-001", item.SKU)
	}
}

func TestServicePatchProductEmptyBodyDoesNotCallRepository(t *testing.T) {
	repository := &fakeRepository{
		getFn: func(ctx context.Context, id string) (Product, error) {
			t.Fatal("repository Get should not be called when patch body is empty")
			return Product{}, nil
		},
		updateFn: func(ctx context.Context, id string, input UpdateProductInput) (Product, error) {
			t.Fatal("repository Update should not be called when patch body is empty")
			return Product{}, nil
		},
	}

	service := NewService(repository)

	_, err := service.Patch(context.Background(), "123", PatchProductInput{})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	var validationErr ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}

	if repository.getCalls != 0 {
		t.Fatalf("expected Get not to be called, got %d calls", repository.getCalls)
	}

	if repository.updateCalls != 0 {
		t.Fatalf("expected Update not to be called, got %d calls", repository.updateCalls)
	}

	if len(validationErr.Fields) != 1 {
		t.Fatalf("expected 1 field violation, got %d", len(validationErr.Fields))
	}

	if validationErr.Fields[0].Field != "body" {
		t.Fatalf("expected field %q, got %q", "body", validationErr.Fields[0].Field)
	}
}

func TestServicePatchProductNotFound(t *testing.T) {
	newName := "Laptop Pro"

	repository := &fakeRepository{
		getFn: func(ctx context.Context, id string) (Product, error) {
			return Product{}, ErrNotFound
		},
		updateFn: func(ctx context.Context, id string, input UpdateProductInput) (Product, error) {
			t.Fatal("repository Update should not be called when product is not found")
			return Product{}, nil
		},
	}

	service := NewService(repository)

	_, err := service.Patch(context.Background(), "999", PatchProductInput{
		Name: &newName,
	})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	if repository.getCalls != 1 {
		t.Fatalf("expected Get to be called once, got %d", repository.getCalls)
	}

	if repository.updateCalls != 0 {
		t.Fatalf("expected Update not to be called, got %d calls", repository.updateCalls)
	}
}

func TestServicePatchProductDuplicateSKU(t *testing.T) {
	newSKU := "DUPLICATE-SKU"

	repository := &fakeRepository{
		getFn: func(ctx context.Context, id string) (Product, error) {
			return Product{
				ID:          id,
				SKU:         "CURRENT-SKU",
				Name:        "Laptop",
				Description: "Laptop básica",
				Price:       3500,
			}, nil
		},
		updateFn: func(ctx context.Context, id string, input UpdateProductInput) (Product, error) {
			return Product{}, ErrSKUAlreadyExists
		},
	}

	service := NewService(repository)

	_, err := service.Patch(context.Background(), "123", PatchProductInput{
		SKU: &newSKU,
	})
	if !errors.Is(err, ErrSKUAlreadyExists) {
		t.Fatalf("expected ErrSKUAlreadyExists, got %v", err)
	}
}

func TestServiceRestoreProduct(t *testing.T) {
	now := time.Now().UTC()

	repository := &fakeRepository{
		restoreFn: func(ctx context.Context, id string) (Product, error) {
			if id != "123" {
				t.Fatalf("expected trimmed id %q, got %q", "123", id)
			}

			return Product{
				ID:          id,
				SKU:         "RESTORE-001",
				Name:        "Restored Product",
				Description: "Producto restaurado",
				Price:       100,
				CreatedAt:   now,
				UpdatedAt:   now,
			}, nil
		},
	}

	service := NewService(repository)

	item, err := service.Restore(context.Background(), "  123  ")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repository.restoreCalls != 1 {
		t.Fatalf("expected Restore to be called once, got %d", repository.restoreCalls)
	}

	if item.ID != "123" {
		t.Fatalf("expected ID %q, got %q", "123", item.ID)
	}

	if item.SKU != "RESTORE-001" {
		t.Fatalf("expected SKU %q, got %q", "RESTORE-001", item.SKU)
	}
}

func TestServiceRestoreProductNotFound(t *testing.T) {
	repository := &fakeRepository{
		restoreFn: func(ctx context.Context, id string) (Product, error) {
			return Product{}, ErrNotFound
		},
	}

	service := NewService(repository)

	_, err := service.Restore(context.Background(), "999")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestServiceRestoreProductDuplicateSKU(t *testing.T) {
	repository := &fakeRepository{
		restoreFn: func(ctx context.Context, id string) (Product, error) {
			return Product{}, ErrSKUAlreadyExists
		},
	}

	service := NewService(repository)

	_, err := service.Restore(context.Background(), "123")
	if !errors.Is(err, ErrSKUAlreadyExists) {
		t.Fatalf("expected ErrSKUAlreadyExists, got %v", err)
	}
}

func TestServiceListDeletedProducts(t *testing.T) {
	deletedAt := time.Now().UTC()

	expectedResult := ListProductsResult{
		Items: []Product{
			{
				ID:        "1",
				SKU:       "DELETED-001",
				Name:      "Deleted Product",
				Price:     100,
				DeletedAt: &deletedAt,
			},
		},
		Total:      1,
		Page:       1,
		PageSize:   10,
		TotalPages: 1,
		Sort:       SortFieldID,
		Order:      SortOrderAsc,
	}

	repository := &fakeRepository{
		listDeletedFn: func(ctx context.Context, input ListProductsInput) (ListProductsResult, error) {
			if input.Page != 1 {
				t.Fatalf("expected page %d, got %d", 1, input.Page)
			}

			if input.PageSize != 10 {
				t.Fatalf("expected page size %d, got %d", 10, input.PageSize)
			}

			return expectedResult, nil
		},
	}

	service := NewService(repository)

	result, err := service.ListDeleted(context.Background(), ListProductsInput{
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repository.listDeletedCalls != 1 {
		t.Fatalf("expected ListDeleted to be called once, got %d", repository.listDeletedCalls)
	}

	if result.Total != 1 {
		t.Fatalf("expected total %d, got %d", 1, result.Total)
	}

	if len(result.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result.Items))
	}

	if result.Items[0].DeletedAt == nil {
		t.Fatal("expected deleted_at to be set")
	}
}

func TestServiceHardDeleteProduct(t *testing.T) {
	repository := &fakeRepository{
		hardDeleteFn: func(ctx context.Context, id string) error {
			if id != "123" {
				t.Fatalf("expected trimmed id %q, got %q", "123", id)
			}

			return nil
		},
	}

	service := NewService(repository)

	if err := service.HardDelete(context.Background(), "  123  "); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repository.hardDeleteCalls != 1 {
		t.Fatalf("expected HardDelete to be called once, got %d", repository.hardDeleteCalls)
	}
}

func TestServiceHardDeleteProductNotFound(t *testing.T) {
	repository := &fakeRepository{
		hardDeleteFn: func(ctx context.Context, id string) error {
			return ErrNotFound
		},
	}

	service := NewService(repository)

	err := service.HardDelete(context.Background(), "999")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestServiceHardDeleteProductMustBeSoftDeletedFirst(t *testing.T) {
	repository := &fakeRepository{
		hardDeleteFn: func(ctx context.Context, id string) error {
			return ErrProductMustBeDeletedFirst
		},
	}

	service := NewService(repository)

	err := service.HardDelete(context.Background(), "123")
	if !errors.Is(err, ErrProductMustBeDeletedFirst) {
		t.Fatalf("expected ErrProductMustBeDeletedFirst, got %v", err)
	}
}

func TestServiceCreateProductRecordsAuditEvent(t *testing.T) {
	now := time.Now().UTC()

	auditor := &fakeAuditRecorder{}

	repository := &fakeRepository{
		createFn: func(ctx context.Context, input CreateProductInput) (Product, error) {
			return Product{
				ID:          "1",
				SKU:         input.SKU,
				Name:        input.Name,
				Description: input.Description,
				Price:       input.Price,
				CreatedAt:   now,
				UpdatedAt:   now,
			}, nil
		},
	}

	service := NewServiceWithAuditor(repository, auditor)

	_, err := service.Create(context.Background(), CreateProductInput{
		SKU:         "AUDIT-CREATE-001",
		Name:        "Audit Create",
		Description: "Producto auditado",
		Price:       100,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(auditor.events) != 1 {
		t.Fatalf("expected 1 audit event, got %d", len(auditor.events))
	}

	event := auditor.events[0]

	if event.Type != AuditEventProductCreated {
		t.Fatalf("expected event type %q, got %q", AuditEventProductCreated, event.Type)
	}

	if event.AggregateType != AuditAggregateProduct {
		t.Fatalf("expected aggregate type %q, got %q", AuditAggregateProduct, event.AggregateType)
	}

	if event.AggregateID != "1" {
		t.Fatalf("expected aggregate id %q, got %q", "1", event.AggregateID)
	}
}

func TestServiceUpdateProductRecordsAuditEvent(t *testing.T) {
	now := time.Now().UTC()

	auditor := &fakeAuditRecorder{}

	repository := &fakeRepository{
		updateFn: func(ctx context.Context, id string, input UpdateProductInput) (Product, error) {
			return Product{
				ID:          id,
				SKU:         input.SKU,
				Name:        input.Name,
				Description: input.Description,
				Price:       input.Price,
				CreatedAt:   now,
				UpdatedAt:   now,
			}, nil
		},
	}

	service := NewServiceWithAuditor(repository, auditor)

	_, err := service.Update(context.Background(), "123", UpdateProductInput{
		SKU:         "AUDIT-UPDATE-001",
		Name:        "Audit Update",
		Description: "Producto actualizado",
		Price:       200,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(auditor.events) != 1 {
		t.Fatalf("expected 1 audit event, got %d", len(auditor.events))
	}

	if auditor.events[0].Type != AuditEventProductUpdated {
		t.Fatalf("expected event type %q, got %q", AuditEventProductUpdated, auditor.events[0].Type)
	}
}

func TestServiceDeleteProductRecordsAuditEvent(t *testing.T) {
	now := time.Now().UTC()

	auditor := &fakeAuditRecorder{}

	repository := &fakeRepository{
		getFn: func(ctx context.Context, id string) (Product, error) {
			return Product{
				ID:          id,
				SKU:         "AUDIT-DELETE-001",
				Name:        "Audit Delete",
				Description: "Producto eliminado",
				Price:       100,
				CreatedAt:   now,
				UpdatedAt:   now,
			}, nil
		},
		deleteFn: func(ctx context.Context, id string) error {
			return nil
		},
	}

	service := NewServiceWithAuditor(repository, auditor)

	if err := service.Delete(context.Background(), "123"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(auditor.events) != 1 {
		t.Fatalf("expected 1 audit event, got %d", len(auditor.events))
	}

	if auditor.events[0].Type != AuditEventProductDeleted {
		t.Fatalf("expected event type %q, got %q", AuditEventProductDeleted, auditor.events[0].Type)
	}
}

func TestServiceRestoreProductRecordsAuditEvent(t *testing.T) {
	now := time.Now().UTC()

	auditor := &fakeAuditRecorder{}

	repository := &fakeRepository{
		restoreFn: func(ctx context.Context, id string) (Product, error) {
			return Product{
				ID:          id,
				SKU:         "AUDIT-RESTORE-001",
				Name:        "Audit Restore",
				Description: "Producto restaurado",
				Price:       100,
				CreatedAt:   now,
				UpdatedAt:   now,
			}, nil
		},
	}

	service := NewServiceWithAuditor(repository, auditor)

	_, err := service.Restore(context.Background(), "123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(auditor.events) != 1 {
		t.Fatalf("expected 1 audit event, got %d", len(auditor.events))
	}

	if auditor.events[0].Type != AuditEventProductRestored {
		t.Fatalf("expected event type %q, got %q", AuditEventProductRestored, auditor.events[0].Type)
	}
}

func TestServiceHardDeleteProductRecordsAuditEvent(t *testing.T) {
	now := time.Now().UTC()
	deletedAt := now.Add(time.Minute)

	auditor := &fakeAuditRecorder{}

	repository := &fakeRepository{
		getDeletedFn: func(ctx context.Context, id string) (Product, error) {
			return Product{
				ID:          id,
				SKU:         "AUDIT-HARD-DELETE-001",
				Name:        "Audit Hard Delete",
				Description: "Producto eliminado físicamente",
				Price:       100,
				CreatedAt:   now,
				UpdatedAt:   deletedAt,
				DeletedAt:   &deletedAt,
			}, nil
		},
		hardDeleteFn: func(ctx context.Context, id string) error {
			return nil
		},
	}

	service := NewServiceWithAuditor(repository, auditor)

	if err := service.HardDelete(context.Background(), "123"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(auditor.events) != 1 {
		t.Fatalf("expected 1 audit event, got %d", len(auditor.events))
	}

	if auditor.events[0].Type != AuditEventProductHardDeleted {
		t.Fatalf("expected event type %q, got %q", AuditEventProductHardDeleted, auditor.events[0].Type)
	}
}
