package product

import "time"

type Product struct {
	ID          string
	Name        string
	Description string
	Price       float64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CreateProductInput struct {
	Name        string
	Description string
	Price       float64
}

type UpdateProductInput struct {
	Name        string
	Description string
	Price       float64
}
