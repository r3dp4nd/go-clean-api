package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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

	createdProduct := decodeProductResourceResponse(t, createRecorder)

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

	var fetchedResponse ProductResourceResponse
	if err := json.NewDecoder(getRecorder.Body).Decode(&fetchedResponse); err != nil {
		t.Fatalf("failed to decode fetched product response: %v", err)
	}

	fetchedProduct := fetchedResponse.Data

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

	var updatedResponse ProductResourceResponse
	if err := json.NewDecoder(updateRecorder.Body).Decode(&updatedResponse); err != nil {
		t.Fatalf("failed to decode updated product response: %v", err)
	}

	updatedProduct := updatedResponse.Data

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

	var secondResponse ProductResourceResponse
	if err := json.NewDecoder(secondRecorder.Body).Decode(&secondResponse); err != nil {
		t.Fatalf("failed to decode second product response: %v", err)
	}

	secondProduct := secondResponse.Data

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

	var response ProductResourceResponse
	if err := json.NewDecoder(responseRecorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode product resource response: %v", err)
	}

	if response.Data.SKU != "LAPTOP-SKU-TEST" {
		t.Fatalf("expected sku %q, got %q", "LAPTOP-SKU-TEST", response.Data.SKU)
	}

	if response.Data.Name != "Laptop SKU Test" {
		t.Fatalf("expected name %q, got %q", "Laptop SKU Test", response.Data.Name)
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

func TestListProductsFilterByCreatedRange(t *testing.T) {
	handler := newTestHTTPHandler()

	products := []string{
		`{"sku":"DATE-LAPTOP-BASIC","name":"Laptop Basic","description":"Laptop para oficina","price":2500}`,
		`{"sku":"DATE-LAPTOP-PRO","name":"Laptop Pro","description":"Laptop para desarrollo backend","price":4500}`,
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

	createdFrom := time.Now().UTC().Add(-24 * time.Hour).Format("2006-01-02")
	createdTo := time.Now().UTC().Add(24 * time.Hour).Format("2006-01-02")

	request := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf(
			"/api/v1/products?created_from=%s&created_to=%s&sort=created_at&order=asc",
			createdFrom,
			createdTo,
		),
		nil,
	)
	request.Header.Set(requestIDHeader, "created-range-products")

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

	if response.Meta.Total != 2 {
		t.Fatalf("expected total %d, got %d", 2, response.Meta.Total)
	}

	if response.Meta.CreatedFrom == nil {
		t.Fatal("expected created_from meta to be set")
	}

	if response.Meta.CreatedTo == nil {
		t.Fatal("expected created_to meta to be set")
	}

	if response.Meta.Sort != product.SortFieldCreatedAt {
		t.Fatalf("expected sort %q, got %q", product.SortFieldCreatedAt, response.Meta.Sort)
	}

	if response.Meta.Order != product.SortOrderAsc {
		t.Fatalf("expected order %q, got %q", product.SortOrderAsc, response.Meta.Order)
	}
}

func TestListProductsInvalidCreatedRange(t *testing.T) {
	handler := newTestHTTPHandler()

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/products?created_from=2026-12-31&created_to=2026-01-01",
		nil,
	)
	request.Header.Set(requestIDHeader, "invalid-created-range")

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

	if response.Error.Fields[0].Field != "created_range" {
		t.Fatalf("expected field %q, got %q", "created_range", response.Error.Fields[0].Field)
	}

	if response.Error.RequestID != "invalid-created-range" {
		t.Fatalf("expected request id %q, got %q", "invalid-created-range", response.Error.RequestID)
	}
}

func TestListProductsInvalidCreatedFrom(t *testing.T) {
	handler := newTestHTTPHandler()

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/products?created_from=not-a-date",
		nil,
	)
	request.Header.Set(requestIDHeader, "invalid-created-from")

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

	if response.Error.Fields[0].Field != "created_from" {
		t.Fatalf("expected field %q, got %q", "created_from", response.Error.Fields[0].Field)
	}
}

func TestCreateProductReturnsResourceEnvelope(t *testing.T) {
	handler := newTestHTTPHandler()

	body := []byte(`{
		"sku": "ENVELOPE-001",
		"name": "Envelope Product",
		"description": "Producto para validar envelope",
		"price": 100
	}`)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/products",
		bytes.NewReader(body),
	)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set(requestIDHeader, "envelope-create")

	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d. body: %s", http.StatusCreated, responseRecorder.Code, responseRecorder.Body.String())
	}

	var response ProductResourceResponse
	if err := json.NewDecoder(responseRecorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode product resource response: %v", err)
	}

	if response.Data.ID == "" {
		t.Fatal("expected data.id to be generated")
	}

	if response.Data.SKU != "ENVELOPE-001" {
		t.Fatalf("expected data.sku %q, got %q", "ENVELOPE-001", response.Data.SKU)
	}

	if response.Data.Name != "Envelope Product" {
		t.Fatalf("expected data.name %q, got %q", "Envelope Product", response.Data.Name)
	}
}

