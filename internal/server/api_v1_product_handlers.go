package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/r3dp4nd/go-clean-api/internal/audit"
	"github.com/r3dp4nd/go-clean-api/internal/product"
)

const (
	apiV1ProductsPrefix           = "/api/v1/products/"
	apiV1ProductsSKUPrefix        = "/api/v1/products/sku/"
	apiV1ProductRestoreSuffix     = "/restore"
	apiV1ProductHardDeleteSuffix  = "/hard"
	apiV1ProductAuditEventsSuffix = "/audit-events"
)

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
	pathValue := strings.TrimPrefix(r.URL.Path, apiV1ProductsPrefix)

	if pathValue == "" {
		writeNotFound(w, r)
		return
	}

	if strings.HasSuffix(pathValue, apiV1ProductRestoreSuffix) {
		id := strings.TrimSuffix(pathValue, apiV1ProductRestoreSuffix)
		id = strings.TrimSuffix(id, "/")

		if id == "" || strings.Contains(id, "/") {
			writeNotFound(w, r)
			return
		}

		if r.Method != http.MethodPost {
			writeMethodNotAllowed(w, r, http.MethodPost)
			return
		}

		h.restoreProduct(w, r, id)
		return
	}

	if strings.HasSuffix(pathValue, apiV1ProductHardDeleteSuffix) {
		id := strings.TrimSuffix(pathValue, apiV1ProductHardDeleteSuffix)
		id = strings.TrimSuffix(id, "/")

		if id == "" || strings.Contains(id, "/") {
			writeNotFound(w, r)
			return
		}

		if r.Method != http.MethodDelete {
			writeMethodNotAllowed(w, r, http.MethodDelete)
			return
		}

		h.hardDeleteProduct(w, r, id)
		return
	}

	if strings.HasSuffix(pathValue, apiV1ProductAuditEventsSuffix) {
		id := strings.TrimSuffix(pathValue, apiV1ProductAuditEventsSuffix)
		id = strings.TrimSuffix(id, "/")

		if id == "" || strings.Contains(id, "/") {
			writeNotFound(w, r)
			return
		}

		if r.Method != http.MethodGet {
			writeMethodNotAllowed(w, r, http.MethodGet)
			return
		}

		h.listProductAuditEvents(w, r, id)
		return
	}

	if strings.Contains(pathValue, "/") {
		writeNotFound(w, r)
		return
	}

	id := pathValue

	switch r.Method {
	case http.MethodGet:
		h.getProduct(w, r, id)

	case http.MethodPut:
		h.updateProduct(w, r, id)

	case http.MethodPatch:
		h.patchProduct(w, r, id)

	case http.MethodDelete:
		h.deleteProduct(w, r, id)

	default:
		writeMethodNotAllowed(
			w,
			r,
			fmt.Sprintf("%s,%s,%s,%s",
				http.MethodGet,
				http.MethodPut,
				http.MethodPatch,
				http.MethodDelete,
			),
		)
	}
}

func (h *Handler) handleAPIV1ProductBySKU(w http.ResponseWriter, r *http.Request) {
	sku, ok := pathParamAfterPrefix(r.URL.Path, apiV1ProductsSKUPrefix)
	if !ok {
		writeNotFound(w, r)
		return
	}

	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r, http.MethodGet)
		return
	}

	h.getProductBySKU(w, r, sku)
}

func (h *Handler) handleAPIV1ProductExists(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r, http.MethodGet)
		return
	}

	sku := strings.TrimSpace(r.URL.Query().Get("sku"))
	if sku == "" {
		writeValidationError(w, r, []FieldError{
			{
				Field:   "sku",
				Message: "sku is required",
			},
		})
		return
	}

	result, err := h.productService.SKUExists(r.Context(), sku)
	if err != nil {
		h.logger.Error(
			"error checking product sku existence",
			"error", err,
			"request_id", getRequestID(r.Context()),
		)

		writeInternalError(w, r)
		return
	}

	response := ProductSKUExistsResponse{
		Data: ProductSKUExistsData{
			SKU:    result.SKU,
			Exists: result.Exists,
		},
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) handleAPIV1DeletedProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r, http.MethodGet)
		return
	}

	h.listDeletedProducts(w, r)
}

