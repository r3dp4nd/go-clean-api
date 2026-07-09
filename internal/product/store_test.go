package product

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestStoreCreateProduct(t *testing.T) {
	store := NewStore()
	ctx := context.Background()

	item, err := store.Create(ctx, CreateProductInput{
		SKU:         "LAPTOP-001",
		Name:        "Laptop",
		Description: "Laptop para desarrollo backend",
		Price:       3500,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if item.ID == "" {
		t.Fatal("expected product ID to be generated")
	}

	if item.Name != "Laptop" {
		t.Fatalf("expected name %q, got %q", "Laptop", item.Name)
	}

	if item.Description != "Laptop para desarrollo backend" {
		t.Fatalf("expected description %q, got %q", "Laptop para desarrollo backend", item.Description)
	}

	if item.Price != 3500 {
		t.Fatalf("expected price %v, got %v", 3500.0, item.Price)
	}

	if item.CreatedAt.IsZero() {
		t.Fatal("expected CreatedAt to be set")
	}

	if item.UpdatedAt.IsZero() {
		t.Fatal("expected UpdatedAt to be set")
	}

	if !item.CreatedAt.Equal(item.UpdatedAt) {
		t.Fatal("expected CreatedAt and UpdatedAt to be equal on create")
	}
}

func TestStoreListProducts(t *testing.T) {
	store := NewStore()
	ctx := context.Background()

	_, err := store.Create(ctx, CreateProductInput{
		SKU:         "MOUSE-001",
		Name:        "Mouse",
		Description: "Mouse inalámbrico",
		Price:       120,
	})
	if err != nil {
		t.Fatalf("expected no error creating mouse, got %v", err)
	}

	_, err = store.Create(ctx, CreateProductInput{
		SKU:         "KEYBOARD-001",
		Name:        "Keyboard",
		Description: "Teclado mecánico",
		Price:       250,
	})
	if err != nil {
		t.Fatalf("expected no error creating keyboard, got %v", err)
	}

	result, err := store.List(ctx, ListProductsInput{
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("expected no error listing products, got %v", err)
	}

	if len(result.Items) != 2 {
		t.Fatalf("expected 2 products, got %d", len(result.Items))
	}

	if result.Total != 2 {
		t.Fatalf("expected total %d, got %d", 2, result.Total)
	}

	if result.TotalPages != 1 {
		t.Fatalf("expected total pages %d, got %d", 1, result.TotalPages)
	}

	if result.Items[0].ID != "1" {
		t.Fatalf("expected first product ID %q, got %q", "1", result.Items[0].ID)
	}

	if result.Items[1].ID != "2" {
		t.Fatalf("expected second product ID %q, got %q", "2", result.Items[1].ID)
	}
}

func TestStoreGetProduct(t *testing.T) {
	store := NewStore()
	ctx := context.Background()

	created, err := store.Create(ctx, CreateProductInput{
		SKU:         "MONITOR-001",
		Name:        "Monitor",
		Description: "Monitor 27 pulgadas",
		Price:       900,
	})
	if err != nil {
		t.Fatalf("expected no error creating product, got %v", err)
	}

	found, err := store.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("expected no error getting product, got %v", err)
	}

	if found.ID != created.ID {
		t.Fatalf("expected ID %q, got %q", created.ID, found.ID)
	}

	if found.Name != created.Name {
		t.Fatalf("expected name %q, got %q", created.Name, found.Name)
	}
}

func TestStoreGetProductNotFound(t *testing.T) {
	store := NewStore()
	ctx := context.Background()

	_, err := store.Get(ctx, "999")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestStoreUpdateProduct(t *testing.T) {
	store := NewStore()
	ctx := context.Background()

	created, err := store.Create(ctx, CreateProductInput{
		SKU:         "LAPTOP-001",
		Name:        "Laptop",
		Description: "Laptop básica",
		Price:       3000,
	})
	if err != nil {
		t.Fatalf("expected no error creating product, got %v", err)
	}

	updated, err := store.Update(ctx, created.ID, UpdateProductInput{
		SKU:         "LAPTOP-PRO-001",
		Name:        "Laptop Pro",
		Description: "Laptop para Go, Docker y Kubernetes",
		Price:       4200,
	})
	if err != nil {
		t.Fatalf("expected no error updating product, got %v", err)
	}

	if updated.ID != created.ID {
		t.Fatalf("expected same ID %q, got %q", created.ID, updated.ID)
	}

	if updated.Name != "Laptop Pro" {
		t.Fatalf("expected name %q, got %q", "Laptop Pro", updated.Name)
	}

	if updated.SKU != "LAPTOP-PRO-001" {
		t.Fatalf("expected sku %q, got %q", "LAPTOP-PRO-001", updated.SKU)
	}

	if updated.Description != "Laptop para Go, Docker y Kubernetes" {
		t.Fatalf("expected updated description, got %q", updated.Description)
	}

	if updated.Price != 4200 {
		t.Fatalf("expected price %v, got %v", 4200.0, updated.Price)
	}

	if !updated.CreatedAt.Equal(created.CreatedAt) {
		t.Fatal("expected CreatedAt to remain unchanged")
	}

	if !updated.UpdatedAt.After(created.UpdatedAt) && !updated.UpdatedAt.Equal(created.UpdatedAt) {
		t.Fatal("expected UpdatedAt to be updated")
	}
}

func TestStoreUpdateProductNotFound(t *testing.T) {
	store := NewStore()
	ctx := context.Background()

	_, err := store.Update(ctx, "999", UpdateProductInput{
		SKU:         "NOT-FOUND-001",
		Name:        "Not found",
		Description: "Producto inexistente",
		Price:       100,
	})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestStoreDeleteProduct(t *testing.T) {
	store := NewStore()
	ctx := context.Background()

	created, err := store.Create(ctx, CreateProductInput{
		SKU:         "TABLET-001",
		Name:        "Tablet",
		Description: "Tablet para pruebas",
		Price:       1500,
	})
	if err != nil {
		t.Fatalf("expected no error creating product, got %v", err)
	}

	if err := store.Delete(ctx, created.ID); err != nil {
		t.Fatalf("expected no error deleting product, got %v", err)
	}

	_, err = store.Get(ctx, created.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestStoreDeleteProductNotFound(t *testing.T) {
	store := NewStore()
	ctx := context.Background()

	err := store.Delete(ctx, "999")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestStoreContextCanceled(t *testing.T) {
	store := NewStore()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := store.List(ctx, ListProductsInput{
		Page:     1,
		PageSize: 10,
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled on List, got %v", err)
	}

	_, err = store.Get(ctx, "1")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled on Get, got %v", err)
	}

	_, err = store.Create(ctx, CreateProductInput{
		SKU:         "CANCELED-001",
		Name:        "Canceled",
		Description: "No debe crearse",
		Price:       10,
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled on Create, got %v", err)
	}

	_, err = store.Update(ctx, "1", UpdateProductInput{
		SKU:         "CANCELED-002",
		Name:        "Canceled",
		Description: "No debe actualizarse",
		Price:       10,
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled on Update, got %v", err)
	}

	err = store.Delete(ctx, "1")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled on Delete, got %v", err)
	}
}

func TestStoreConcurrentCreates(t *testing.T) {
	store := NewStore()
	ctx := context.Background()

	const totalProducts = 100

	var wg sync.WaitGroup
	wg.Add(totalProducts)

	for i := 0; i < totalProducts; i++ {
		go func(index int) {
			defer wg.Done()

			_, err := store.Create(ctx, CreateProductInput{
				SKU:         fmt.Sprintf("PRODUCT-%03d", index),
				Name:        fmt.Sprintf("Product %d", index),
				Description: "Producto creado concurrentemente",
				Price:       float64(index),
			})
			if err != nil {
				t.Errorf("expected no error creating product %d, got %v", index, err)
			}
		}(i)
	}

	wg.Wait()

	result, err := store.List(ctx, ListProductsInput{
		Page:     1,
		PageSize: totalProducts,
	})
	if err != nil {
		t.Fatalf("expected no error listing products, got %v", err)
	}

	if len(result.Items) != totalProducts {
		t.Fatalf("expected %d products, got %d", totalProducts, len(result.Items))
	}

	seenIDs := make(map[string]bool, totalProducts)

	for _, item := range result.Items {
		if item.ID == "" {
			t.Fatal("expected product ID to not be empty")
		}

		if seenIDs[item.ID] {
			t.Fatalf("duplicated product ID detected: %s", item.ID)
		}

		seenIDs[item.ID] = true
	}
}

func TestStoreListProductsPagination(t *testing.T) {
	store := NewStore()
	ctx := context.Background()

	for i := 1; i <= 5; i++ {
		_, err := store.Create(ctx, CreateProductInput{
			SKU:         fmt.Sprintf("PRODUCT-PAGE-%03d", i),
			Name:        fmt.Sprintf("Product %d", i),
			Description: "Producto paginado",
			Price:       float64(i),
		})
		if err != nil {
			t.Fatalf("expected no error creating product %d, got %v", i, err)
		}
	}

	result, err := store.List(ctx, ListProductsInput{
		Page:     2,
		PageSize: 2,
	})
	if err != nil {
		t.Fatalf("expected no error listing products, got %v", err)
	}

	if len(result.Items) != 2 {
		t.Fatalf("expected 2 products on page 2, got %d", len(result.Items))
	}

	if result.Items[0].ID != "3" {
		t.Fatalf("expected first item on page 2 to have ID %q, got %q", "3", result.Items[0].ID)
	}

	if result.Items[1].ID != "4" {
		t.Fatalf("expected second item on page 2 to have ID %q, got %q", "4", result.Items[1].ID)
	}

	if result.Total != 5 {
		t.Fatalf("expected total %d, got %d", 5, result.Total)
	}

	if result.TotalPages != 3 {
		t.Fatalf("expected total pages %d, got %d", 3, result.TotalPages)
	}
}

func TestStoreListProductsSearchByName(t *testing.T) {
	store := NewStore()
	ctx := context.Background()

	_, err := store.Create(ctx, CreateProductInput{
		SKU:         "LAPTOP-SEARCH-001",
		Name:        "Laptop",
		Description: "Equipo para desarrollo backend",
		Price:       3500,
	})
	if err != nil {
		t.Fatalf("expected no error creating laptop, got %v", err)
	}

	_, err = store.Create(ctx, CreateProductInput{
		SKU:         "MOUSE-SEARCH-001",
		Name:        "Mouse",
		Description: "Mouse inalámbrico",
		Price:       120,
	})
	if err != nil {
		t.Fatalf("expected no error creating mouse, got %v", err)
	}

	result, err := store.List(ctx, ListProductsInput{
		Page:     1,
		PageSize: 10,
		Search:   "lap",
	})
	if err != nil {
		t.Fatalf("expected no error listing products, got %v", err)
	}

	if len(result.Items) != 1 {
		t.Fatalf("expected 1 product, got %d", len(result.Items))
	}

	if result.Items[0].Name != "Laptop" {
		t.Fatalf("expected product %q, got %q", "Laptop", result.Items[0].Name)
	}

	if result.Total != 1 {
		t.Fatalf("expected total %d, got %d", 1, result.Total)
	}

	if result.Search != "lap" {
		t.Fatalf("expected search %q, got %q", "lap", result.Search)
	}
}

func TestStoreListProductsSearchByDescription(t *testing.T) {
	store := NewStore()
	ctx := context.Background()

	_, err := store.Create(ctx, CreateProductInput{
		SKU:         "LAPTOP-DESC-001",
		Name:        "Laptop",
		Description: "Equipo para desarrollo backend",
		Price:       3500,
	})
	if err != nil {
		t.Fatalf("expected no error creating laptop, got %v", err)
	}

	_, err = store.Create(ctx, CreateProductInput{
		SKU:         "KEYBOARD-DESC-001",
		Name:        "Keyboard",
		Description: "Teclado mecánico",
		Price:       250,
	})
	if err != nil {
		t.Fatalf("expected no error creating keyboard, got %v", err)
	}

	result, err := store.List(ctx, ListProductsInput{
		Page:     1,
		PageSize: 10,
		Search:   "backend",
	})
	if err != nil {
		t.Fatalf("expected no error listing products, got %v", err)
	}

	if len(result.Items) != 1 {
		t.Fatalf("expected 1 product, got %d", len(result.Items))
	}

	if result.Items[0].Name != "Laptop" {
		t.Fatalf("expected product %q, got %q", "Laptop", result.Items[0].Name)
	}
}

func TestStoreListProductsSortByNameAscending(t *testing.T) {
	store := NewStore()
	ctx := context.Background()

	products := []CreateProductInput{
		{
			SKU:         "MOUSE-SORT-NAME",
			Name:        "Mouse",
			Description: "Mouse inalámbrico",
			Price:       120,
		},
		{
			SKU:         "LAPTOP-SORT-NAME",
			Name:        "Laptop",
			Description: "Equipo para desarrollo backend",
			Price:       3500,
		},
		{
			SKU:         "KEYBOARD-SORT-NAME",
			Name:        "Keyboard",
			Description: "Teclado mecánico",
			Price:       250,
		},
	}

	for _, input := range products {
		if _, err := store.Create(ctx, input); err != nil {
			t.Fatalf("expected no error creating product, got %v", err)
		}
	}

	result, err := store.List(ctx, ListProductsInput{
		Page:     1,
		PageSize: 10,
		Sort:     SortFieldName,
		Order:    SortOrderAsc,
	})
	if err != nil {
		t.Fatalf("expected no error listing products, got %v", err)
	}

	if len(result.Items) != 3 {
		t.Fatalf("expected 3 products, got %d", len(result.Items))
	}

	expectedNames := []string{"Keyboard", "Laptop", "Mouse"}

	for index, expectedName := range expectedNames {
		if result.Items[index].Name != expectedName {
			t.Fatalf("expected product at index %d to be %q, got %q", index, expectedName, result.Items[index].Name)
		}
	}

	if result.Sort != SortFieldName {
		t.Fatalf("expected sort %q, got %q", SortFieldName, result.Sort)
	}

	if result.Order != SortOrderAsc {
		t.Fatalf("expected order %q, got %q", SortOrderAsc, result.Order)
	}
}

func TestStoreListProductsSortByPriceDescending(t *testing.T) {
	store := NewStore()
	ctx := context.Background()

	products := []CreateProductInput{
		{
			SKU:         "MOUSE-SORT-PRICE",
			Name:        "Mouse",
			Description: "Mouse inalámbrico",
			Price:       120,
		},
		{
			SKU:         "LAPTOP-SORT-PRICE",
			Name:        "Laptop",
			Description: "Equipo para desarrollo backend",
			Price:       3500,
		},
		{
			SKU:         "KEYBOARD-SORT-PRICE",
			Name:        "Keyboard",
			Description: "Teclado mecánico",
			Price:       250,
		},
	}

	for _, input := range products {
		if _, err := store.Create(ctx, input); err != nil {
			t.Fatalf("expected no error creating product, got %v", err)
		}
	}

	result, err := store.List(ctx, ListProductsInput{
		Page:     1,
		PageSize: 10,
		Sort:     SortFieldPrice,
		Order:    SortOrderDesc,
	})
	if err != nil {
		t.Fatalf("expected no error listing products, got %v", err)
	}

	if len(result.Items) != 3 {
		t.Fatalf("expected 3 products, got %d", len(result.Items))
	}

	expectedNames := []string{"Laptop", "Keyboard", "Mouse"}

	for index, expectedName := range expectedNames {
		if result.Items[index].Name != expectedName {
			t.Fatalf("expected product at index %d to be %q, got %q", index, expectedName, result.Items[index].Name)
		}
	}

	if result.Sort != SortFieldPrice {
		t.Fatalf("expected sort %q, got %q", SortFieldPrice, result.Sort)
	}

	if result.Order != SortOrderDesc {
		t.Fatalf("expected order %q, got %q", SortOrderDesc, result.Order)
	}
}

func TestStoreListProductsSearchSortAndPagination(t *testing.T) {
	store := NewStore()
	ctx := context.Background()

	products := []CreateProductInput{
		{
			SKU:         "LAPTOP-BASIC-SEARCH-SORT",
			Name:        "Laptop Basic",
			Description: "Laptop para oficina",
			Price:       2500,
		},
		{
			SKU:         "LAPTOP-PRO-SEARCH-SORT",
			Name:        "Laptop Pro",
			Description: "Laptop para desarrollo backend",
			Price:       4500,
		},
		{
			SKU:         "LAPTOP-AIR-SEARCH-SORT",
			Name:        "Laptop Air",
			Description: "Laptop ligera",
			Price:       3500,
		},
		{
			SKU:         "MOUSE-SEARCH-SORT",
			Name:        "Mouse",
			Description: "Mouse inalámbrico",
			Price:       120,
		},
	}

	for _, input := range products {
		if _, err := store.Create(ctx, input); err != nil {
			t.Fatalf("expected no error creating product, got %v", err)
		}
	}

	result, err := store.List(ctx, ListProductsInput{
		Page:     1,
		PageSize: 2,
		Search:   "laptop",
		Sort:     SortFieldPrice,
		Order:    SortOrderDesc,
	})
	if err != nil {
		t.Fatalf("expected no error listing products, got %v", err)
	}

	if len(result.Items) != 2 {
		t.Fatalf("expected 2 products, got %d", len(result.Items))
	}

	expectedNames := []string{"Laptop Pro", "Laptop Air"}

	for index, expectedName := range expectedNames {
		if result.Items[index].Name != expectedName {
			t.Fatalf("expected product at index %d to be %q, got %q", index, expectedName, result.Items[index].Name)
		}
	}

	if result.Total != 3 {
		t.Fatalf("expected total %d, got %d", 3, result.Total)
	}

	if result.TotalPages != 2 {
		t.Fatalf("expected total pages %d, got %d", 2, result.TotalPages)
	}

	if result.Search != "laptop" {
		t.Fatalf("expected search %q, got %q", "laptop", result.Search)
	}

	if result.Sort != SortFieldPrice {
		t.Fatalf("expected sort %q, got %q", SortFieldPrice, result.Sort)
	}

	if result.Order != SortOrderDesc {
		t.Fatalf("expected order %q, got %q", SortOrderDesc, result.Order)
	}
}

func TestStoreCreateProductDuplicateSKU(t *testing.T) {
	store := NewStore()
	ctx := context.Background()

	_, err := store.Create(ctx, CreateProductInput{
		SKU:         "DUPLICATE-SKU",
		Name:        "Product 1",
		Description: "Primer producto",
		Price:       100,
	})
	if err != nil {
		t.Fatalf("expected no error creating first product, got %v", err)
	}

	_, err = store.Create(ctx, CreateProductInput{
		SKU:         "duplicate-sku",
		Name:        "Product 2",
		Description: "Segundo producto",
		Price:       200,
	})
	if !errors.Is(err, ErrSKUAlreadyExists) {
		t.Fatalf("expected ErrSKUAlreadyExists, got %v", err)
	}
}

func TestStoreUpdateProductDuplicateSKU(t *testing.T) {
	store := NewStore()
	ctx := context.Background()

	_, err := store.Create(ctx, CreateProductInput{
		SKU:         "PRODUCT-001",
		Name:        "Product 1",
		Description: "Primer producto",
		Price:       100,
	})
	if err != nil {
		t.Fatalf("expected no error creating first product, got %v", err)
	}

	second, err := store.Create(ctx, CreateProductInput{
		SKU:         "PRODUCT-002",
		Name:        "Product 2",
		Description: "Segundo producto",
		Price:       200,
	})
	if err != nil {
		t.Fatalf("expected no error creating second product, got %v", err)
	}

	_, err = store.Update(ctx, second.ID, UpdateProductInput{
		SKU:         "product-001",
		Name:        "Product 2 Updated",
		Description: "Intento de SKU duplicado",
		Price:       250,
	})
	if !errors.Is(err, ErrSKUAlreadyExists) {
		t.Fatalf("expected ErrSKUAlreadyExists, got %v", err)
	}
}

func TestStoreGetProductBySKU(t *testing.T) {
	store := NewStore()
	ctx := context.Background()

	created, err := store.Create(ctx, CreateProductInput{
		SKU:         "LAPTOP-SKU-001",
		Name:        "Laptop",
		Description: "Laptop para desarrollo backend",
		Price:       3500,
	})
	if err != nil {
		t.Fatalf("expected no error creating product, got %v", err)
	}

	found, err := store.GetBySKU(ctx, "laptop-sku-001")
	if err != nil {
		t.Fatalf("expected no error getting product by sku, got %v", err)
	}

	if found.ID != created.ID {
		t.Fatalf("expected ID %q, got %q", created.ID, found.ID)
	}

	if found.SKU != "LAPTOP-SKU-001" {
		t.Fatalf("expected SKU %q, got %q", "LAPTOP-SKU-001", found.SKU)
	}
}

func TestStoreGetProductBySKUNotFound(t *testing.T) {
	store := NewStore()
	ctx := context.Background()

	_, err := store.GetBySKU(ctx, "missing-sku")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestStoreListProductsFilterByPriceRange(t *testing.T) {
	store := NewStore()
	ctx := context.Background()

	products := []CreateProductInput{
		{
			SKU:         "PRICE-LAPTOP-BASIC",
			Name:        "Laptop Basic",
			Description: "Laptop para oficina",
			Price:       2500,
		},
		{
			SKU:         "PRICE-LAPTOP-PRO",
			Name:        "Laptop Pro",
			Description: "Laptop para desarrollo backend",
			Price:       4500,
		},
		{
			SKU:         "PRICE-LAPTOP-AIR",
			Name:        "Laptop Air",
			Description: "Laptop ligera",
			Price:       3500,
		},
		{
			SKU:         "PRICE-MOUSE",
			Name:        "Mouse",
			Description: "Mouse inalámbrico",
			Price:       120,
		},
	}

	for _, input := range products {
		if _, err := store.Create(ctx, input); err != nil {
			t.Fatalf("expected no error creating product, got %v", err)
		}
	}

	minPrice := 2000.0
	maxPrice := 4000.0

	result, err := store.List(ctx, ListProductsInput{
		Page:     1,
		PageSize: 10,
		MinPrice: &minPrice,
		MaxPrice: &maxPrice,
		Sort:     SortFieldPrice,
		Order:    SortOrderAsc,
	})
	if err != nil {
		t.Fatalf("expected no error listing products, got %v", err)
	}

	if len(result.Items) != 2 {
		t.Fatalf("expected 2 products, got %d", len(result.Items))
	}

	expectedNames := []string{
		"Laptop Basic",
		"Laptop Air",
	}

	for index, expectedName := range expectedNames {
		if result.Items[index].Name != expectedName {
			t.Fatalf("expected product at index %d to be %q, got %q", index, expectedName, result.Items[index].Name)
		}
	}

	if result.Total != 2 {
		t.Fatalf("expected total %d, got %d", 2, result.Total)
	}

	if result.MinPrice == nil || *result.MinPrice != minPrice {
		t.Fatalf("expected min price %v, got %v", minPrice, result.MinPrice)
	}

	if result.MaxPrice == nil || *result.MaxPrice != maxPrice {
		t.Fatalf("expected max price %v, got %v", maxPrice, result.MaxPrice)
	}
}

func TestStoreListProductsFilterByCreatedRange(t *testing.T) {
	store := NewStore()
	ctx := context.Background()

	_, err := store.Create(ctx, CreateProductInput{
		SKU:         "DATE-LAPTOP-BASIC",
		Name:        "Laptop Basic",
		Description: "Laptop para oficina",
		Price:       2500,
	})
	if err != nil {
		t.Fatalf("expected no error creating product, got %v", err)
	}

	_, err = store.Create(ctx, CreateProductInput{
		SKU:         "DATE-LAPTOP-PRO",
		Name:        "Laptop Pro",
		Description: "Laptop para desarrollo backend",
		Price:       4500,
	})
	if err != nil {
		t.Fatalf("expected no error creating product, got %v", err)
	}

	createdFrom := time.Now().UTC().Add(-24 * time.Hour)
	createdTo := time.Now().UTC().Add(24 * time.Hour)

	result, err := store.List(ctx, ListProductsInput{
		Page:        1,
		PageSize:    10,
		CreatedFrom: &createdFrom,
		CreatedTo:   &createdTo,
		Sort:        SortFieldCreatedAt,
		Order:       SortOrderAsc,
	})
	if err != nil {
		t.Fatalf("expected no error listing products, got %v", err)
	}

	if len(result.Items) != 2 {
		t.Fatalf("expected 2 products, got %d", len(result.Items))
	}

	if result.Total != 2 {
		t.Fatalf("expected total %d, got %d", 2, result.Total)
	}

	if result.CreatedFrom == nil || !result.CreatedFrom.Equal(createdFrom) {
		t.Fatalf("expected created from %v, got %v", createdFrom, result.CreatedFrom)
	}

	if result.CreatedTo == nil || !result.CreatedTo.Equal(createdTo) {
		t.Fatalf("expected created to %v, got %v", createdTo, result.CreatedTo)
	}
}

func TestStoreCreateProductAllowsReusingSKUAfterSoftDelete(t *testing.T) {
	store := NewStore()
	ctx := context.Background()

	first, err := store.Create(ctx, CreateProductInput{
		SKU:         "REUSABLE-SKU",
		Name:        "First Product",
		Description: "Primer producto",
		Price:       100,
	})
	if err != nil {
		t.Fatalf("expected no error creating first product, got %v", err)
	}

	if err := store.Delete(ctx, first.ID); err != nil {
		t.Fatalf("expected no error soft deleting first product, got %v", err)
	}

	second, err := store.Create(ctx, CreateProductInput{
		SKU:         "reusable-sku",
		Name:        "Second Product",
		Description: "Segundo producto",
		Price:       200,
	})
	if err != nil {
		t.Fatalf("expected no error reusing sku after soft delete, got %v", err)
	}

	if second.ID == first.ID {
		t.Fatalf("expected different product IDs, got same ID %q", second.ID)
	}

	if second.SKU != "reusable-sku" && second.SKU != "REUSABLE-SKU" {
		t.Fatalf("expected reused sku, got %q", second.SKU)
	}
}

func TestStoreRestoreProduct(t *testing.T) {
	store := NewStore()
	ctx := context.Background()

	created, err := store.Create(ctx, CreateProductInput{
		SKU:         "STORE-RESTORE-001",
		Name:        "Store Restore",
		Description: "Producto para restaurar",
		Price:       100,
	})
	if err != nil {
		t.Fatalf("expected no error creating product, got %v", err)
	}

	if err := store.Delete(ctx, created.ID); err != nil {
		t.Fatalf("expected no error soft deleting product, got %v", err)
	}

	restored, err := store.Restore(ctx, created.ID)
	if err != nil {
		t.Fatalf("expected no error restoring product, got %v", err)
	}

	if restored.ID != created.ID {
		t.Fatalf("expected restored ID %q, got %q", created.ID, restored.ID)
	}

	if restored.DeletedAt != nil {
		t.Fatal("expected restored product DeletedAt to be nil")
	}

	found, err := store.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("expected no error getting restored product, got %v", err)
	}

	if found.ID != created.ID {
		t.Fatalf("expected found ID %q, got %q", created.ID, found.ID)
	}
}

func TestStoreRestoreProductDuplicateSKU(t *testing.T) {
	store := NewStore()
	ctx := context.Background()

	first, err := store.Create(ctx, CreateProductInput{
		SKU:         "STORE-RESTORE-CONFLICT",
		Name:        "First Product",
		Description: "Primer producto",
		Price:       100,
	})
	if err != nil {
		t.Fatalf("expected no error creating first product, got %v", err)
	}

	if err := store.Delete(ctx, first.ID); err != nil {
		t.Fatalf("expected no error soft deleting first product, got %v", err)
	}

	_, err = store.Create(ctx, CreateProductInput{
		SKU:         "store-restore-conflict",
		Name:        "Second Product",
		Description: "Segundo producto activo",
		Price:       200,
	})
	if err != nil {
		t.Fatalf("expected no error creating second product, got %v", err)
	}

	_, err = store.Restore(ctx, first.ID)
	if !errors.Is(err, ErrSKUAlreadyExists) {
		t.Fatalf("expected ErrSKUAlreadyExists, got %v", err)
	}
}

func TestStoreListDeletedProducts(t *testing.T) {
	store := NewStore()
	ctx := context.Background()

	active, err := store.Create(ctx, CreateProductInput{
		SKU:         "STORE-ACTIVE-001",
		Name:        "Active Product",
		Description: "Producto activo",
		Price:       100,
	})
	if err != nil {
		t.Fatalf("expected no error creating active product, got %v", err)
	}

	deleted, err := store.Create(ctx, CreateProductInput{
		SKU:         "STORE-DELETED-001",
		Name:        "Deleted Product",
		Description: "Producto eliminado",
		Price:       200,
	})
	if err != nil {
		t.Fatalf("expected no error creating deleted product, got %v", err)
	}

	if err := store.Delete(ctx, deleted.ID); err != nil {
		t.Fatalf("expected no error deleting product, got %v", err)
	}

	result, err := store.ListDeleted(ctx, ListProductsInput{
		Page:     1,
		PageSize: 10,
		Sort:     SortFieldID,
		Order:    SortOrderAsc,
	})
	if err != nil {
		t.Fatalf("expected no error listing deleted products, got %v", err)
	}

	if result.Total != 1 {
		t.Fatalf("expected total %d, got %d", 1, result.Total)
	}

	if len(result.Items) != 1 {
		t.Fatalf("expected 1 deleted product, got %d", len(result.Items))
	}

	if result.Items[0].ID != deleted.ID {
		t.Fatalf("expected deleted product ID %q, got %q", deleted.ID, result.Items[0].ID)
	}

	if result.Items[0].ID == active.ID {
		t.Fatal("expected active product to not be listed as deleted")
	}

	if result.Items[0].DeletedAt == nil {
		t.Fatal("expected deleted_at to be set")
	}
}
