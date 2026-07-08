package server

import (
	"strings"
	"testing"
)

func TestValidateCreateProductRequestValid(t *testing.T) {
	fields := validateCreateProductRequest(CreateProductRequest{
		Name:        "Laptop",
		Description: "Laptop para desarrollo backend",
		Price:       3500,
	})

	if len(fields) != 0 {
		t.Fatalf("expected no validation errors, got %d", len(fields))
	}
}

func TestValidateCreateProductRequestRequiredName(t *testing.T) {
	fields := validateCreateProductRequest(CreateProductRequest{
		Name:        "   ",
		Description: "Producto sin nombre",
		Price:       100,
	})

	if len(fields) != 1 {
		t.Fatalf("expected 1 validation error, got %d", len(fields))
	}

	if fields[0].Field != "name" {
		t.Fatalf("expected field %q, got %q", "name", fields[0].Field)
	}

	if fields[0].Message != "name is required" {
		t.Fatalf("expected message %q, got %q", "name is required", fields[0].Message)
	}
}

func TestValidateCreateProductRequestNegativePrice(t *testing.T) {
	fields := validateCreateProductRequest(CreateProductRequest{
		Name:        "Producto inválido",
		Description: "Precio negativo",
		Price:       -1,
	})

	if len(fields) != 1 {
		t.Fatalf("expected 1 validation error, got %d", len(fields))
	}

	if fields[0].Field != "price" {
		t.Fatalf("expected field %q, got %q", "price", fields[0].Field)
	}
}

func TestValidateCreateProductRequestMaxLengths(t *testing.T) {
	fields := validateCreateProductRequest(CreateProductRequest{
		Name:        strings.Repeat("a", maxProductNameLength+1),
		Description: strings.Repeat("b", maxProductDescriptionLength+1),
		Price:       100,
	})

	if len(fields) != 2 {
		t.Fatalf("expected 2 validation errors, got %d", len(fields))
	}

	expectedFields := map[string]bool{
		"name":        false,
		"description": false,
	}

	for _, field := range fields {
		if _, ok := expectedFields[field.Field]; !ok {
			t.Fatalf("unexpected field %q", field.Field)
		}

		expectedFields[field.Field] = true
	}

	for field, found := range expectedFields {
		if !found {
			t.Fatalf("expected validation error for field %q", field)
		}
	}
}
