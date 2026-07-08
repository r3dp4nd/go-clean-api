package product

import "time"

type Product struct {
	ID          string
	SKU         string
	Name        string
	Description string
	Price       float64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CreateProductInput struct {
	SKU         string
	Name        string
	Description string
	Price       float64
}

type UpdateProductInput struct {
	SKU         string
	Name        string
	Description string
	Price       float64
}

type PatchProductInput struct {
	SKU         *string
	Name        *string
	Description *string
	Price       *float64
}

type ListProductsInput struct {
	Page        int
	PageSize    int
	Search      string
	Sort        string
	Order       string
	MinPrice    *float64
	MaxPrice    *float64
	CreatedFrom *time.Time
	CreatedTo   *time.Time
}

type ListProductsResult struct {
	Items       []Product
	Total       int
	Page        int
	PageSize    int
	TotalPages  int
	Search      string
	Sort        string
	Order       string
	MinPrice    *float64
	MaxPrice    *float64
	CreatedFrom *time.Time
	CreatedTo   *time.Time
}

type SKUExistsResult struct {
	SKU    string
	Exists bool
}
