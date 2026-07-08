package product

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

var _ Repository = (*PostgresRepository)(nil)

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		pool: pool,
	}
}

func (r *PostgresRepository) List(ctx context.Context, input ListProductsInput) (ListProductsResult, error) {
	normalizedInput, err := normalizeListProductsInput(input)
	if err != nil {
		return ListProductsResult{}, err
	}

	total, err := r.countProducts(ctx, normalizedInput.Search)
	if err != nil {
		return ListProductsResult{}, err
	}

	items, err := r.listProducts(ctx, normalizedInput)
	if err != nil {
		return ListProductsResult{}, err
	}

	return ListProductsResult{
		Items:      items,
		Total:      total,
		Page:       normalizedInput.Page,
		PageSize:   normalizedInput.PageSize,
		TotalPages: calculateTotalPages(total, normalizedInput.PageSize),
		Search:     normalizedInput.Search,
		Sort:       normalizedInput.Sort,
		Order:      normalizedInput.Order,
	}, nil
}

func (r *PostgresRepository) Get(ctx context.Context, id string) (Product, error) {
	const query = `
		SELECT
			id::text,
			name,
			description,
			price::float8,
			created_at,
			updated_at
		FROM products
		WHERE id::text = $1
	`

	var item Product

	err := r.pool.QueryRow(ctx, query, strings.TrimSpace(id)).Scan(
		&item.ID,
		&item.Name,
		&item.Description,
		&item.Price,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Product{}, ErrNotFound
		}

		return Product{}, fmt.Errorf("get product: %w", err)
	}

	return item, nil
}

func (r *PostgresRepository) Create(ctx context.Context, input CreateProductInput) (Product, error) {
	const query = `
		INSERT INTO products (
			name,
			description,
			price
		)
		VALUES ($1, $2, $3)
		RETURNING
			id::text,
			name,
			description,
			price::float8,
			created_at,
			updated_at
	`

	var item Product

	err := r.pool.QueryRow(
		ctx,
		query,
		input.Name,
		input.Description,
		input.Price,
	).Scan(
		&item.ID,
		&item.Name,
		&item.Description,
		&item.Price,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return Product{}, fmt.Errorf("create product: %w", err)
	}

	return item, nil
}

func (r *PostgresRepository) Update(ctx context.Context, id string, input UpdateProductInput) (Product, error) {
	const query = `
		UPDATE products
		SET
			name = $2,
			description = $3,
			price = $4,
			updated_at = NOW()
		WHERE id::text = $1
		RETURNING
			id::text,
			name,
			description,
			price::float8,
			created_at,
			updated_at
	`

	var item Product

	err := r.pool.QueryRow(
		ctx,
		query,
		strings.TrimSpace(id),
		input.Name,
		input.Description,
		input.Price,
	).Scan(
		&item.ID,
		&item.Name,
		&item.Description,
		&item.Price,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Product{}, ErrNotFound
		}

		return Product{}, fmt.Errorf("update product: %w", err)
	}

	return item, nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id string) error {
	const query = `
		DELETE FROM products
		WHERE id::text = $1
	`

	commandTag, err := r.pool.Exec(ctx, query, strings.TrimSpace(id))
	if err != nil {
		return fmt.Errorf("delete product: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *PostgresRepository) countProducts(ctx context.Context, search string) (int, error) {
	search = strings.TrimSpace(search)

	query := `
		SELECT COUNT(*)
		FROM products
	`

	args := make([]any, 0, 1)

	if search != "" {
		query += `
			WHERE name ILIKE '%' || $1 || '%'
			   OR description ILIKE '%' || $1 || '%'
		`

		args = append(args, search)
	}

	var total int

	if err := r.pool.QueryRow(ctx, query, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count products: %w", err)
	}

	return total, nil
}

func (r *PostgresRepository) listProducts(ctx context.Context, input ListProductsInput) ([]Product, error) {
	offset := (input.Page - 1) * input.PageSize

	args := make([]any, 0, 3)

	whereClause := ""
	if input.Search != "" {
		whereClause = `
			WHERE name ILIKE '%' || $1 || '%'
			   OR description ILIKE '%' || $1 || '%'
		`

		args = append(args, input.Search)
	}

	limitPosition := len(args) + 1
	offsetPosition := len(args) + 2

	args = append(args, input.PageSize, offset)

	orderBy := postgresSortExpression(input.Sort)
	orderDirection := postgresSortDirection(input.Order)
	tieBreaker := ", id ASC"

	if input.Sort == SortFieldID {
		tieBreaker = ""
	}

	query := fmt.Sprintf(`
		SELECT
			id::text,
			name,
			description,
			price::float8,
			created_at,
			updated_at
		FROM products
		%s
		ORDER BY %s %s%s
		LIMIT $%d OFFSET $%d
	`,
		whereClause,
		orderBy,
		orderDirection,
		tieBreaker,
		limitPosition,
		offsetPosition,
	)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	defer rows.Close()

	items := make([]Product, 0, input.PageSize)

	for rows.Next() {
		var item Product

		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Description,
			&item.Price,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan product: %w", err)
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate products: %w", err)
	}

	return items, nil
}

func postgresSortExpression(sortField string) string {
	switch sortField {
	case SortFieldName:
		return "lower(name)"
	case SortFieldPrice:
		return "price"
	case SortFieldCreatedAt:
		return "created_at"
	case SortFieldUpdatedAt:
		return "updated_at"
	case SortFieldID:
		fallthrough
	default:
		return "id"
	}
}

func postgresSortDirection(order string) string {
	if order == SortOrderDesc {
		return "DESC"
	}

	return "ASC"
}
