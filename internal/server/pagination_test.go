package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/r3dp4nd/go-clean-api/internal/product"
)

func TestReadProductListQueryDefaults(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)

	input, fields := readProductListQuery(request)

	if len(fields) != 0 {
		t.Fatalf("expected no validation fields, got %d", len(fields))
	}

	if input.Page != product.DefaultPage {
		t.Fatalf("expected page %d, got %d", product.DefaultPage, input.Page)
	}

	if input.PageSize != product.DefaultPageSize {
		t.Fatalf("expected page size %d, got %d", product.DefaultPageSize, input.PageSize)
	}

	if input.Search != "" {
		t.Fatalf("expected empty search, got %q", input.Search)
	}
}

func TestReadProductListQueryValidValues(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/products?page=2&page_size=25&search=laptop", nil)

	input, fields := readProductListQuery(request)

	if len(fields) != 0 {
		t.Fatalf("expected no validation fields, got %d", len(fields))
	}

	if input.Page != 2 {
		t.Fatalf("expected page %d, got %d", 2, input.Page)
	}

	if input.PageSize != 25 {
		t.Fatalf("expected page size %d, got %d", 25, input.PageSize)
	}

	if input.Search != "laptop" {
		t.Fatalf("expected search %q, got %q", "laptop", input.Search)
	}
}

func TestReadProductListQueryTrimsSearch(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/products?search=%20%20laptop%20%20", nil)

	input, fields := readProductListQuery(request)

	if len(fields) != 0 {
		t.Fatalf("expected no validation fields, got %d", len(fields))
	}

	if input.Search != "laptop" {
		t.Fatalf("expected trimmed search %q, got %q", "laptop", input.Search)
	}
}

func TestReadProductListQueryInvalidPaginationValues(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/products?page=abc&page_size=0", nil)

	_, fields := readProductListQuery(request)

	if len(fields) != 2 {
		t.Fatalf("expected 2 validation fields, got %d", len(fields))
	}
}

func TestReadProductListQueryPageSizeTooLarge(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/products?page_size=101", nil)

	_, fields := readProductListQuery(request)

	if len(fields) != 1 {
		t.Fatalf("expected 1 validation field, got %d", len(fields))
	}

	if fields[0].Field != "page_size" {
		t.Fatalf("expected field %q, got %q", "page_size", fields[0].Field)
	}
}

func TestReadProductListQuerySearchTooLarge(t *testing.T) {
	longSearch := strings.Repeat("a", product.MaxSearchLength+1)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/products?search="+longSearch, nil)

	_, fields := readProductListQuery(request)

	if len(fields) != 1 {
		t.Fatalf("expected 1 validation field, got %d", len(fields))
	}

	if fields[0].Field != "search" {
		t.Fatalf("expected field %q, got %q", "search", fields[0].Field)
	}
}
