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

	return input, fields
}
