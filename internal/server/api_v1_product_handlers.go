package server

import (
	"errors"
	"net/http"
	"strings"

	"github.com/r3dp4nd/go-clean-api/internal/product"
)

const apiV1ProductsPrefix = "/api/v1/products/"

func (h *Handler) handleAPIV1Products(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listProducts(w, r)

	case http.MethodPost:
		h.createProduct(w, r)

	default:
		writeMethodNotAllowed(w, r, "GET, POST")
	}
}

func (h *Handler) handleAPIV1ProductByID(w http.ResponseWriter, r *http.Request) {
	id, ok := productIDFromPath(r.URL.Path)
	if !ok {
		writeNotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getProduct(w, r, id)

	case http.MethodPut:
		h.updateProduct(w, r, id)

	case http.MethodDelete:
		h.deleteProduct(w, r, id)

	default:
		writeMethodNotAllowed(w, r, "GET, PUT, DELETE")
	}
}

func (h *Handler) listProducts(w http.ResponseWriter, r *http.Request) {
	items, err := h.productStore.List(r.Context())
	if err != nil {
		h.logger.Error("error listing products", "error", err, "request_id", getRequestID(r.Context()))
		writeInternalError(w, r)
		return
	}

	response := ProductListResponse{
		Data: make([]ProductResponse, 0, len(items)),
	}

	for _, item := range items {
		response.Data = append(response.Data, toProductResponse(item))
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) createProduct(w http.ResponseWriter, r *http.Request) {
	var request CreateProductRequest

	if err := readJSON(r, &request); err != nil {
		writeBadRequest(w, r, "invalid request body")
		return
	}

	if fields := validateCreateProductRequest(request); len(fields) > 0 {
		writeValidationError(w, r, fields)
		return
	}

	item, err := h.productStore.Create(r.Context(), product.CreateProductInput{
		Name:        strings.TrimSpace(request.Name),
		Description: strings.TrimSpace(request.Description),
		Price:       request.Price,
	})
	if err != nil {
		h.logger.Error("error creating product", "error", err, "request_id", getRequestID(r.Context()))
		writeInternalError(w, r)
		return
	}

	writeJSON(w, http.StatusCreated, toProductResponse(item))
}

func (h *Handler) getProduct(w http.ResponseWriter, r *http.Request, id string) {
	item, err := h.productStore.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, product.ErrNotFound) {
			writeError(w, r, http.StatusNotFound, errorCodeNotFound, "product not found")
			return
		}

		h.logger.Error("error getting product", "error", err, "request_id", getRequestID(r.Context()))
		writeInternalError(w, r)
		return
	}

	writeJSON(w, http.StatusOK, toProductResponse(item))
}

func (h *Handler) updateProduct(w http.ResponseWriter, r *http.Request, id string) {
	var request UpdateProductRequest

	if err := readJSON(r, &request); err != nil {
		writeBadRequest(w, r, "invalid request body")
		return
	}

	if fields := validateUpdateProductRequest(request); len(fields) > 0 {
		writeValidationError(w, r, fields)
		return
	}

	item, err := h.productStore.Update(r.Context(), id, product.UpdateProductInput{
		Name:        strings.TrimSpace(request.Name),
		Description: strings.TrimSpace(request.Description),
		Price:       request.Price,
	})
	if err != nil {
		if errors.Is(err, product.ErrNotFound) {
			writeError(w, r, http.StatusNotFound, errorCodeNotFound, "product not found")
			return
		}

		h.logger.Error("error updating product", "error", err, "request_id", getRequestID(r.Context()))
		writeInternalError(w, r)
		return
	}

	writeJSON(w, http.StatusOK, toProductResponse(item))
}

func (h *Handler) deleteProduct(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.productStore.Delete(r.Context(), id); err != nil {
		if errors.Is(err, product.ErrNotFound) {
			writeError(w, r, http.StatusNotFound, errorCodeNotFound, "product not found")
			return
		}

		h.logger.Error("error deleting product", "error", err, "request_id", getRequestID(r.Context()))
		writeInternalError(w, r)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func productIDFromPath(path string) (string, bool) {
	id := strings.TrimPrefix(path, apiV1ProductsPrefix)
	id = strings.TrimSpace(id)

	if id == "" {
		return "", false
	}

	if strings.Contains(id, "/") {
		return "", false
	}

	return id, true
}

func toProductResponse(item product.Product) ProductResponse {
	return ProductResponse{
		ID:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		Price:       item.Price,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}
