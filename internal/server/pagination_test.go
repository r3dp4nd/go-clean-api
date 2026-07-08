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

	if input.Sort != product.DefaultSort {
		t.Fatalf("expected sort %q, got %q", product.DefaultSort, input.Sort)
	}

	if input.Order != product.DefaultOrder {
		t.Fatalf("expected order %q, got %q", product.DefaultOrder, input.Order)
	}
}

func TestReadProductListQueryValidValues(t *testing.T) {
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/products?page=2&page_size=25&search=laptop&sort=price&order=desc",
		nil,
	)

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

	if input.Sort != product.SortFieldPrice {
		t.Fatalf("expected sort %q, got %q", product.SortFieldPrice, input.Sort)
	}

	if input.Order != product.SortOrderDesc {
		t.Fatalf("expected order %q, got %q", product.SortOrderDesc, input.Order)
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

func TestReadProductListQueryNormalizesSortAndOrder(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/products?sort=NAME&order=DESC", nil)

	input, fields := readProductListQuery(request)

	if len(fields) != 0 {
		t.Fatalf("expected no validation fields, got %d", len(fields))
	}

	if input.Sort != product.SortFieldName {
		t.Fatalf("expected sort %q, got %q", product.SortFieldName, input.Sort)
	}

	if input.Order != product.SortOrderDesc {
		t.Fatalf("expected order %q, got %q", product.SortOrderDesc, input.Order)
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

func TestReadProductListQueryInvalidSortAndOrder(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/products?sort=unknown&order=random", nil)

	_, fields := readProductListQuery(request)

	if len(fields) != 2 {
		t.Fatalf("expected 2 validation fields, got %d", len(fields))
	}

	expectedFields := map[string]bool{
		"sort":  false,
		"order": false,
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
