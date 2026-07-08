package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/r3dp4nd/go-clean-api/internal/product"
)

func TestProductsCRUD(t *testing.T) {
	handler := newTestHTTPHandler()

	createBody := []byte(`{
		"name": "Laptop",
		"description": "Laptop para desarrollo backend",
		"price": 3500
	}`)

	createRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/products",
		bytes.NewReader(createBody),
	)
	createRequest.Header.Set("Content-Type", "application/json")
	createRequest.Header.Set(requestIDHeader, "test-create-product")

	createRecorder := httptest.NewRecorder()

	handler.ServeHTTP(createRecorder, createRequest)

	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d. body: %s", http.StatusCreated, createRecorder.Code, createRecorder.Body.String())
	}

	var createdProduct ProductResponse
	if err := json.NewDecoder(createRecorder.Body).Decode(&createdProduct); err != nil {
		t.Fatalf("failed to decode created product: %v", err)
	}

	if createdProduct.ID == "" {
		t.Fatal("expected product id to be generated")
	}

	if createdProduct.Name != "Laptop" {
		t.Fatalf("expected product name %q, got %q", "Laptop", createdProduct.Name)
	}

	if createdProduct.Price != 3500 {
		t.Fatalf("expected product price %v, got %v", 3500.0, createdProduct.Price)
	}

	listRequest := httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)
	listRecorder := httptest.NewRecorder()

	handler.ServeHTTP(listRecorder, listRequest)

	if listRecorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listRecorder.Code)
	}

	var listResponse ProductListResponse
	if err := json.NewDecoder(listRecorder.Body).Decode(&listResponse); err != nil {
		t.Fatalf("failed to decode product list: %v", err)
	}

	if len(listResponse.Data) != 1 {
		t.Fatalf("expected 1 product, got %d", len(listResponse.Data))
	}

	if listResponse.Meta.Total != 1 {
		t.Fatalf("expected total %d, got %d", 1, listResponse.Meta.Total)
	}

	if listResponse.Meta.Page != 1 {
		t.Fatalf("expected page %d, got %d", 1, listResponse.Meta.Page)
	}

	if listResponse.Meta.PageSize != 10 {
		t.Fatalf("expected page size %d, got %d", 10, listResponse.Meta.PageSize)
	}

	getRequest := httptest.NewRequest(http.MethodGet, "/api/v1/products/"+createdProduct.ID, nil)
	getRecorder := httptest.NewRecorder()

	handler.ServeHTTP(getRecorder, getRequest)

	if getRecorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, getRecorder.Code)
	}

	var fetchedProduct ProductResponse
	if err := json.NewDecoder(getRecorder.Body).Decode(&fetchedProduct); err != nil {
		t.Fatalf("failed to decode fetched product: %v", err)
	}

	if fetchedProduct.ID != createdProduct.ID {
		t.Fatalf("expected product id %q, got %q", createdProduct.ID, fetchedProduct.ID)
	}

	updateBody := []byte(`{
		"name": "Laptop Pro",
		"description": "Laptop para Go, Docker y Kubernetes",
		"price": 4200
	}`)

	updateRequest := httptest.NewRequest(
		http.MethodPut,
		"/api/v1/products/"+createdProduct.ID,
		bytes.NewReader(updateBody),
	)
	updateRequest.Header.Set("Content-Type", "application/json")

	updateRecorder := httptest.NewRecorder()

	handler.ServeHTTP(updateRecorder, updateRequest)

	if updateRecorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d. body: %s", http.StatusOK, updateRecorder.Code, updateRecorder.Body.String())
	}

	var updatedProduct ProductResponse
	if err := json.NewDecoder(updateRecorder.Body).Decode(&updatedProduct); err != nil {
		t.Fatalf("failed to decode updated product: %v", err)
	}

	if updatedProduct.Name != "Laptop Pro" {
		t.Fatalf("expected product name %q, got %q", "Laptop Pro", updatedProduct.Name)
	}

	if updatedProduct.Price != 4200 {
		t.Fatalf("expected product price %v, got %v", 4200.0, updatedProduct.Price)
	}

	deleteRequest := httptest.NewRequest(http.MethodDelete, "/api/v1/products/"+createdProduct.ID, nil)
	deleteRecorder := httptest.NewRecorder()

	handler.ServeHTTP(deleteRecorder, deleteRequest)

	if deleteRecorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, deleteRecorder.Code)
	}

	getDeletedRequest := httptest.NewRequest(http.MethodGet, "/api/v1/products/"+createdProduct.ID, nil)
	getDeletedRecorder := httptest.NewRecorder()

	handler.ServeHTTP(getDeletedRecorder, getDeletedRequest)

	if getDeletedRecorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, getDeletedRecorder.Code)
	}
}

