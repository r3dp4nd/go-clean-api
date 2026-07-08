//go:build integration

package product

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestIntegrationPostgresProductRepositoryCRUD(t *testing.T) {
	repository := newPostgresIntegrationRepository(t)

	ctx := context.Background()

	created, err := repository.Create(ctx, CreateProductInput{
		SKU:         "LAPTOP-INTEGRATION",
		Name:        "Laptop Integration",
		Description: "Producto creado desde test de integración",
		Price:       3500,
	})
	if err != nil {
		t.Fatalf("expected no error creating product, got %v", err)
	}

	if created.ID == "" {
		t.Fatal("expected product ID to be generated")
	}

	if created.SKU != "LAPTOP-INTEGRATION" {
		t.Fatalf("expected product sku %q, got %q", "LAPTOP-INTEGRATION", created.SKU)
	}

	if created.Name != "Laptop Integration" {
		t.Fatalf("expected product name %q, got %q", "Laptop Integration", created.Name)
	}

	if created.CreatedAt.IsZero() {
		t.Fatal("expected created_at to be set")
	}

	if created.UpdatedAt.IsZero() {
		t.Fatal("expected updated_at to be set")
	}

	found, err := repository.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("expected no error getting product, got %v", err)
	}

	if found.ID != created.ID {
		t.Fatalf("expected ID %q, got %q", created.ID, found.ID)
	}

	if found.SKU != created.SKU {
		t.Fatalf("expected SKU %q, got %q", created.SKU, found.SKU)
	}

	updated, err := repository.Update(ctx, created.ID, UpdateProductInput{
		SKU:         "LAPTOP-INTEGRATION-PRO",
		Name:        "Laptop Integration Pro",
		Description: "Producto actualizado desde test de integración",
		Price:       4200,
	})
	if err != nil {
		t.Fatalf("expected no error updating product, got %v", err)
	}

	if updated.SKU != "LAPTOP-INTEGRATION-PRO" {
		t.Fatalf("expected updated sku %q, got %q", "LAPTOP-INTEGRATION-PRO", updated.SKU)
	}

	if updated.Name != "Laptop Integration Pro" {
		t.Fatalf("expected updated name %q, got %q", "Laptop Integration Pro", updated.Name)
	}

	if updated.Price != 4200 {
		t.Fatalf("expected updated price %v, got %v", 4200.0, updated.Price)
	}

	listResult, err := repository.List(ctx, ListProductsInput{
		Page:     1,
		PageSize: 10,
		Sort:     SortFieldID,
		Order:    SortOrderAsc,
	})
	if err != nil {
		t.Fatalf("expected no error listing products, got %v", err)
	}

	if listResult.Total != 1 {
		t.Fatalf("expected total %d, got %d", 1, listResult.Total)
	}

	if len(listResult.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(listResult.Items))
	}

	if listResult.Items[0].ID != created.ID {
		t.Fatalf("expected listed product ID %q, got %q", created.ID, listResult.Items[0].ID)
	}

	if listResult.Items[0].SKU != "LAPTOP-INTEGRATION-PRO" {
		t.Fatalf("expected listed product SKU %q, got %q", "LAPTOP-INTEGRATION-PRO", listResult.Items[0].SKU)
	}

	if err := repository.Delete(ctx, created.ID); err != nil {
		t.Fatalf("expected no error deleting product, got %v", err)
	}

	_, err = repository.Get(ctx, created.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestIntegrationPostgresProductRepositoryListSearchSortAndPagination(t *testing.T) {
	repository := newPostgresIntegrationRepository(t)

	ctx := context.Background()

	products := []CreateProductInput{
		{
			SKU:         "LAPTOP-BASIC-IT",
			Name:        "Laptop Basic",
			Description: "Laptop para oficina",
			Price:       2500,
		},
		{
			SKU:         "LAPTOP-PRO-IT",
			Name:        "Laptop Pro",
			Description: "Laptop para desarrollo backend",
			Price:       4500,
		},
		{
			SKU:         "LAPTOP-AIR-IT",
			Name:        "Laptop Air",
			Description: "Laptop ligera",
			Price:       3500,
		},
		{
			SKU:         "MOUSE-IT",
			Name:        "Mouse",
			Description: "Mouse inalámbrico",
			Price:       120,
		},
	}

	for _, input := range products {
		if _, err := repository.Create(ctx, input); err != nil {
			t.Fatalf("expected no error creating product %q, got %v", input.Name, err)
		}
	}

	result, err := repository.List(ctx, ListProductsInput{
		Page:     1,
		PageSize: 2,
		Search:   "laptop",
		Sort:     SortFieldPrice,
		Order:    SortOrderDesc,
	})
	if err != nil {
		t.Fatalf("expected no error listing products, got %v", err)
	}

	if result.Total != 3 {
		t.Fatalf("expected total %d, got %d", 3, result.Total)
	}

	if result.TotalPages != 2 {
		t.Fatalf("expected total pages %d, got %d", 2, result.TotalPages)
	}

	if len(result.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result.Items))
	}

	expectedNames := []string{
		"Laptop Pro",
		"Laptop Air",
	}

	for index, expectedName := range expectedNames {
		if result.Items[index].Name != expectedName {
			t.Fatalf(
				"expected product at index %d to be %q, got %q",
				index,
				expectedName,
				result.Items[index].Name,
			)
		}
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

func TestIntegrationPostgresProductRepositoryNotFound(t *testing.T) {
	repository := newPostgresIntegrationRepository(t)

	ctx := context.Background()

	missingID := "00000000-0000-0000-0000-000000000000"

	_, err := repository.Get(ctx, missingID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound getting missing product, got %v", err)
	}

	_, err = repository.Update(ctx, missingID, UpdateProductInput{
		SKU:         "MISSING-PRODUCT",
		Name:        "Missing",
		Description: "Missing product",
		Price:       10,
	})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound updating missing product, got %v", err)
	}

	err = repository.Delete(ctx, missingID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound deleting missing product, got %v", err)
	}
}

func newPostgresIntegrationRepository(t *testing.T) *PostgresRepository {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}

	if dsn == "" {
		t.Skip("TEST_DATABASE_URL or DATABASE_URL is required for integration tests")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to create postgres pool: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("failed to ping postgres: %v", err)
	}

	truncateProducts(t, pool)

	t.Cleanup(func() {
		truncateProducts(t, pool)
		pool.Close()
	})

	return NewPostgresRepository(pool)
}

func truncateProducts(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := pool.Exec(ctx, "TRUNCATE TABLE products"); err != nil {
		t.Fatalf("failed to truncate products table: %v", err)
	}
}