func TestProductSKUExistsReturnsTrue(t *testing.T) {
	handler := newTestHTTPHandler()

	createBody := []byte(`{
		"sku": "EXISTS-SKU-001",
		"name": "Exists Product",
		"description": "Producto para validar existencia de SKU",
		"price": 100
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
		"/api/v1/products/exists?sku=exists-sku-001",
		nil,
	)
	request.Header.Set(requestIDHeader, "sku-exists-true")

	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d. body: %s", http.StatusOK, responseRecorder.Code, responseRecorder.Body.String())
	}

	var response ProductSKUExistsResponse
	if err := json.NewDecoder(responseRecorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode sku exists response: %v", err)
	}

	if response.Data.SKU != "EXISTS-SKU-001" {
		t.Fatalf("expected sku %q, got %q", "EXISTS-SKU-001", response.Data.SKU)
	}

	if !response.Data.Exists {
		t.Fatal("expected exists to be true")
	}
}

func TestProductSKUExistsReturnsFalse(t *testing.T) {
	handler := newTestHTTPHandler()

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/products/exists?sku=missing-sku",
		nil,
	)
	request.Header.Set(requestIDHeader, "sku-exists-false")

	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d. body: %s", http.StatusOK, responseRecorder.Code, responseRecorder.Body.String())
	}

	var response ProductSKUExistsResponse
	if err := json.NewDecoder(responseRecorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode sku exists response: %v", err)
	}

	if response.Data.SKU != "MISSING-SKU" {
		t.Fatalf("expected sku %q, got %q", "MISSING-SKU", response.Data.SKU)
	}

	if response.Data.Exists {
		t.Fatal("expected exists to be false")
	}
}

func TestProductSKUExistsRequiresSKU(t *testing.T) {
	handler := newTestHTTPHandler()

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/products/exists",
		nil,
	)
	request.Header.Set(requestIDHeader, "sku-exists-required")

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

	if response.Error.RequestID != "sku-exists-required" {
		t.Fatalf("expected request id %q, got %q", "sku-exists-required", response.Error.RequestID)
	}

	if len(response.Error.Fields) != 1 {
		t.Fatalf("expected 1 field error, got %d", len(response.Error.Fields))
	}

	if response.Error.Fields[0].Field != "sku" {
		t.Fatalf("expected field %q, got %q", "sku", response.Error.Fields[0].Field)
	}
}

func TestProductSKUExistsMethodNotAllowed(t *testing.T) {
	handler := newTestHTTPHandler()

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/products/exists?sku=LAPTOP-001",
		nil,
	)
	request.Header.Set(requestIDHeader, "sku-exists-method-not-allowed")

	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, responseRecorder.Code)
	}

	if got := responseRecorder.Header().Get("Allow"); got != http.MethodGet {
		t.Fatalf("expected Allow header %q, got %q", http.MethodGet, got)
	}
}

func TestPatchProductPartialUpdate(t *testing.T) {
	handler := newTestHTTPHandler()

	createBody := []byte(`{
		"sku": "PATCH-LAPTOP-001",
		"name": "Laptop",
		"description": "Laptop básica",
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

	createdProduct := decodeProductResourceResponse(t, createRecorder)

	patchBody := []byte(`{
		"name": "Laptop Pro",
		"price": 4200
	}`)

	patchRequest := httptest.NewRequest(
		http.MethodPatch,
		"/api/v1/products/"+createdProduct.ID,
		bytes.NewReader(patchBody),
	)
	patchRequest.Header.Set("Content-Type", "application/json")
	patchRequest.Header.Set(requestIDHeader, "patch-product")

	patchRecorder := httptest.NewRecorder()

	handler.ServeHTTP(patchRecorder, patchRequest)

	if patchRecorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d. body: %s", http.StatusOK, patchRecorder.Code, patchRecorder.Body.String())
	}

	updatedProduct := decodeProductResourceResponse(t, patchRecorder)

	if updatedProduct.ID != createdProduct.ID {
		t.Fatalf("expected id %q, got %q", createdProduct.ID, updatedProduct.ID)
	}

	if updatedProduct.SKU != "PATCH-LAPTOP-001" {
		t.Fatalf("expected sku to be preserved, got %q", updatedProduct.SKU)
	}

	if updatedProduct.Name != "Laptop Pro" {
		t.Fatalf("expected name %q, got %q", "Laptop Pro", updatedProduct.Name)
	}

	if updatedProduct.Description != "Laptop básica" {
		t.Fatalf("expected description to be preserved, got %q", updatedProduct.Description)
	}

	if updatedProduct.Price != 4200 {
		t.Fatalf("expected price %v, got %v", 4200.0, updatedProduct.Price)
	}
}

func TestPatchProductEmptyBodyReturnsValidationError(t *testing.T) {
	handler := newTestHTTPHandler()

	createBody := []byte(`{
		"sku": "PATCH-EMPTY-001",
		"name": "Laptop",
		"description": "Laptop básica",
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

	createdProduct := decodeProductResourceResponse(t, createRecorder)

	patchRequest := httptest.NewRequest(
		http.MethodPatch,
		"/api/v1/products/"+createdProduct.ID,
		bytes.NewReader([]byte(`{}`)),
	)
	patchRequest.Header.Set("Content-Type", "application/json")
	patchRequest.Header.Set(requestIDHeader, "patch-empty-body")

	patchRecorder := httptest.NewRecorder()

	handler.ServeHTTP(patchRecorder, patchRequest)

	if patchRecorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status %d, got %d. body: %s", http.StatusUnprocessableEntity, patchRecorder.Code, patchRecorder.Body.String())
	}

	var response ErrorResponse
	if err := json.NewDecoder(patchRecorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if response.Error.Code != errorCodeValidation {
		t.Fatalf("expected error code %q, got %q", errorCodeValidation, response.Error.Code)
	}

	if len(response.Error.Fields) != 1 {
		t.Fatalf("expected 1 field error, got %d", len(response.Error.Fields))
	}

	if response.Error.Fields[0].Field != "body" {
		t.Fatalf("expected field %q, got %q", "body", response.Error.Fields[0].Field)
	}
}

func TestPatchProductNotFound(t *testing.T) {
	handler := newTestHTTPHandler()

	patchBody := []byte(`{
		"name": "Laptop Pro"
	}`)

	request := httptest.NewRequest(
		http.MethodPatch,
		"/api/v1/products/999",
		bytes.NewReader(patchBody),
	)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set(requestIDHeader, "patch-product-not-found")

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
}

func TestPatchProductDuplicateSKUReturnsConflict(t *testing.T) {
	handler := newTestHTTPHandler()

	firstBody := []byte(`{
		"sku": "PATCH-DUPLICATE-001",
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
		"sku": "PATCH-DUPLICATE-002",
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

	secondProduct := decodeProductResourceResponse(t, secondRecorder)

	patchBody := []byte(`{
		"sku": "PATCH-DUPLICATE-001"
	}`)

	patchRequest := httptest.NewRequest(
		http.MethodPatch,
		"/api/v1/products/"+secondProduct.ID,
		bytes.NewReader(patchBody),
	)
	patchRequest.Header.Set("Content-Type", "application/json")
	patchRequest.Header.Set(requestIDHeader, "patch-duplicate-sku")

	patchRecorder := httptest.NewRecorder()

	handler.ServeHTTP(patchRecorder, patchRequest)

	if patchRecorder.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d. body: %s", http.StatusConflict, patchRecorder.Code, patchRecorder.Body.String())
	}

	var response ErrorResponse
	if err := json.NewDecoder(patchRecorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if response.Error.Code != errorCodeConflict {
		t.Fatalf("expected error code %q, got %q", errorCodeConflict, response.Error.Code)
	}

	if response.Error.Message != "product sku already exists" {
		t.Fatalf("expected message %q, got %q", "product sku already exists", response.Error.Message)
	}
}

func TestPatchProductInvalidJSON(t *testing.T) {
	handler := newTestHTTPHandler()

	request := httptest.NewRequest(
		http.MethodPatch,
		"/api/v1/products/1",
		bytes.NewReader([]byte(`{"name":`)),
	)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set(requestIDHeader, "patch-invalid-json")

	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d. body: %s", http.StatusBadRequest, responseRecorder.Code, responseRecorder.Body.String())
	}

	var response ErrorResponse
	if err := json.NewDecoder(responseRecorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if response.Error.Code != errorCodeInvalidRequest {
		t.Fatalf("expected error code %q, got %q", errorCodeInvalidRequest, response.Error.Code)
	}
}

func decodeProductResourceResponse(t *testing.T, recorder *httptest.ResponseRecorder) ProductResponse {
	t.Helper()

	var response ProductResourceResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode product resource response: %v", err)
	}

	return response.Data
}

func TestDeleteProductSoftDeletesFromPublicAPI(t *testing.T) {
	handler := newTestHTTPHandler()

	createBody := []byte(`{
		"sku": "HTTP-SOFT-DELETE-001",
		"name": "HTTP Soft Delete",
		"description": "Producto para probar soft delete desde HTTP",
		"price": 100
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
		t.Fatalf(
			"expected status %d, got %d. body: %s",
			http.StatusCreated,
			createRecorder.Code,
			createRecorder.Body.String(),
		)
	}

	createdProduct := decodeProductResourceResponse(t, createRecorder)

	if createdProduct.ID == "" {
		t.Fatal("expected created product ID to be generated")
	}

	if createdProduct.SKU != "HTTP-SOFT-DELETE-001" {
		t.Fatalf("expected sku %q, got %q", "HTTP-SOFT-DELETE-001", createdProduct.SKU)
	}

	deleteRequest := httptest.NewRequest(
		http.MethodDelete,
		"/api/v1/products/"+createdProduct.ID,
		nil,
	)
	deleteRequest.Header.Set(requestIDHeader, "soft-delete-product")

	deleteRecorder := httptest.NewRecorder()
	handler.ServeHTTP(deleteRecorder, deleteRequest)

	if deleteRecorder.Code != http.StatusNoContent {
		t.Fatalf(
			"expected status %d, got %d. body: %s",
			http.StatusNoContent,
			deleteRecorder.Code,
			deleteRecorder.Body.String(),
		)
	}

	getRequest := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/products/"+createdProduct.ID,
		nil,
	)
	getRequest.Header.Set(requestIDHeader, "get-soft-deleted-product")

	getRecorder := httptest.NewRecorder()
	handler.ServeHTTP(getRecorder, getRequest)

	if getRecorder.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d. body: %s",
			http.StatusNotFound,
			getRecorder.Code,
			getRecorder.Body.String(),
		)
	}

	var getErrorResponse ErrorResponse
	if err := json.NewDecoder(getRecorder.Body).Decode(&getErrorResponse); err != nil {
		t.Fatalf("failed to decode get error response: %v", err)
	}

	if getErrorResponse.Error.Code != errorCodeNotFound {
		t.Fatalf("expected error code %q, got %q", errorCodeNotFound, getErrorResponse.Error.Code)
	}

	getBySKURequest := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/products/sku/HTTP-SOFT-DELETE-001",
		nil,
	)
	getBySKURequest.Header.Set(requestIDHeader, "get-soft-deleted-product-by-sku")

	getBySKURecorder := httptest.NewRecorder()
	handler.ServeHTTP(getBySKURecorder, getBySKURequest)

	if getBySKURecorder.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d. body: %s",
			http.StatusNotFound,
			getBySKURecorder.Code,
			getBySKURecorder.Body.String(),
		)
	}

	var getBySKUErrorResponse ErrorResponse
	if err := json.NewDecoder(getBySKURecorder.Body).Decode(&getBySKUErrorResponse); err != nil {
		t.Fatalf("failed to decode get by sku error response: %v", err)
	}

	if getBySKUErrorResponse.Error.Code != errorCodeNotFound {
		t.Fatalf("expected error code %q, got %q", errorCodeNotFound, getBySKUErrorResponse.Error.Code)
	}

	skuExistsRequest := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/products/exists?sku=HTTP-SOFT-DELETE-001",
		nil,
	)
	skuExistsRequest.Header.Set(requestIDHeader, "soft-deleted-sku-exists")

	skuExistsRecorder := httptest.NewRecorder()
	handler.ServeHTTP(skuExistsRecorder, skuExistsRequest)

	if skuExistsRecorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d. body: %s",
			http.StatusOK,
			skuExistsRecorder.Code,
			skuExistsRecorder.Body.String(),
		)
	}

	var skuExistsResponse ProductSKUExistsResponse
	if err := json.NewDecoder(skuExistsRecorder.Body).Decode(&skuExistsResponse); err != nil {
		t.Fatalf("failed to decode sku exists response: %v", err)
	}

	if skuExistsResponse.Data.SKU != "HTTP-SOFT-DELETE-001" {
		t.Fatalf("expected sku %q, got %q", "HTTP-SOFT-DELETE-001", skuExistsResponse.Data.SKU)
	}

	if skuExistsResponse.Data.Exists {
		t.Fatal("expected soft deleted sku to not exist in public API")
	}

	listRequest := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/products",
		nil,
	)

	listRecorder := httptest.NewRecorder()
	handler.ServeHTTP(listRecorder, listRequest)

	if listRecorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d. body: %s",
			http.StatusOK,
			listRecorder.Code,
			listRecorder.Body.String(),
		)
	}

	var listResponse ProductListResponse
	if err := json.NewDecoder(listRecorder.Body).Decode(&listResponse); err != nil {
		t.Fatalf("failed to decode product list response: %v", err)
	}

	if listResponse.Meta.Total != 0 {
		t.Fatalf("expected total %d after soft delete, got %d", 0, listResponse.Meta.Total)
	}

	if len(listResponse.Data) != 0 {
		t.Fatalf("expected 0 products after soft delete, got %d", len(listResponse.Data))
	}
}

func TestCreateProductAllowsReusingSKUAfterSoftDelete(t *testing.T) {
	handler := newTestHTTPHandler()

	firstBody := []byte(`{
		"sku": "HTTP-REUSABLE-SKU",
		"name": "First Product",
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

	firstProduct := decodeProductResourceResponse(t, firstRecorder)

	deleteRequest := httptest.NewRequest(
		http.MethodDelete,
		"/api/v1/products/"+firstProduct.ID,
		nil,
	)

	deleteRecorder := httptest.NewRecorder()
	handler.ServeHTTP(deleteRecorder, deleteRequest)

	if deleteRecorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d. body: %s", http.StatusNoContent, deleteRecorder.Code, deleteRecorder.Body.String())
	}

	secondBody := []byte(`{
		"sku": "http-reusable-sku",
		"name": "Second Product",
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

	secondProduct := decodeProductResourceResponse(t, secondRecorder)

	if secondProduct.ID == firstProduct.ID {
		t.Fatalf("expected different product IDs, got same ID %q", secondProduct.ID)
	}

	if secondProduct.SKU != "HTTP-REUSABLE-SKU" {
		t.Fatalf("expected normalized sku %q, got %q", "HTTP-REUSABLE-SKU", secondProduct.SKU)
	}
}

func TestRestoreProduct(t *testing.T) {
	handler := newTestHTTPHandler()

	createBody := []byte(`{
		"sku": "HTTP-RESTORE-001",
		"name": "HTTP Restore",
		"description": "Producto para probar restore",
		"price": 100
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

	createdProduct := decodeProductResourceResponse(t, createRecorder)

	deleteRequest := httptest.NewRequest(
		http.MethodDelete,
		"/api/v1/products/"+createdProduct.ID,
		nil,
	)

	deleteRecorder := httptest.NewRecorder()
	handler.ServeHTTP(deleteRecorder, deleteRequest)

	if deleteRecorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d. body: %s", http.StatusNoContent, deleteRecorder.Code, deleteRecorder.Body.String())
	}

	restoreRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/products/"+createdProduct.ID+"/restore",
		nil,
	)
	restoreRequest.Header.Set(requestIDHeader, "restore-product")

	restoreRecorder := httptest.NewRecorder()
	handler.ServeHTTP(restoreRecorder, restoreRequest)

	if restoreRecorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d. body: %s", http.StatusOK, restoreRecorder.Code, restoreRecorder.Body.String())
	}

	restoredProduct := decodeProductResourceResponse(t, restoreRecorder)

	if restoredProduct.ID != createdProduct.ID {
		t.Fatalf("expected restored ID %q, got %q", createdProduct.ID, restoredProduct.ID)
	}

	if restoredProduct.SKU != "HTTP-RESTORE-001" {
		t.Fatalf("expected restored SKU %q, got %q", "HTTP-RESTORE-001", restoredProduct.SKU)
	}

	getRequest := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/products/"+createdProduct.ID,
		nil,
	)

	getRecorder := httptest.NewRecorder()
	handler.ServeHTTP(getRecorder, getRequest)

	if getRecorder.Code != http.StatusOK {
		t.Fatalf("expected status %d after restore, got %d. body: %s", http.StatusOK, getRecorder.Code, getRecorder.Body.String())
	}
}

func TestRestoreProductDuplicateSKUReturnsConflict(t *testing.T) {
	handler := newTestHTTPHandler()

	firstBody := []byte(`{
		"sku": "HTTP-RESTORE-CONFLICT",
		"name": "First Product",
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

	firstProduct := decodeProductResourceResponse(t, firstRecorder)

	deleteRequest := httptest.NewRequest(
		http.MethodDelete,
		"/api/v1/products/"+firstProduct.ID,
		nil,
	)

	deleteRecorder := httptest.NewRecorder()
	handler.ServeHTTP(deleteRecorder, deleteRequest)

	if deleteRecorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d. body: %s", http.StatusNoContent, deleteRecorder.Code, deleteRecorder.Body.String())
	}

	secondBody := []byte(`{
		"sku": "HTTP-RESTORE-CONFLICT",
		"name": "Second Product",
		"description": "Segundo producto activo",
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

	restoreRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/products/"+firstProduct.ID+"/restore",
		nil,
	)
	restoreRequest.Header.Set(requestIDHeader, "restore-duplicate-sku")

	restoreRecorder := httptest.NewRecorder()
	handler.ServeHTTP(restoreRecorder, restoreRequest)

	if restoreRecorder.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d. body: %s", http.StatusConflict, restoreRecorder.Code, restoreRecorder.Body.String())
	}

	var response ErrorResponse
	if err := json.NewDecoder(restoreRecorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if response.Error.Code != errorCodeConflict {
		t.Fatalf("expected error code %q, got %q", errorCodeConflict, response.Error.Code)
	}

	if response.Error.Message != "product sku already exists" {
		t.Fatalf("expected message %q, got %q", "product sku already exists", response.Error.Message)
	}
}

func TestRestoreProductMethodNotAllowed(t *testing.T) {
	handler := newTestHTTPHandler()

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/products/123/restore",
		nil,
	)
	request.Header.Set(requestIDHeader, "restore-method-not-allowed")

	responseRecorder := httptest.NewRecorder()
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, responseRecorder.Code)
	}

	if got := responseRecorder.Header().Get("Allow"); got != http.MethodPost {
		t.Fatalf("expected Allow header %q, got %q", http.MethodPost, got)
	}
}

