package server

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/r3dp4nd/go-clean-api/internal/audit"
)

func readAuditEventListQuery(r *http.Request) (audit.ListEventsInput, []FieldError) {
	query := r.URL.Query()

	var fieldErrors []FieldError

	input := audit.ListEventsInput{
		Page:     audit.DefaultPage,
		PageSize: audit.DefaultPageSize,
	}

	if rawPage := strings.TrimSpace(query.Get("page")); rawPage != "" {
		page, err := strconv.Atoi(rawPage)
		if err != nil || page <= 0 {
			fieldErrors = append(fieldErrors, FieldError{
				Field:   "page",
				Message: "page must be a positive integer",
			})
		} else {
			input.Page = page
		}
	}

	if rawPageSize := strings.TrimSpace(query.Get("page_size")); rawPageSize != "" {
		pageSize, err := strconv.Atoi(rawPageSize)
		if err != nil || pageSize <= 0 {
			fieldErrors = append(fieldErrors, FieldError{
				Field:   "page_size",
				Message: "page_size must be a positive integer",
			})
		} else if pageSize > audit.MaxPageSize {
			fieldErrors = append(fieldErrors, FieldError{
				Field:   "page_size",
				Message: "page_size must be less than or equal to 100",
			})
		} else {
			input.PageSize = pageSize
		}
	}

	return input, fieldErrors
}
