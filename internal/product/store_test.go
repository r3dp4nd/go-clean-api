package product

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
)

func TestStoreCreateProduct(t *testing.T) {
	store := NewStore()
	ctx := context.Background()

	item, err := store.Create(ctx, CreateProductInput{
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
		Name:        "Mouse",
		Description: "Mouse inalámbrico",
		Price:       120,
	})
	if err != nil {
		t.Fatalf("expected no error creating mouse, got %v", err)
	}

	_, err = store.Create(ctx, CreateProductInput{
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
		Name:        "Laptop",
		Description: "Laptop básica",
		Price:       3000,
	})
	if err != nil {
		t.Fatalf("expected no error creating product, got %v", err)
	}

	updated, err := store.Update(ctx, created.ID, UpdateProductInput{
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
		Name:        "Canceled",
		Description: "No debe crearse",
		Price:       10,
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled on Create, got %v", err)
	}

	_, err = store.Update(ctx, "1", UpdateProductInput{
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