func TestListDeletedProducts(t *testing.T) {
	handler := newTestHTTPHandler()

	activeBody := []byte(`{
		"sku": "HTTP-ACTIVE-001",
		"name": "Active Product",
		"description": "Producto activo",
		"price": 100
	}`)

	activeRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/products",
		bytes.NewReader(activeBody),
	)
	activeRequest.Header.Set("Content-Type", "application/json")

	activeRecorder := httptest.NewRecorder()
	handler.ServeHTTP(activeRecorder, activeRequest)

	if activeRecorder.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d. body: %s", http.StatusCreated, activeRecorder.Code, activeRecorder.Body.String())
	}

	deletedBody := []byte(`{
		"sku": "HTTP-DELETED-001",
		"name": "Deleted Product",
		"description": "Producto eliminado",
		"price": 200
	}`)

	deletedRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/products",
		bytes.NewReader(deletedBody),
	)
	deletedRequest.Header.Set("Content-Type", "application/json")

	deletedRecorder := httptest.NewRecorder()
	handler.ServeHTTP(deletedRecorder, deletedRequest)

	if deletedRecorder.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d. body: %s", http.StatusCreated, deletedRecorder.Code, deletedRecorder.Body.String())
	}

	deletedProduct := decodeProductResourceResponse(t, deletedRecorder)

	deleteRequest := httptest.NewRequest(
		http.MethodDelete,
		"/api/v1/products/"+deletedProduct.ID,
		nil,
	)

	deleteRecorder := httptest.NewRecorder()
	handler.ServeHTTP(deleteRecorder, deleteRequest)

	if deleteRecorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d. body: %s", http.StatusNoContent, deleteRecorder.Code, deleteRecorder.Body.String())
	}

	listDeletedRequest := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/products/deleted",
		nil,
	)
	listDeletedRequest.Header.Set(requestIDHeader, "list-deleted-products")

	listDeletedRecorder := httptest.NewRecorder()
	handler.ServeHTTP(listDeletedRecorder, listDeletedRequest)

	if listDeletedRecorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d. body: %s", http.StatusOK, listDeletedRecorder.Code, listDeletedRecorder.Body.String())
	}

	var response DeletedProductListResponse
	if err := json.NewDecoder(listDeletedRecorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode deleted product list response: %v", err)
	}

	if response.Meta.Total != 1 {
		t.Fatalf("expected total %d, got %d", 1, response.Meta.Total)
	}

	if len(response.Data) != 1 {
		t.Fatalf("expected 1 deleted product, got %d", len(response.Data))
	}

	if response.Data[0].ID != deletedProduct.ID {
		t.Fatalf("expected deleted product ID %q, got %q", deletedProduct.ID, response.Data[0].ID)
	}

	if response.Data[0].SKU != "HTTP-DELETED-001" {
		t.Fatalf("expected deleted sku %q, got %q", "HTTP-DELETED-001", response.Data[0].SKU)
	}

	if response.Data[0].DeletedAt == nil {
		t.Fatal("expected deleted_at to be set")
	}
}

