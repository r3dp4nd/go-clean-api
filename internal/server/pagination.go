package server

import (
	"net/http"
	"strconv"
	"strings"
	"time"

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

	if rawMinPrice := strings.TrimSpace(query.Get("min_price")); rawMinPrice != "" {
		minPrice, err := strconv.ParseFloat(rawMinPrice, 64)
		if err != nil {
			fields = append(fields, FieldError{
				Field:   "min_price",
				Message: "min_price must be a valid number",
			})
		} else if minPrice < 0 {
			fields = append(fields, FieldError{
				Field:   "min_price",
				Message: "min_price must be greater than or equal to zero",
			})
		} else {
			input.MinPrice = &minPrice
		}
	}

	if rawMaxPrice := strings.TrimSpace(query.Get("max_price")); rawMaxPrice != "" {
		maxPrice, err := strconv.ParseFloat(rawMaxPrice, 64)
		if err != nil {
			fields = append(fields, FieldError{
				Field:   "max_price",
				Message: "max_price must be a valid number",
			})
		} else if maxPrice < 0 {
			fields = append(fields, FieldError{
				Field:   "max_price",
				Message: "max_price must be greater than or equal to zero",
			})
		} else {
			input.MaxPrice = &maxPrice
		}
	}

	if input.MinPrice != nil &&
		input.MaxPrice != nil &&
		*input.MinPrice > *input.MaxPrice {
		fields = append(fields, FieldError{
			Field:   "price_range",
			Message: "min_price must be less than or equal to max_price",
		})
	}

	if rawCreatedFrom := strings.TrimSpace(query.Get("created_from")); rawCreatedFrom != "" {
		createdFrom, err := parseProductDateTimeQuery(rawCreatedFrom, false)
		if err != nil {
			fields = append(fields, FieldError{
				Field:   "created_from",
				Message: "created_from must be a valid date in YYYY-MM-DD or RFC3339 format",
			})
		} else {
			input.CreatedFrom = &createdFrom
		}
	}

	if rawCreatedTo := strings.TrimSpace(query.Get("created_to")); rawCreatedTo != "" {
		createdTo, err := parseProductDateTimeQuery(rawCreatedTo, true)
		if err != nil {
			fields = append(fields, FieldError{
				Field:   "created_to",
				Message: "created_to must be a valid date in YYYY-MM-DD or RFC3339 format",
			})
		} else {
			input.CreatedTo = &createdTo
		}
	}

	if input.CreatedFrom != nil &&
		input.CreatedTo != nil &&
		input.CreatedFrom.After(*input.CreatedTo) {
		fields = append(fields, FieldError{
			Field:   "created_range",
			Message: "created_from must be less than or equal to created_to",
		})
	}

	if rawSort := strings.ToLower(strings.TrimSpace(query.Get("sort"))); rawSort != "" {
		if !product.IsSupportedSortField(rawSort) {
			fields = append(fields, FieldError{
				Field:   "sort",
				Message: "sort must be one of: id, sku, name, price, created_at, updated_at",
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

func parseProductDateTimeQuery(value string, endOfDay bool) (time.Time, error) {
	value = strings.TrimSpace(value)

	if parsedValue, err := time.Parse(time.RFC3339, value); err == nil {
		return parsedValue.UTC(), nil
	}

	parsedDate, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, err
	}

	parsedDate = parsedDate.UTC()

	if endOfDay {
		return parsedDate.Add(24 * time.Hour).Add(-time.Nanosecond), nil
	}

	return parsedDate, nil
}
