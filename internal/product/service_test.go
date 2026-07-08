package product

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestServiceCreateProduct(t *testing.T) {
	store := NewStore()
	service := NewService(store)

	item, err := service.Create(context.Background(), CreateProductInput{
		Name:        "  Laptop  ",
		Description: "  Laptop para desarrollo backend  ",
		Price:       3500,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if item.Name != "Laptop" {
		t.Fatalf("expected trimmed name %q, got %q", "Laptop", item.Name)
	}

	if item.Description != "Laptop para desarrollo backend" {
		t.Fatalf("expected trimmed description, got %q", item.Description)
	}
}

func TestServiceCreateProductValidationError(t *testing.T) {
	store := NewStore()
	service := NewService(store)

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

func TestServiceUpdateProduct(t *testing.T) {
	store := NewStore()
	service := NewService(store)

	created, err := service.Create(context.Background(), CreateProductInput{
		Name:        "Laptop",
		Description: "Laptop básica",
		Price:       3000,
	})
	if err != nil {
		t.Fatalf("expected no error creating product, got %v", err)
	}

	updated, err := service.Update(context.Background(), created.ID, UpdateProductInput{
		Name:        "  Laptop Pro  ",
		Description: "  Laptop para Go  ",
		Price:       4200,
	})
	if err != nil {
		t.Fatalf("expected no error updating product, got %v", err)
	}

	if updated.Name != "Laptop Pro" {
		t.Fatalf("expected trimmed name %q, got %q", "Laptop Pro", updated.Name)
	}

	if updated.Description != "Laptop para Go" {
		t.Fatalf("expected trimmed description %q, got %q", "Laptop para Go", updated.Description)
	}
}

func TestServiceGetProductNotFound(t *testing.T) {
	store := NewStore()
	service := NewService(store)

	_, err := service.Get(context.Background(), "999")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestServiceDeleteProductNotFound(t *testing.T) {
	store := NewStore()
	service := NewService(store)

	err := service.Delete(context.Background(), "999")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestServiceCreateProductMaxLengthValidation(t *testing.T) {
	store := NewStore()
	service := NewService(store)

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

	if len(validationErr.Fields) != 2 {
		t.Fatalf("expected 2 field violations, got %d", len(validationErr.Fields))
	}
}