func TestListDeletedProductsMethodNotAllowed(t *testing.T) {
	handler := newTestHTTPHandler()

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/products/deleted",
		nil,
	)
	request.Header.Set(requestIDHeader, "list-deleted-method-not-allowed")

	responseRecorder := httptest.NewRecorder()
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, responseRecorder.Code)
	}

	if got := responseRecorder.Header().Get("Allow"); got != http.MethodGet {
		t.Fatalf("expected Allow header %q, got %q", http.MethodGet, got)
	}
}

func TestHardDeleteProduct(t *testing.T) {
	handler := newTestHTTPHandler()

	createBody := []byte(`{
		"sku": "HTTP-HARD-DELETE-001",
		"name": "HTTP Hard Delete",
		"description": "Producto para probar hard delete",
		"price": 100
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

	createdProduct := decodeProductResourceResponse(t, createRecorder)

	deleteRequest := httptest.NewRequest(
		http.MethodDelete,
		"/api/v1/products/"+createdProduct.ID,
		nil,
	)

	deleteRecorder := httptest.NewRecorder()
	handler.ServeHTTP(deleteRecorder, deleteRequest)

	if deleteRecorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d. body: %s", http.StatusNoContent, deleteRecorder.Code, deleteRecorder.Body.String())
	}

	hardDeleteRequest := httptest.NewRequest(
		http.MethodDelete,
		"/api/v1/products/"+createdProduct.ID+"/hard",
		nil,
	)
	hardDeleteRequest.Header.Set(requestIDHeader, "hard-delete-product")

	hardDeleteRecorder := httptest.NewRecorder()
	handler.ServeHTTP(hardDeleteRecorder, hardDeleteRequest)

	if hardDeleteRecorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d. body: %s", http.StatusNoContent, hardDeleteRecorder.Code, hardDeleteRecorder.Body.String())
	}

	listDeletedRequest := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/products/deleted",
		nil,
	)

	listDeletedRecorder := httptest.NewRecorder()
	handler.ServeHTTP(listDeletedRecorder, listDeletedRequest)

	if listDeletedRecorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d. body: %s", http.StatusOK, listDeletedRecorder.Code, listDeletedRecorder.Body.String())
	}

	var listDeletedResponse DeletedProductListResponse
	if err := json.NewDecoder(listDeletedRecorder.Body).Decode(&listDeletedResponse); err != nil {
		t.Fatalf("failed to decode deleted products response: %v", err)
	}

	if listDeletedResponse.Meta.Total != 0 {
		t.Fatalf("expected total deleted %d after hard delete, got %d", 0, listDeletedResponse.Meta.Total)
	}
}

func TestHardDeleteActiveProductReturnsConflict(t *testing.T) {
	handler := newTestHTTPHandler()

	createBody := []byte(`{
		"sku": "HTTP-HARD-ACTIVE-001",
		"name": "HTTP Active Product",
		"description": "Producto activo",
		"price": 100
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

	createdProduct := decodeProductResourceResponse(t, createRecorder)

	hardDeleteRequest := httptest.NewRequest(
		http.MethodDelete,
		"/api/v1/products/"+createdProduct.ID+"/hard",
		nil,
	)
	hardDeleteRequest.Header.Set(requestIDHeader, "hard-delete-active-product")

	hardDeleteRecorder := httptest.NewRecorder()
	handler.ServeHTTP(hardDeleteRecorder, hardDeleteRequest)

	if hardDeleteRecorder.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d. body: %s", http.StatusConflict, hardDeleteRecorder.Code, hardDeleteRecorder.Body.String())
	}

	var response ErrorResponse
	if err := json.NewDecoder(hardDeleteRecorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if response.Error.Code != errorCodeConflict {
		t.Fatalf("expected error code %q, got %q", errorCodeConflict, response.Error.Code)
	}

	if response.Error.Message != "product must be soft deleted before hard delete" {
		t.Fatalf(
			"expected message %q, got %q",
			"product must be soft deleted before hard delete",
			response.Error.Message,
		)
	}
}

func TestHardDeleteProductMethodNotAllowed(t *testing.T) {
	handler := newTestHTTPHandler()

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/products/123/hard",
		nil,
	)
	request.Header.Set(requestIDHeader, "hard-delete-method-not-allowed")

	responseRecorder := httptest.NewRecorder()
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, responseRecorder.Code)
	}

	if got := responseRecorder.Header().Get("Allow"); got != http.MethodDelete {
		t.Fatalf("expected Allow header %q, got %q", http.MethodDelete, got)
	}
}

