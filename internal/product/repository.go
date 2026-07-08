package product

import "context"

type Repository interface {
	List(ctx context.Context, input ListProductsInput) (ListProductsResult, error)
	Get(ctx context.Context, id string) (Product, error)
	GetBySKU(ctx context.Context, sku string) (Product, error)
	Create(ctx context.Context, input CreateProductInput) (Product, error)
	Update(ctx context.Context, id string, input UpdateProductInput) (Product, error)
	Delete(ctx context.Context, id string) error
}
