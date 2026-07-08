package product

import (
	"context"
	"errors"
	"fmt"
	"sort"
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

func (s *Store) List(ctx context.Context) ([]Product, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	products := make([]Product, 0, len(s.products))

	for _, item := range s.products {
		products = append(products, item)
	}

	sort.Slice(products, func(i, j int) bool {
		return products[i].ID < products[j].ID
	})

	return products, nil
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
