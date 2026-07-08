package server

import "time"

type ProductResponse struct {
	ID          string    `json:"id"`
	SKU         string    `json:"sku"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ProductResourceResponse struct {
	Data ProductResponse `json:"data"`
}

type ProductSKUExistsResponse struct {
	Data ProductSKUExistsData `json:"data"`
}

type ProductSKUExistsData struct {
	SKU    string `json:"sku"`
	Exists bool   `json:"exists"`
}

type ProductListResponse struct {
	Data []ProductResponse `json:"data"`
	Meta PaginationMeta    `json:"meta"`
}

type PaginationMeta struct {
	Page        int        `json:"page"`
	PageSize    int        `json:"page_size"`
	Total       int        `json:"total"`
	TotalPages  int        `json:"total_pages"`
	Search      string     `json:"search,omitempty"`
	Sort        string     `json:"sort"`
	Order       string     `json:"order"`
	MinPrice    *float64   `json:"min_price,omitempty"`
	MaxPrice    *float64   `json:"max_price,omitempty"`
	CreatedFrom *time.Time `json:"created_from,omitempty"`
	CreatedTo   *time.Time `json:"created_to,omitempty"`
}

type CreateProductRequest struct {
	SKU         string  `json:"sku"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

type UpdateProductRequest struct {
	SKU         string  `json:"sku"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}
