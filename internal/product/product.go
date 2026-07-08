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

type ListProductsInput struct {
	Page     int
	PageSize int
	Search   string
	Sort     string
	Order    string
}

type ListProductsResult struct {
	Items      []Product
	Total      int
	Page       int
	PageSize   int
	TotalPages int
	Search     string
	Sort       string
	Order      string
}
