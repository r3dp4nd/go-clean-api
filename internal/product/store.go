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

var ErrNotFound = errors.New("product not found")

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
	normalizedSearch := strings.ToLower(input.Search)

	s.mu.RLock()
	defer s.mu.RUnlock()

	products := make([]Product, 0, len(s.products))

	for _, item := range s.products {
		if normalizedSearch == "" || productMatchesSearch(item, normalizedSearch) {
			products = append(products, item)
		}
	}

	sort.Slice(products, func(i, j int) bool {
		return productIDLess(products[i].ID, products[j].ID)
	})

	total := len(products)
	totalPages := calculateTotalPages(total, input.PageSize)

	offset := (input.Page - 1) * input.PageSize
	if offset >= total {
		return ListProductsResult{
			Items:      []Product{},
			Total:      total,
			Page:       input.Page,
			PageSize:   input.PageSize,
			TotalPages: totalPages,
			Search:     input.Search,
		}, nil
	}

	end := offset + input.PageSize
	if end > total {
		end = total
	}

	return ListProductsResult{
		Items:      products[offset:end],
		Total:      total,
		Page:       input.Page,
		PageSize:   input.PageSize,
		TotalPages: totalPages,
		Search:     input.Search,
	}, nil
}

func (s *Store) Get(ctx context.Context, id string) (Product, error) {
	if err := ctx.Err(); err != nil {
		return Product{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.products[id]
	if !ok {
		return Product{}, ErrNotFound
	}

	return item, nil
}

func (s *Store) Create(ctx context.Context, input CreateProductInput) (Product, error) {
	if err := ctx.Err(); err != nil {
		return Product{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	id := fmt.Sprintf("%d", s.nextID)
	s.nextID++

	item := Product{
		ID:          id,
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
	if !ok {
		return Product{}, ErrNotFound
	}

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

	if _, ok := s.products[id]; !ok {
		return ErrNotFound
	}

	delete(s.products, id)

	return nil
}

func calculateTotalPages(total int, pageSize int) int {
	if total == 0 {
		return 0
	}

	return (total + pageSize - 1) / pageSize
}

func productIDLess(left string, right string) bool {
	leftID, leftErr := strconv.ParseInt(left, 10, 64)
	rightID, rightErr := strconv.ParseInt(right, 10, 64)

	if leftErr == nil && rightErr == nil {
		return leftID < rightID
	}

	return left < right
}

func productMatchesSearch(item Product, search string) bool {
	name := strings.ToLower(item.Name)
	description := strings.ToLower(item.Description)

	return strings.Contains(name, search) ||
		strings.Contains(description, search)
}
