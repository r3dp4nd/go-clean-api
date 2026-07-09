package product

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	ErrNotFound                  = errors.New("product not found")
	ErrSKUAlreadyExists          = errors.New("product sku already exists")
	ErrProductMustBeDeletedFirst = errors.New("product must be soft deleted before hard delete")
)

type Store struct {
	mu       sync.RWMutex
	products map[string]Product
	nextID   int64
}

var _ Repository = (*Store)(nil)

func NewStore() *Store {
	return &Store{
		products: make(map[string]Product),
		nextID:   1,
	}
}

func (s *Store) List(ctx context.Context, input ListProductsInput) (ListProductsResult, error) {
	if err := ctx.Err(); err != nil {
		return ListProductsResult{}, err
	}

	if input.Page < 1 {
		input.Page = DefaultPage
	}

	if input.PageSize < 1 {
		input.PageSize = DefaultPageSize
	}

	input.Search = strings.TrimSpace(input.Search)
	input.Sort = strings.ToLower(strings.TrimSpace(input.Sort))
	input.Order = strings.ToLower(strings.TrimSpace(input.Order))

	if input.Sort == "" {
		input.Sort = DefaultSort
	}

	if input.Order == "" {
		input.Order = DefaultOrder
	}

	normalizedSearch := strings.ToLower(input.Search)

	s.mu.RLock()
	defer s.mu.RUnlock()

	products := make([]Product, 0, len(s.products))

	for _, item := range s.products {
		if item.DeletedAt != nil {
			continue
		}

		if productMatchesFilters(
			item,
			normalizedSearch,
			input.MinPrice,
			input.MaxPrice,
			input.CreatedFrom,
			input.CreatedTo,
		) {
			products = append(products, item)
		}
	}

	sort.Slice(products, func(i, j int) bool {
		return productLess(products[i], products[j], input.Sort, input.Order)
	})

	total := len(products)
	totalPages := calculateTotalPages(total, input.PageSize)

	offset := (input.Page - 1) * input.PageSize
	if offset >= total {
		return ListProductsResult{
			Items:       []Product{},
			Total:       total,
			Page:        input.Page,
			PageSize:    input.PageSize,
			TotalPages:  totalPages,
			Search:      input.Search,
			Sort:        input.Sort,
			Order:       input.Order,
			MinPrice:    input.MinPrice,
			MaxPrice:    input.MaxPrice,
			CreatedFrom: input.CreatedFrom,
			CreatedTo:   input.CreatedTo,
		}, nil
	}

	end := offset + input.PageSize
	if end > total {
		end = total
	}

	return ListProductsResult{
		Items:       products[offset:end],
		Total:       total,
		Page:        input.Page,
		PageSize:    input.PageSize,
		TotalPages:  totalPages,
		Search:      input.Search,
		Sort:        input.Sort,
		Order:       input.Order,
		MinPrice:    input.MinPrice,
		MaxPrice:    input.MaxPrice,
		CreatedFrom: input.CreatedFrom,
		CreatedTo:   input.CreatedTo,
	}, nil
}