func TestCreateProductValidationError(t *testing.T) {
	handler := newTestHTTPHandler()

	body := []byte(`{
		"name": "",
		"description": "Producto inválido",
		"price": -10
	}`)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/products", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set(requestIDHeader, "invalid-product-data")

	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status %d, got %d", http.StatusUnprocessableEntity, responseRecorder.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(responseRecorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if response.Error.Code != errorCodeValidation {
		t.Fatalf("expected error code %q, got %q", errorCodeValidation, response.Error.Code)
	}

	if response.Error.Message != "validation failed" {
		t.Fatalf("expected error message %q, got %q", "validation failed", response.Error.Message)
	}

	if response.Error.RequestID != "invalid-product-data" {
		t.Fatalf("expected request id %q, got %q", "invalid-product-data", response.Error.RequestID)
	}

	if len(response.Error.Fields) != 2 {
		t.Fatalf("expected 2 field errors, got %d", len(response.Error.Fields))
	}

	expectedFields := map[string]string{
		"name":  "name is required",
		"price": "price must be greater than or equal to zero",
	}

	for _, field := range response.Error.Fields {
		expectedMessage, ok := expectedFields[field.Field]
		if !ok {
			t.Fatalf("unexpected field error: %s", field.Field)
		}

		if field.Message != expectedMessage {
			t.Fatalf("expected message %q for field %q, got %q", expectedMessage, field.Field, field.Message)
		}
	}
}

func TestCreateProductInvalidJSON(t *testing.T) {
	handler := newTestHTTPHandler()

	body := []byte(`{"name": "Laptop",`)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/products", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set(requestIDHeader, "invalid-json")

	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, responseRecorder.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(responseRecorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if response.Error.Code != errorCodeInvalidRequest {
		t.Fatalf("expected error code %q, got %q", errorCodeInvalidRequest, response.Error.Code)
	}
}

func TestGetProductNotFound(t *testing.T) {
	handler := newTestHTTPHandler()

	request := httptest.NewRequest(http.MethodGet, "/api/v1/products/999", nil)
	request.Header.Set(requestIDHeader, "product-not-found")

	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, responseRecorder.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(responseRecorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if response.Error.Code != errorCodeNotFound {
		t.Fatalf("expected error code %q, got %q", errorCodeNotFound, response.Error.Code)
	}

	if response.Error.Message != "product not found" {
		t.Fatalf("expected error message %q, got %q", "product not found", response.Error.Message)
	}

	if response.Error.RequestID != "product-not-found" {
		t.Fatalf("expected request id %q, got %q", "product-not-found", response.Error.RequestID)
	}
}

func TestProductsMethodNotAllowed(t *testing.T) {
	handler := newTestHTTPHandler()

	request := httptest.NewRequest(http.MethodPatch, "/api/v1/products", nil)
	request.Header.Set(requestIDHeader, "products-405")

	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, responseRecorder.Code)
	}

	if got := responseRecorder.Header().Get("Allow"); got != "GET, POST" {
		t.Fatalf("expected Allow header %q, got %q", "GET, POST", got)
	}
}

