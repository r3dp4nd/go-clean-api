package server

import (
	"errors"
	"net/http"

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
	id, ok := pathParamAfterPrefix(r.URL.Path, apiV1ProductsPrefix)
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
	items, err := h.productService.List(r.Context())
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

	item, err := h.productService.Create(r.Context(), product.CreateProductInput{
		Name:        request.Name,
		Description: request.Description,
		Price:       request.Price,
	})
	if err != nil {
		if validationErr, ok := err.(product.ValidationError); ok {
			writeProductValidationError(w, r, validationErr)
			return
		}

		h.logger.Error("error creating product", "error", err, "request_id", getRequestID(r.Context()))
		writeInternalError(w, r)
		return
	}

	writeJSON(w, http.StatusCreated, toProductResponse(item))
}

func (h *Handler) getProduct(w http.ResponseWriter, r *http.Request, id string) {
	item, err := h.productService.Get(r.Context(), id)
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

	item, err := h.productService.Update(r.Context(), id, product.UpdateProductInput{
		Name:        request.Name,
		Description: request.Description,
		Price:       request.Price,
	})
	if err != nil {
		if validationErr, ok := err.(product.ValidationError); ok {
			writeProductValidationError(w, r, validationErr)
			return
		}

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
	if err := h.productService.Delete(r.Context(), id); err != nil {
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

func writeProductValidationError(w http.ResponseWriter, r *http.Request, validationErr product.ValidationError) {
	fields := make([]FieldError, 0, len(validationErr.Fields))

	for _, field := range validationErr.Fields {
		fields = append(fields, FieldError{
			Field:   field.Field,
			Message: field.Message,
		})
	}

	writeValidationError(w, r, fields)
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
