package server

import "time"

type ProductResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ProductListResponse struct {
	Data []ProductResponse `json:"data"`
	Meta PaginationMeta    `json:"meta"`
}

type PaginationMeta struct {
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
	Total      int    `json:"total"`
	TotalPages int    `json:"total_pages"`
	Search     string `json:"search,omitempty"`
	Sort       string `json:"sort"`
	Order      string `json:"order"`
}

type CreateProductRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

type UpdateProductRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}