func TestGetProductWithEmptyIDReturnsNotFound(t *testing.T) {
	handler := newTestHTTPHandler()

	request := httptest.NewRequest(http.MethodGet, "/api/v1/products/", nil)
	request.Header.Set(requestIDHeader, "empty-product-id")

	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, responseRecorder.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(responseRecorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if response.Error.Code != errorCodeNotFound {
		t.Fatalf("expected error code %q, got %q", errorCodeNotFound, response.Error.Code)
	}

	if response.Error.RequestID != "empty-product-id" {
		t.Fatalf("expected request id %q, got %q", "empty-product-id", response.Error.RequestID)
	}
}

func TestGetProductWithNestedPathReturnsNotFound(t *testing.T) {
	handler := newTestHTTPHandler()

	request := httptest.NewRequest(http.MethodGet, "/api/v1/products/1/details", nil)
	request.Header.Set(requestIDHeader, "nested-product-path")

	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, responseRecorder.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(responseRecorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if response.Error.Code != errorCodeNotFound {
		t.Fatalf("expected error code %q, got %q", errorCodeNotFound, response.Error.Code)
	}
}

func TestListProductsPagination(t *testing.T) {
	handler := newTestHTTPHandler()

	for i := 1; i <= 3; i++ {
		body := []byte(`{
			"name": "Product",
			"description": "Producto de prueba",
			"price": 100
		}`)

		request := httptest.NewRequest(
			http.MethodPost,
			"/api/v1/products",
			bytes.NewReader(body),
		)
		request.Header.Set("Content-Type", "application/json")

		responseRecorder := httptest.NewRecorder()

		handler.ServeHTTP(responseRecorder, request)

		if responseRecorder.Code != http.StatusCreated {
			t.Fatalf("expected status %d, got %d", http.StatusCreated, responseRecorder.Code)
		}
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/products?page=2&page_size=2", nil)
	request.Header.Set(requestIDHeader, "pagination-test")

	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, responseRecorder.Code)
	}

	var response ProductListResponse
	if err := json.NewDecoder(responseRecorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode product list response: %v", err)
	}

	if len(response.Data) != 1 {
		t.Fatalf("expected 1 product on page 2, got %d", len(response.Data))
	}

	if response.Meta.Page != 2 {
		t.Fatalf("expected page %d, got %d", 2, response.Meta.Page)
	}

	if response.Meta.PageSize != 2 {
		t.Fatalf("expected page size %d, got %d", 2, response.Meta.PageSize)
	}

	if response.Meta.Total != 3 {
		t.Fatalf("expected total %d, got %d", 3, response.Meta.Total)
	}

	if response.Meta.TotalPages != 2 {
		t.Fatalf("expected total pages %d, got %d", 2, response.Meta.TotalPages)
	}
}

func TestListProductsSearch(t *testing.T) {
	handler := newTestHTTPHandler()

	products := []string{
		`{"name":"Laptop","description":"Equipo para desarrollo backend","price":3500}`,
		`{"name":"Mouse","description":"Mouse inalámbrico","price":120}`,
		`{"name":"Keyboard","description":"Teclado mecánico","price":250}`,
	}

	for _, body := range products {
		request := httptest.NewRequest(
			http.MethodPost,
			"/api/v1/products",
			bytes.NewReader([]byte(body)),
		)
		request.Header.Set("Content-Type", "application/json")

		responseRecorder := httptest.NewRecorder()

		handler.ServeHTTP(responseRecorder, request)

		if responseRecorder.Code != http.StatusCreated {
			t.Fatalf("expected status %d, got %d", http.StatusCreated, responseRecorder.Code)
		}
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/products?search=laptop&page=1&page_size=10", nil)
	request.Header.Set(requestIDHeader, "search-products")

	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, responseRecorder.Code)
	}

	var response ProductListResponse
	if err := json.NewDecoder(responseRecorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode product list response: %v", err)
	}

	if len(response.Data) != 1 {
		t.Fatalf("expected 1 product, got %d", len(response.Data))
	}

	if response.Data[0].Name != "Laptop" {
		t.Fatalf("expected product %q, got %q", "Laptop", response.Data[0].Name)
	}

	if response.Meta.Total != 1 {
		t.Fatalf("expected total %d, got %d", 1, response.Meta.Total)
	}

	if response.Meta.Search != "laptop" {
		t.Fatalf("expected search %q, got %q", "laptop", response.Meta.Search)
	}
}

func TestListProductsSearchByDescription(t *testing.T) {
	handler := newTestHTTPHandler()

	body := []byte(`{
		"name": "Laptop",
		"description": "Equipo para desarrollo backend",
		"price": 3500
	}`)

	createRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/products",
		bytes.NewReader(body),
	)
	createRequest.Header.Set("Content-Type", "application/json")

	createRecorder := httptest.NewRecorder()

	handler.ServeHTTP(createRecorder, createRequest)

	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createRecorder.Code)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/products?search=backend", nil)

	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, responseRecorder.Code)
	}

	var response ProductListResponse
	if err := json.NewDecoder(responseRecorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode product list response: %v", err)
	}

	if len(response.Data) != 1 {
		t.Fatalf("expected 1 product, got %d", len(response.Data))
	}

	if response.Data[0].Name != "Laptop" {
		t.Fatalf("expected product %q, got %q", "Laptop", response.Data[0].Name)
	}
}

func TestListProductsInvalidSearch(t *testing.T) {
	handler := newTestHTTPHandler()

	longSearch := strings.Repeat("a", product.MaxSearchLength+1)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/products?search="+longSearch, nil)
	request.Header.Set(requestIDHeader, "invalid-search")

	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status %d, got %d", http.StatusUnprocessableEntity, responseRecorder.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(responseRecorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if response.Error.Code != errorCodeValidation {
		t.Fatalf("expected error code %q, got %q", errorCodeValidation, response.Error.Code)
	}

	if len(response.Error.Fields) != 1 {
		t.Fatalf("expected 1 field error, got %d", len(response.Error.Fields))
	}

	if response.Error.Fields[0].Field != "search" {
		t.Fatalf("expected field %q, got %q", "search", response.Error.Fields[0].Field)
	}
}
