package server

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/r3dp4nd/go-clean-api/internal/product"
)

func readProductListQuery(r *http.Request) (product.ListProductsInput, []FieldError) {
	query := r.URL.Query()

	input := product.ListProductsInput{
		Page:     product.DefaultPage,
		PageSize: product.DefaultPageSize,
		Sort:     product.DefaultSort,
		Order:    product.DefaultOrder,
	}

	var fields []FieldError

	if rawPage := strings.TrimSpace(query.Get("page")); rawPage != "" {
		page, err := strconv.Atoi(rawPage)
		if err != nil || page < 1 {
			fields = append(fields, FieldError{
				Field:   "page",
				Message: "page must be a positive integer",
			})
		} else {
			input.Page = page
		}
	}

	if rawPageSize := strings.TrimSpace(query.Get("page_size")); rawPageSize != "" {
		pageSize, err := strconv.Atoi(rawPageSize)
		if err != nil || pageSize < 1 {
			fields = append(fields, FieldError{
				Field:   "page_size",
				Message: "page_size must be a positive integer",
			})
		} else if pageSize > product.MaxPageSize {
			fields = append(fields, FieldError{
				Field:   "page_size",
				Message: "page_size must be less than or equal to 100",
			})
		} else {
			input.PageSize = pageSize
		}
	}

	if rawSearch := strings.TrimSpace(query.Get("search")); rawSearch != "" {
		if len(rawSearch) > product.MaxSearchLength {
			fields = append(fields, FieldError{
				Field:   "search",
				Message: "search must be less than or equal to 120 characters",
			})
		} else {
			input.Search = rawSearch
		}
	}

	if rawSort := strings.ToLower(strings.TrimSpace(query.Get("sort"))); rawSort != "" {
		if !product.IsSupportedSortField(rawSort) {
			fields = append(fields, FieldError{
				Field:   "sort",
				Message: "sort must be one of: id, name, price, created_at, updated_at",
			})
		} else {
			input.Sort = rawSort
		}
	}

	if rawOrder := strings.ToLower(strings.TrimSpace(query.Get("order"))); rawOrder != "" {
		if !product.IsSupportedSortOrder(rawOrder) {
			fields = append(fields, FieldError{
				Field:   "order",
				Message: "order must be one of: asc, desc",
			})
		} else {
			input.Order = rawOrder
		}
	}

	return input, fields
}
