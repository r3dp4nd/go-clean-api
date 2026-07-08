package server

import "strings"

const (
	maxProductNameLength        = 120
	maxProductDescriptionLength = 500
)

func validateCreateProductRequest(request CreateProductRequest) []FieldError {
	return validateProductPayload(
		request.Name,
		request.Description,
		request.Price,
	)
}

func validateUpdateProductRequest(request UpdateProductRequest) []FieldError {
	return validateProductPayload(
		request.Name,
		request.Description,
		request.Price,
	)
}

func validateProductPayload(name string, description string, price float64) []FieldError {
	var fields []FieldError

	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)

	if name == "" {
		fields = append(fields, FieldError{
			Field:   "name",
			Message: "name is required",
		})
	}

	if len(name) > maxProductNameLength {
		fields = append(fields, FieldError{
			Field:   "name",
			Message: "name must be less than or equal to 120 characters",
		})
	}

	if len(description) > maxProductDescriptionLength {
		fields = append(fields, FieldError{
			Field:   "description",
			Message: "description must be less than or equal to 500 characters",
		})
	}

	if price < 0 {
		fields = append(fields, FieldError{
			Field:   "price",
			Message: "price must be greater than or equal to zero",
		})
	}

	return fields
}