func TestListProductAuditEventsAfterCreate(t *testing.T) {
	handler := newTestHTTPHandler()

	createBody := []byte(`{
		"sku": "HTTP-AUDIT-001",
		"name": "HTTP Audit",
		"description": "Producto para probar auditoría",
		"price": 100
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

	createdProduct := decodeProductResourceResponse(t, createRecorder)

	auditRequest := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/products/"+createdProduct.ID+"/audit-events",
		nil,
	)
	auditRequest.Header.Set(requestIDHeader, "list-product-audit-events")

	auditRecorder := httptest.NewRecorder()
	handler.ServeHTTP(auditRecorder, auditRequest)

	if auditRecorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d. body: %s", http.StatusOK, auditRecorder.Code, auditRecorder.Body.String())
	}

	var response AuditEventListResponse
	if err := json.NewDecoder(auditRecorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode audit event list response: %v", err)
	}

	if response.Meta.Total != 1 {
		t.Fatalf("expected total %d, got %d", 1, response.Meta.Total)
	}

	if len(response.Data) != 1 {
		t.Fatalf("expected 1 audit event, got %d", len(response.Data))
	}

	event := response.Data[0]

	if event.EventType != product.AuditEventProductCreated {
		t.Fatalf("expected event type %q, got %q", product.AuditEventProductCreated, event.EventType)
	}

	if event.AggregateType != product.AuditAggregateProduct {
		t.Fatalf("expected aggregate type %q, got %q", product.AuditAggregateProduct, event.AggregateType)
	}

	if event.AggregateID != createdProduct.ID {
		t.Fatalf("expected aggregate id %q, got %q", createdProduct.ID, event.AggregateID)
	}

	if event.Payload["sku"] != "HTTP-AUDIT-001" {
		t.Fatalf("expected payload sku %q, got %v", "HTTP-AUDIT-001", event.Payload["sku"])
	}
}

func TestListProductAuditEventsAfterUpdateAndDelete(t *testing.T) {
	handler := newTestHTTPHandler()

	createBody := []byte(`{
		"sku": "HTTP-AUDIT-FLOW-001",
		"name": "HTTP Audit Flow",
		"description": "Producto para flujo de auditoría",
		"price": 100
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

	createdProduct := decodeProductResourceResponse(t, createRecorder)

	patchBody := []byte(`{
		"price": 150
	}`)

	patchRequest := httptest.NewRequest(
		http.MethodPatch,
		"/api/v1/products/"+createdProduct.ID,
		bytes.NewReader(patchBody),
	)
	patchRequest.Header.Set("Content-Type", "application/json")

	patchRecorder := httptest.NewRecorder()
	handler.ServeHTTP(patchRecorder, patchRequest)

	if patchRecorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d. body: %s", http.StatusOK, patchRecorder.Code, patchRecorder.Body.String())
	}

	deleteRequest := httptest.NewRequest(
		http.MethodDelete,
		"/api/v1/products/"+createdProduct.ID,
		nil,
	)

	deleteRecorder := httptest.NewRecorder()
	handler.ServeHTTP(deleteRecorder, deleteRequest)

	if deleteRecorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d. body: %s", http.StatusNoContent, deleteRecorder.Code, deleteRecorder.Body.String())
	}

	auditRequest := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/products/"+createdProduct.ID+"/audit-events",
		nil,
	)

	auditRecorder := httptest.NewRecorder()
	handler.ServeHTTP(auditRecorder, auditRequest)

	if auditRecorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d. body: %s", http.StatusOK, auditRecorder.Code, auditRecorder.Body.String())
	}

	var response AuditEventListResponse
	if err := json.NewDecoder(auditRecorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode audit event list response: %v", err)
	}

	if response.Meta.Total != 3 {
		t.Fatalf("expected total %d, got %d", 3, response.Meta.Total)
	}

	if len(response.Data) != 3 {
		t.Fatalf("expected 3 audit events, got %d", len(response.Data))
	}

	expectedTypes := map[string]bool{
		product.AuditEventProductCreated: false,
		product.AuditEventProductUpdated: false,
		product.AuditEventProductDeleted: false,
	}

	for _, event := range response.Data {
		if _, ok := expectedTypes[event.EventType]; ok {
			expectedTypes[event.EventType] = true
		}
	}

	for eventType, found := range expectedTypes {
		if !found {
			t.Fatalf("expected audit event type %q to be present", eventType)
		}
	}
}

