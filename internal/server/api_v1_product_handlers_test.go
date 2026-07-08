package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/r3dp4nd/go-clean-api/internal/product"
)

func TestProductsCRUD(t *testing.T) {
	handler := newTestHTTPHandler()

	createBody := []byte(`{
		"sku": "LAPTOP-IT",
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

	if listResponse.Meta.Sort != product.DefaultSort {
		t.Fatalf("expected sort %q, got %q", product.DefaultSort, listResponse.Meta.Sort)
	}

	if listResponse.Meta.Order != product.DefaultOrder {
		t.Fatalf("expected order %q, got %q", product.DefaultOrder, listResponse.Meta.Order)
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
		"sku": "LAPTOP-PRO-IT",
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

	if updatedProduct.SKU != "LAPTOP-PRO-IT" {
		t.Fatalf("expected product sku %q, got %q", "LAPTOP-PRO-IT", updatedProduct.SKU)
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

	if len(response.Error.Fields) != 3 {
		t.Fatalf("expected 3 field errors, got %d", len(response.Error.Fields))
	}

	expectedFields := map[string]string{
		"sku":   "sku is required",
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
		body := []byte(fmt.Sprintf(`{
			"sku": "PRODUCT-IT-%03d",
			"name": "Product",
			"description": "Producto de prueba",
			"price": 100
		}`, i))

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
		`{"sku": "LAPTOP-IT", "name":"Laptop","description":"Equipo para desarrollo backend","price":3500}`,
		`{"sku": "MOUSE-IT", "name":"Mouse","description":"Mouse inalámbrico","price":120}`,
		`{"sku": "KEYBOARD-IT", "name":"Keyboard","description":"Teclado mecánico","price":250}`,
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
		"sku": "LAPTOP-IT",
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

func TestListProductsSortByPriceDescending(t *testing.T) {
	handler := newTestHTTPHandler()

	products := []string{
		`{"sku": "MOUSE-IT", "name":"Mouse","description":"Mouse inalámbrico","price":120}`,
		`{"sku": "LAPTOP-IT", "name":"Laptop","description":"Equipo para desarrollo backend","price":3500}`,
		`{"sku": "KEYBOARD-IT", "name":"Keyboard","description":"Teclado mecánico","price":250}`,
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

	request := httptest.NewRequest(http.MethodGet, "/api/v1/products?sort=price&order=desc", nil)
	request.Header.Set(requestIDHeader, "sort-products")

	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, responseRecorder.Code)
	}

	var response ProductListResponse
	if err := json.NewDecoder(responseRecorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode product list response: %v", err)
	}

	if len(response.Data) != 3 {
		t.Fatalf("expected 3 products, got %d", len(response.Data))
	}

	expectedNames := []string{"Laptop", "Keyboard", "Mouse"}

	for index, expectedName := range expectedNames {
		if response.Data[index].Name != expectedName {
			t.Fatalf("expected product at index %d to be %q, got %q", index, expectedName, response.Data[index].Name)
		}
	}

	if response.Meta.Sort != product.SortFieldPrice {
		t.Fatalf("expected sort %q, got %q", product.SortFieldPrice, response.Meta.Sort)
	}

	if response.Meta.Order != product.SortOrderDesc {
		t.Fatalf("expected order %q, got %q", product.SortOrderDesc, response.Meta.Order)
	}
}

func TestListProductsInvalidSortAndOrder(t *testing.T) {
	handler := newTestHTTPHandler()

	request := httptest.NewRequest(http.MethodGet, "/api/v1/products?sort=unknown&order=random", nil)
	request.Header.Set(requestIDHeader, "invalid-sort-order")

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

	if response.Error.RequestID != "invalid-sort-order" {
		t.Fatalf("expected request id %q, got %q", "invalid-sort-order", response.Error.RequestID)
	}

	if len(response.Error.Fields) != 2 {
		t.Fatalf("expected 2 field errors, got %d", len(response.Error.Fields))
	}
}

func TestListProductsSearchAndSort(t *testing.T) {
	handler := newTestHTTPHandler()

	products := []string{
		`{"sku":"LAPTOP-BASIC-IT","name":"Laptop Basic","description":"Laptop para oficina","price":2500}`,
		`{"sku":"LAPTOP-PRO-IT","name":"Laptop Pro","description":"Laptop para desarrollo backend","price":4500}`,
		`{"sku":"LAPTOP-AIR-IT","name":"Laptop Air","description":"Laptop ligera","price":3500}`,
		`{"sku":"MOUSE-IT","name":"Mouse","description":"Mouse inalámbrico","price":120}`,
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

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/products?search=laptop&sort=price&order=desc&page=1&page_size=2",
		nil,
	)
	request.Header.Set(requestIDHeader, "search-sort-products")

	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, responseRecorder.Code)
	}

	var response ProductListResponse
	if err := json.NewDecoder(responseRecorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode product list response: %v", err)
	}

	if len(response.Data) != 2 {
		t.Fatalf("expected 2 products, got %d", len(response.Data))
	}

	expectedNames := []string{"Laptop Pro", "Laptop Air"}

	for index, expectedName := range expectedNames {
		if response.Data[index].Name != expectedName {
			t.Fatalf("expected product at index %d to be %q, got %q", index, expectedName, response.Data[index].Name)
		}
	}

	if response.Meta.Total != 3 {
		t.Fatalf("expected total %d, got %d", 3, response.Meta.Total)
	}

	if response.Meta.TotalPages != 2 {
		t.Fatalf("expected total pages %d, got %d", 2, response.Meta.TotalPages)
	}

	if response.Meta.Search != "laptop" {
		t.Fatalf("expected search %q, got %q", "laptop", response.Meta.Search)
	}

	if response.Meta.Sort != product.SortFieldPrice {
		t.Fatalf("expected sort %q, got %q", product.SortFieldPrice, response.Meta.Sort)
	}

	if response.Meta.Order != product.SortOrderDesc {
		t.Fatalf("expected order %q, got %q", product.SortOrderDesc, response.Meta.Order)
	}
}

func TestCreateProductDuplicateSKUReturnsConflict(t *testing.T) {
	handler := newTestHTTPHandler()

	firstBody := []byte(`{
		"sku": "DUPLICATE-SKU",
		"name": "Producto 1",
		"description": "Primer producto",
		"price": 100
	}`)

	firstRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/products",
		bytes.NewReader(firstBody),
	)
	firstRequest.Header.Set("Content-Type", "application/json")

	firstRecorder := httptest.NewRecorder()

	handler.ServeHTTP(firstRecorder, firstRequest)

	if firstRecorder.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d. body: %s", http.StatusCreated, firstRecorder.Code, firstRecorder.Body.String())
	}

	secondBody := []byte(`{
		"sku": "duplicate-sku",
		"name": "Producto 2",
		"description": "Segundo producto",
		"price": 200
	}`)

	secondRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/products",
		bytes.NewReader(secondBody),
	)
	secondRequest.Header.Set("Content-Type", "application/json")
	secondRequest.Header.Set(requestIDHeader, "duplicate-sku-create")

	secondRecorder := httptest.NewRecorder()

	handler.ServeHTTP(secondRecorder, secondRequest)

	if secondRecorder.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d. body: %s", http.StatusConflict, secondRecorder.Code, secondRecorder.Body.String())
	}

	var response ErrorResponse
	if err := json.NewDecoder(secondRecorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if response.Error.Code != errorCodeConflict {
		t.Fatalf("expected error code %q, got %q", errorCodeConflict, response.Error.Code)
	}

	if response.Error.Message != "product sku already exists" {
		t.Fatalf("expected message %q, got %q", "product sku already exists", response.Error.Message)
	}

	if response.Error.RequestID != "duplicate-sku-create" {
		t.Fatalf("expected request id %q, got %q", "duplicate-sku-create", response.Error.RequestID)
	}
}

func TestUpdateProductDuplicateSKUReturnsConflict(t *testing.T) {
	handler := newTestHTTPHandler()

	firstBody := []byte(`{
		"sku": "PRODUCT-001",
		"name": "Producto 1",
		"description": "Primer producto",
		"price": 100
	}`)

	firstRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/products",
		bytes.NewReader(firstBody),
	)
	firstRequest.Header.Set("Content-Type", "application/json")

	firstRecorder := httptest.NewRecorder()
	handler.ServeHTTP(firstRecorder, firstRequest)

	if firstRecorder.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d. body: %s", http.StatusCreated, firstRecorder.Code, firstRecorder.Body.String())
	}

	secondBody := []byte(`{
		"sku": "PRODUCT-002",
		"name": "Producto 2",
		"description": "Segundo producto",
		"price": 200
	}`)

	secondRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/products",
		bytes.NewReader(secondBody),
	)
	secondRequest.Header.Set("Content-Type", "application/json")

	secondRecorder := httptest.NewRecorder()
	handler.ServeHTTP(secondRecorder, secondRequest)

	if secondRecorder.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d. body: %s", http.StatusCreated, secondRecorder.Code, secondRecorder.Body.String())
	}

	var secondProduct ProductResponse
	if err := json.NewDecoder(secondRecorder.Body).Decode(&secondProduct); err != nil {
		t.Fatalf("failed to decode second product: %v", err)
	}

	updateBody := []byte(`{
		"sku": "PRODUCT-001",
		"name": "Producto 2 actualizado",
		"description": "Intento de SKU duplicado",
		"price": 250
	}`)

	updateRequest := httptest.NewRequest(
		http.MethodPut,
		"/api/v1/products/"+secondProduct.ID,
		bytes.NewReader(updateBody),
	)
	updateRequest.Header.Set("Content-Type", "application/json")
	updateRequest.Header.Set(requestIDHeader, "duplicate-sku-update")

	updateRecorder := httptest.NewRecorder()

	handler.ServeHTTP(updateRecorder, updateRequest)

	if updateRecorder.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d. body: %s", http.StatusConflict, updateRecorder.Code, updateRecorder.Body.String())
	}

	var response ErrorResponse
	if err := json.NewDecoder(updateRecorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if response.Error.Code != errorCodeConflict {
		t.Fatalf("expected error code %q, got %q", errorCodeConflict, response.Error.Code)
	}

	if response.Error.Message != "product sku already exists" {
		t.Fatalf("expected message %q, got %q", "product sku already exists", response.Error.Message)
	}

	if response.Error.RequestID != "duplicate-sku-update" {
		t.Fatalf("expected request id %q, got %q", "duplicate-sku-update", response.Error.RequestID)
	}
}

func TestGetProductBySKU(t *testing.T) {
	handler := newTestHTTPHandler()

	createBody := []byte(`{
		"sku": "LAPTOP-SKU-TEST",
		"name": "Laptop SKU Test",
		"description": "Producto para búsqueda por SKU",
		"price": 3500
	}`)

	createRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/products",
		bytes.NewReader(createBody),
	)
	createRequest.Header.Set("Content-Type", "application/json")

	createRecorder := httptest.NewRecorder()

	handler.ServeHTTP(createRecorder, createRequest)

	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d. body: %s", http.StatusCreated, createRecorder.Code, createRecorder.Body.String())
	}

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/products/sku/laptop-sku-test",
		nil,
	)
	request.Header.Set(requestIDHeader, "get-product-by-sku")

	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d. body: %s", http.StatusOK, responseRecorder.Code, responseRecorder.Body.String())
	}

	var response ProductResponse
	if err := json.NewDecoder(responseRecorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode product response: %v", err)
	}

	if response.SKU != "LAPTOP-SKU-TEST" {
		t.Fatalf("expected sku %q, got %q", "LAPTOP-SKU-TEST", response.SKU)
	}

	if response.Name != "Laptop SKU Test" {
		t.Fatalf("expected name %q, got %q", "Laptop SKU Test", response.Name)
	}
}

func TestGetProductBySKUNotFound(t *testing.T) {
	handler := newTestHTTPHandler()

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/products/sku/missing-sku",
		nil,
	)
	request.Header.Set(requestIDHeader, "product-sku-not-found")

	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d. body: %s", http.StatusNotFound, responseRecorder.Code, responseRecorder.Body.String())
	}

	var response ErrorResponse
	if err := json.NewDecoder(responseRecorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if response.Error.Code != errorCodeNotFound {
		t.Fatalf("expected error code %q, got %q", errorCodeNotFound, response.Error.Code)
	}

	if response.Error.Message != "product not found" {
		t.Fatalf("expected message %q, got %q", "product not found", response.Error.Message)
	}

	if response.Error.RequestID != "product-sku-not-found" {
		t.Fatalf("expected request id %q, got %q", "product-sku-not-found", response.Error.RequestID)
	}
}

func TestGetProductBySKUMethodNotAllowed(t *testing.T) {
	handler := newTestHTTPHandler()

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/products/sku/LAPTOP-001",
		nil,
	)
	request.Header.Set(requestIDHeader, "product-sku-method-not-allowed")

	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, responseRecorder.Code)
	}

	if got := responseRecorder.Header().Get("Allow"); got != http.MethodGet {
		t.Fatalf("expected Allow header %q, got %q", http.MethodGet, got)
	}
}

func TestListProductsFilterByPriceRange(t *testing.T) {
	handler := newTestHTTPHandler()

	products := []string{
		`{"sku":"LAPTOP-BASIC-PRICE","name":"Laptop Basic","description":"Laptop para oficina","price":2500}`,
		`{"sku":"LAPTOP-PRO-PRICE","name":"Laptop Pro","description":"Laptop para desarrollo backend","price":4500}`,
		`{"sku":"LAPTOP-AIR-PRICE","name":"Laptop Air","description":"Laptop ligera","price":3500}`,
		`{"sku":"MOUSE-PRICE","name":"Mouse","description":"Mouse inalámbrico","price":120}`,
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
			t.Fatalf("expected status %d, got %d. body: %s", http.StatusCreated, responseRecorder.Code, responseRecorder.Body.String())
		}
	}

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/products?min_price=2000&max_price=4000&sort=price&order=asc",
		nil,
	)
	request.Header.Set(requestIDHeader, "price-range-products")

	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d. body: %s", http.StatusOK, responseRecorder.Code, responseRecorder.Body.String())
	}

	var response ProductListResponse
	if err := json.NewDecoder(responseRecorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode product list response: %v", err)
	}

	if len(response.Data) != 2 {
		t.Fatalf("expected 2 products, got %d", len(response.Data))
	}

	expectedNames := []string{
		"Laptop Basic",
		"Laptop Air",
	}

	for index, expectedName := range expectedNames {
		if response.Data[index].Name != expectedName {
			t.Fatalf("expected product at index %d to be %q, got %q", index, expectedName, response.Data[index].Name)
		}
	}

	if response.Meta.Total != 2 {
		t.Fatalf("expected total %d, got %d", 2, response.Meta.Total)
	}

	if response.Meta.MinPrice == nil || *response.Meta.MinPrice != 2000 {
		t.Fatalf("expected min_price %v, got %v", 2000.0, response.Meta.MinPrice)
	}

	if response.Meta.MaxPrice == nil || *response.Meta.MaxPrice != 4000 {
		t.Fatalf("expected max_price %v, got %v", 4000.0, response.Meta.MaxPrice)
	}
}

func TestListProductsInvalidPriceRange(t *testing.T) {
	handler := newTestHTTPHandler()

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/products?min_price=5000&max_price=1000",
		nil,
	)
	request.Header.Set(requestIDHeader, "invalid-price-range")

	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status %d, got %d. body: %s", http.StatusUnprocessableEntity, responseRecorder.Code, responseRecorder.Body.String())
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

	if response.Error.Fields[0].Field != "price_range" {
		t.Fatalf("expected field %q, got %q", "price_range", response.Error.Fields[0].Field)
	}

	if response.Error.RequestID != "invalid-price-range" {
		t.Fatalf("expected request id %q, got %q", "invalid-price-range", response.Error.RequestID)
	}
}

func TestListProductsInvalidMinPrice(t *testing.T) {
	handler := newTestHTTPHandler()

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/products?min_price=abc",
		nil,
	)
	request.Header.Set(requestIDHeader, "invalid-min-price")

	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status %d, got %d. body: %s", http.StatusUnprocessableEntity, responseRecorder.Code, responseRecorder.Body.String())
	}

	var response ErrorResponse
	if err := json.NewDecoder(responseRecorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if len(response.Error.Fields) != 1 {
		t.Fatalf("expected 1 field error, got %d", len(response.Error.Fields))
	}

	if response.Error.Fields[0].Field != "min_price" {
		t.Fatalf("expected field %q, got %q", "min_price", response.Error.Fields[0].Field)
	}
}
