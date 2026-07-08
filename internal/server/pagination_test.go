package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/r3dp4nd/go-clean-api/internal/product"
)

func TestReadProductPaginationQueryDefaults(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)

	input, fields := readProductPaginationQuery(request)

	if len(fields) != 0 {
		t.Fatalf("expected no validation fields, got %d", len(fields))
	}

	if input.Page != product.DefaultPage {
		t.Fatalf("expected page %d, got %d", product.DefaultPage, input.Page)
	}

	if input.PageSize != product.DefaultPageSize {
		t.Fatalf("expected page size %d, got %d", product.DefaultPageSize, input.PageSize)
	}
}

func TestReadProductPaginationQueryValidValues(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/products?page=2&page_size=25", nil)

	input, fields := readProductPaginationQuery(request)

	if len(fields) != 0 {
		t.Fatalf("expected no validation fields, got %d", len(fields))
	}

	if input.Page != 2 {
		t.Fatalf("expected page %d, got %d", 2, input.Page)
	}

	if input.PageSize != 25 {
		t.Fatalf("expected page size %d, got %d", 25, input.PageSize)
	}
}

func TestReadProductPaginationQueryInvalidValues(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/products?page=abc&page_size=0", nil)

	_, fields := readProductPaginationQuery(request)

	if len(fields) != 2 {
		t.Fatalf("expected 2 validation fields, got %d", len(fields))
	}
}

func TestReadProductPaginationQueryPageSizeTooLarge(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/products?page_size=101", nil)

	_, fields := readProductPaginationQuery(request)

	if len(fields) != 1 {
		t.Fatalf("expected 1 validation field, got %d", len(fields))
	}

	if fields[0].Field != "page_size" {
		t.Fatalf("expected field %q, got %q", "page_size", fields[0].Field)
	}
}