func (s *Store) ListDeleted(ctx context.Context, input ListProductsInput) (ListProductsResult, error) {
	if err := ctx.Err(); err != nil {
		return ListProductsResult{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	normalizedSearch := strings.ToLower(strings.TrimSpace(input.Search))

	products := make([]Product, 0, len(s.products))

	for _, item := range s.products {
		if item.DeletedAt == nil {
			continue
		}

		if productMatchesFilters(
			item,
			normalizedSearch,
			input.MinPrice,
			input.MaxPrice,
			input.CreatedFrom,
			input.CreatedTo,
		) {
			products = append(products, item)
		}
	}

	sortProducts(products, input.Sort, input.Order)

	total := len(products)
	totalPages := calculateTotalPages(total, input.PageSize)

	offset := (input.Page - 1) * input.PageSize
	if offset >= total {
		return ListProductsResult{
			Items:       []Product{},
			Total:       total,
			Page:        input.Page,
			PageSize:    input.PageSize,
			TotalPages:  totalPages,
			Search:      input.Search,
			Sort:        input.Sort,
			Order:       input.Order,
			MinPrice:    input.MinPrice,
			MaxPrice:    input.MaxPrice,
			CreatedFrom: input.CreatedFrom,
			CreatedTo:   input.CreatedTo,
		}, nil
	}

	end := offset + input.PageSize
	if end > total {
		end = total
	}

	return ListProductsResult{
		Items:       products[offset:end],
		Total:       total,
		Page:        input.Page,
		PageSize:    input.PageSize,
		TotalPages:  totalPages,
		Search:      input.Search,
		Sort:        input.Sort,
		Order:       input.Order,
		MinPrice:    input.MinPrice,
		MaxPrice:    input.MaxPrice,
		CreatedFrom: input.CreatedFrom,
		CreatedTo:   input.CreatedTo,
	}, nil
}

func (s *Store) Get(ctx context.Context, id string) (Product, error) {
	if err := ctx.Err(); err != nil {
		return Product{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.products[strings.TrimSpace(id)]
	if !ok || item.DeletedAt != nil {
		return Product{}, ErrNotFound
	}

	return item, nil
}

func (s *Store) GetBySKU(ctx context.Context, sku string) (Product, error) {
	if err := ctx.Err(); err != nil {
		return Product{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, item := range s.products {
		if item.DeletedAt == nil &&
			strings.EqualFold(item.SKU, strings.TrimSpace(sku)) {
			return item, nil
		}
	}

	return Product{}, ErrNotFound
}

func (s *Store) Create(ctx context.Context, input CreateProductInput) (Product, error) {
	if err := ctx.Err(); err != nil {
		return Product{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, existingProduct := range s.products {
		if existingProduct.DeletedAt == nil &&
			strings.EqualFold(existingProduct.SKU, input.SKU) {
			return Product{}, ErrSKUAlreadyExists
		}
	}

	now := time.Now().UTC()
	id := fmt.Sprintf("%d", s.nextID)
	s.nextID++

	item := Product{
		ID:          id,
		SKU:         input.SKU,
		Name:        input.Name,
		Description: input.Description,
		Price:       input.Price,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	s.products[id] = item

	return item, nil
}

func (s *Store) Update(ctx context.Context, id string, input UpdateProductInput) (Product, error) {
	if err := ctx.Err(); err != nil {
		return Product{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.products[id]
	if !ok || item.DeletedAt != nil {
		return Product{}, ErrNotFound
	}

	for _, existingProduct := range s.products {
		if existingProduct.DeletedAt == nil &&
			existingProduct.ID != id &&
			strings.EqualFold(existingProduct.SKU, input.SKU) {
			return Product{}, ErrSKUAlreadyExists
		}
	}

	item.SKU = input.SKU
	item.Name = input.Name
	item.Description = input.Description
	item.Price = input.Price
	item.UpdatedAt = time.Now().UTC()

	s.products[id] = item

	return item, nil
}

func (s *Store) Delete(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	trimmedID := strings.TrimSpace(id)

	item, ok := s.products[trimmedID]
	if !ok || item.DeletedAt != nil {
		return ErrNotFound
	}

	now := time.Now().UTC()

	item.DeletedAt = &now
	item.UpdatedAt = now

	s.products[trimmedID] = item

	return nil
}

func (s *Store) Restore(ctx context.Context, id string) (Product, error) {
	if err := ctx.Err(); err != nil {
		return Product{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	trimmedID := strings.TrimSpace(id)

	item, ok := s.products[trimmedID]
	if !ok {
		return Product{}, ErrNotFound
	}

	if item.DeletedAt == nil {
		return item, nil
	}

	for _, existingProduct := range s.products {
		if existingProduct.ID != trimmedID &&
			existingProduct.DeletedAt == nil &&
			strings.EqualFold(existingProduct.SKU, item.SKU) {
			return Product{}, ErrSKUAlreadyExists
		}
	}

	now := time.Now().UTC()

	item.DeletedAt = nil
	item.UpdatedAt = now

	s.products[trimmedID] = item

	return item, nil
}

func (s *Store) HardDelete(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	trimmedID := strings.TrimSpace(id)

	item, ok := s.products[trimmedID]
	if !ok {
		return ErrNotFound
	}

	if item.DeletedAt == nil {
		return ErrProductMustBeDeletedFirst
	}

	delete(s.products, trimmedID)

	return nil
}

func calculateTotalPages(total int, pageSize int) int {
	if total == 0 {
		return 0
	}

	return (total + pageSize - 1) / pageSize
}

func productIDLess(left string, right string) bool {
	return compareProductID(left, right) < 0
}

func compareProductID(left string, right string) int {
	leftID, leftErr := strconv.ParseInt(left, 10, 64)
	rightID, rightErr := strconv.ParseInt(right, 10, 64)

	if leftErr == nil && rightErr == nil {
		switch {
		case leftID < rightID:
			return -1
		case leftID > rightID:
			return 1
		default:
			return 0
		}
	}

	return strings.Compare(left, right)
}

func productMatchesSearch(item Product, search string) bool {
	sku := strings.ToLower(item.SKU)
	name := strings.ToLower(item.Name)
	description := strings.ToLower(item.Description)

	return strings.Contains(sku, search) ||
		strings.Contains(name, search) ||
		strings.Contains(description, search)
}

func productLess(left Product, right Product, sortField string, order string) bool {
	comparison := compareProducts(left, right, sortField)

	if order == SortOrderDesc {
		return comparison > 0
	}

	return comparison < 0
}

func compareProducts(left Product, right Product, sortField string) int {
	var comparison int

	switch sortField {
	case SortFieldSKU:
		comparison = strings.Compare(
			strings.ToLower(left.SKU),
			strings.ToLower(right.SKU),
		)

	case SortFieldName:
		comparison = strings.Compare(
			strings.ToLower(left.Name),
			strings.ToLower(right.Name),
		)

	case SortFieldPrice:
		comparison = compareFloat64(left.Price, right.Price)

	case SortFieldCreatedAt:
		comparison = compareTime(left.CreatedAt, right.CreatedAt)

	case SortFieldUpdatedAt:
		comparison = compareTime(left.UpdatedAt, right.UpdatedAt)

	case SortFieldID:
		fallthrough

	default:
		comparison = compareProductID(left.ID, right.ID)
	}

	if comparison != 0 {
		return comparison
	}

	return compareProductID(left.ID, right.ID)
}

func compareFloat64(left float64, right float64) int {
	switch {
	case left < right:
		return -1
	case left > right:
		return 1
	default:
		return 0
	}
}

func compareTime(left time.Time, right time.Time) int {
	switch {
	case left.Before(right):
		return -1
	case left.After(right):
		return 1
	default:
		return 0
	}
}

func sortProducts(products []Product, s string, order string) {
	sort.Slice(products, func(i, j int) bool {
		return productLess(products[i], products[j], s, order)
	})
}

func productMatchesFilters(
	item Product,
	search string,
	minPrice *float64,
	maxPrice *float64,
	createdFrom *time.Time,
	createdTo *time.Time,
) bool {
	if search != "" && !productMatchesSearch(item, search) {
		return false
	}

	if minPrice != nil && item.Price < *minPrice {
		return false
	}

	if maxPrice != nil && item.Price > *maxPrice {
		return false
	}

	if createdFrom != nil && item.CreatedAt.Before(*createdFrom) {
		return false
	}

	if createdTo != nil && item.CreatedAt.After(*createdTo) {
		return false
	}

	return true
}