func (h *Handler) listProducts(w http.ResponseWriter, r *http.Request) {
	input, fields := readProductListQuery(r)
	if len(fields) > 0 {
		writeValidationError(w, r, fields)
		return
	}

	result, err := h.productService.List(r.Context(), input)
	if err != nil {
		if validationErr, ok := err.(product.ValidationError); ok {
			writeProductValidationError(w, r, validationErr)
			return
		}

		h.logger.Error("error listing products", "error", err, "request_id", getRequestID(r.Context()))
		writeInternalError(w, r)
		return
	}

	response := ProductListResponse{
		Data: make([]ProductResponse, 0, len(result.Items)),
		Meta: PaginationMeta{
			Page:        result.Page,
			PageSize:    result.PageSize,
			Total:       result.Total,
			TotalPages:  result.TotalPages,
			Search:      result.Search,
			Sort:        result.Sort,
			Order:       result.Order,
			MinPrice:    result.MinPrice,
			MaxPrice:    result.MaxPrice,
			CreatedFrom: result.CreatedFrom,
			CreatedTo:   result.CreatedTo,
		},
	}

	for _, item := range result.Items {
		response.Data = append(response.Data, toProductResponse(item))
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) listDeletedProducts(w http.ResponseWriter, r *http.Request) {
	input, fieldErrors := readProductListQuery(r)
	if len(fieldErrors) > 0 {
		writeValidationError(w, r, fieldErrors)
		return
	}

	result, err := h.productService.ListDeleted(r.Context(), input)
	if err != nil {
		if validationErr, ok := err.(product.ValidationError); ok {
			writeProductValidationError(w, r, validationErr)
			return
		}

		h.logger.Error(
			"error listing deleted products",
			"error", err,
			"request_id", getRequestID(r.Context()),
		)

		writeInternalError(w, r)
		return
	}

	response := DeletedProductListResponse{
		Data: toDeletedProductResponses(result.Items),
		Meta: PaginationMeta{
			Page:        result.Page,
			PageSize:    result.PageSize,
			Total:       result.Total,
			TotalPages:  result.TotalPages,
			Search:      result.Search,
			Sort:        result.Sort,
			Order:       result.Order,
			MinPrice:    result.MinPrice,
			MaxPrice:    result.MaxPrice,
			CreatedFrom: result.CreatedFrom,
			CreatedTo:   result.CreatedTo,
		},
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) createProduct(w http.ResponseWriter, r *http.Request) {
	var request CreateProductRequest

	if err := readJSON(r, &request); err != nil {
		writeBadRequest(w, r, "invalid request body")
		return
	}

	ctx := auditContextFromRequest(r)

	item, err := h.productService.Create(ctx, product.CreateProductInput{
		SKU:         request.SKU,
		Name:        request.Name,
		Description: request.Description,
		Price:       request.Price,
	})
	if err != nil {
		if validationErr, ok := err.(product.ValidationError); ok {
			writeProductValidationError(w, r, validationErr)
			return
		}

		if errors.Is(err, product.ErrSKUAlreadyExists) {
			writeError(
				w,
				r,
				http.StatusConflict,
				errorCodeConflict,
				"product sku already exists",
			)
			return
		}

		h.logger.Error("error creating product", "error", err, "request_id", getRequestID(r.Context()))
		writeInternalError(w, r)
		return
	}

	writeJSON(w, http.StatusCreated, toProductResourceResponse(item))
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

	writeJSON(w, http.StatusOK, toProductResourceResponse(item))
}

func (h *Handler) getProductBySKU(w http.ResponseWriter, r *http.Request, sku string) {
	item, err := h.productService.GetBySKU(r.Context(), sku)
	if err != nil {
		if errors.Is(err, product.ErrNotFound) {
			writeError(w, r, http.StatusNotFound, errorCodeNotFound, "product not found")
			return
		}

		h.logger.Error("error getting product by sku", "error", err, "request_id", getRequestID(r.Context()))
		writeInternalError(w, r)
		return
	}

	writeJSON(w, http.StatusOK, toProductResourceResponse(item))
}

func (h *Handler) updateProduct(w http.ResponseWriter, r *http.Request, id string) {
	var request UpdateProductRequest

	if err := readJSON(r, &request); err != nil {
		writeBadRequest(w, r, "invalid request body")
		return
	}

	ctx := auditContextFromRequest(r)

	item, err := h.productService.Update(ctx, id, product.UpdateProductInput{
		SKU:         request.SKU,
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

		if errors.Is(err, product.ErrSKUAlreadyExists) {
			writeError(
				w,
				r,
				http.StatusConflict,
				errorCodeConflict,
				"product sku already exists",
			)
			return
		}

		h.logger.Error("error updating product", "error", err, "request_id", getRequestID(r.Context()))
		writeInternalError(w, r)
		return
	}

	writeJSON(w, http.StatusOK, toProductResourceResponse(item))
}

func (h *Handler) patchProduct(w http.ResponseWriter, r *http.Request, id string) {
	var request PatchProductRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, r, http.StatusBadRequest, errorCodeInvalidRequest, "invalid JSON body")
		return
	}

	ctx := auditContextFromRequest(r)

	item, err := h.productService.Patch(ctx, id, product.PatchProductInput{
		SKU:         request.SKU,
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

		if errors.Is(err, product.ErrSKUAlreadyExists) {
			writeError(
				w,
				r,
				http.StatusConflict,
				errorCodeConflict,
				"product sku already exists",
			)
			return
		}

		h.logger.Error("error patching product", "error", err, "request_id", getRequestID(r.Context()))
		writeInternalError(w, r)
		return
	}

	writeJSON(w, http.StatusOK, toProductResourceResponse(item))
}

func (h *Handler) deleteProduct(w http.ResponseWriter, r *http.Request, id string) {
	ctx := auditContextFromRequest(r)

	if err := h.productService.Delete(ctx, id); err != nil {
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

func (h *Handler) restoreProduct(w http.ResponseWriter, r *http.Request, id string) {
	ctx := auditContextFromRequest(r)

	item, err := h.productService.Restore(ctx, id)
	if err != nil {
		if errors.Is(err, product.ErrNotFound) {
			writeError(w, r, http.StatusNotFound, errorCodeNotFound, "product not found")
			return
		}

		if errors.Is(err, product.ErrSKUAlreadyExists) {
			writeError(
				w,
				r,
				http.StatusConflict,
				errorCodeConflict,
				"product sku already exists",
			)
			return
		}

		h.logger.Error("error restoring product", "error", err, "request_id", getRequestID(r.Context()))
		writeInternalError(w, r)
		return
	}

	writeJSON(w, http.StatusOK, toProductResourceResponse(item))
}

func (h *Handler) hardDeleteProduct(w http.ResponseWriter, r *http.Request, id string) {
	ctx := auditContextFromRequest(r)

	err := h.productService.HardDelete(ctx, id)
	if err != nil {
		if errors.Is(err, product.ErrNotFound) {
			writeError(w, r, http.StatusNotFound, errorCodeNotFound, "product not found")
			return
		}

		if errors.Is(err, product.ErrProductMustBeDeletedFirst) {
			writeError(
				w,
				r,
				http.StatusConflict,
				errorCodeConflict,
				"product must be soft deleted before hard delete",
			)
			return
		}

		h.logger.Error(
			"error hard deleting product",
			"error", err,
			"request_id", getRequestID(r.Context()),
		)

		writeInternalError(w, r)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) listProductAuditEvents(w http.ResponseWriter, r *http.Request, id string) {
	input, fieldErrors := readAuditEventListQuery(r)
	if len(fieldErrors) > 0 {
		writeValidationError(w, r, fieldErrors)
		return
	}

	result, err := h.auditReader.ListByAggregate(
		r.Context(),
		product.AuditAggregateProduct,
		id,
		input,
	)
	if err != nil {
		h.logger.Error(
			"error listing product audit events",
			"error", err,
			"request_id", getRequestID(r.Context()),
		)

		writeInternalError(w, r)
		return
	}

	response := AuditEventListResponse{
		Data: toAuditEventResponses(result.Items),
		Meta: PaginationMeta{
			Page:       result.Page,
			PageSize:   result.PageSize,
			Total:      result.Total,
			TotalPages: result.TotalPages,
		},
	}

	writeJSON(w, http.StatusOK, response)
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
		SKU:         item.SKU,
		Name:        item.Name,
		Description: item.Description,
		Price:       item.Price,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}

func toProductResourceResponse(item product.Product) ProductResourceResponse {
	return ProductResourceResponse{
		Data: toProductResponse(item),
	}
}

func toDeletedProductResponse(item product.Product) DeletedProductResponse {
	return DeletedProductResponse{
		ID:          item.ID,
		SKU:         item.SKU,
		Name:        item.Name,
		Description: item.Description,
		Price:       item.Price,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
		DeletedAt:   item.DeletedAt,
	}
}

func toDeletedProductResponses(items []product.Product) []DeletedProductResponse {
	responses := make([]DeletedProductResponse, 0, len(items))

	for _, item := range items {
		responses = append(responses, toDeletedProductResponse(item))
	}

	return responses
}

func toAuditEventResponse(item audit.Event) AuditEventResponse {
	payload := item.Payload
	if payload == nil {
		payload = map[string]any{}
	}

	return AuditEventResponse{
		ID:            item.ID,
		EventType:     item.Type,
		AggregateType: item.AggregateType,
		AggregateID:   item.AggregateID,
		Payload:       payload,
		CreatedAt:     item.CreatedAt,
	}
}

func toAuditEventResponses(items []audit.Event) []AuditEventResponse {
	responses := make([]AuditEventResponse, 0, len(items))

	for _, item := range items {
		responses = append(responses, toAuditEventResponse(item))
	}

	return responses
}