func TestListProductAuditEventsMethodNotAllowed(t *testing.T) {
	handler := newTestHTTPHandler()

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/products/123/audit-events",
		nil,
	)
	request.Header.Set(requestIDHeader, "audit-events-method-not-allowed")

	responseRecorder := httptest.NewRecorder()
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, responseRecorder.Code)
	}

	if got := responseRecorder.Header().Get("Allow"); got != http.MethodGet {
		t.Fatalf("expected Allow header %q, got %q", http.MethodGet, got)
	}
}

func TestListProductAuditEventsInvalidPagination(t *testing.T) {
	handler := newTestHTTPHandler()

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/products/123/audit-events?page=0&page_size=101",
		nil,
	)
	request.Header.Set(requestIDHeader, "audit-events-invalid-pagination")

	responseRecorder := httptest.NewRecorder()
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf(
			"expected status %d, got %d. body: %s",
			http.StatusUnprocessableEntity,
			responseRecorder.Code,
			responseRecorder.Body.String(),
		)
	}

	var response ErrorResponse
	if err := json.NewDecoder(responseRecorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if response.Error.Code != errorCodeValidation {
		t.Fatalf("expected error code %q, got %q", errorCodeValidation, response.Error.Code)
	}

	if len(response.Error.Fields) != 2 {
		t.Fatalf("expected 2 field errors, got %d", len(response.Error.Fields))
	}
}
