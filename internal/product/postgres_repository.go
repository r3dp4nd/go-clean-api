package product

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type productVisibility string

const (
	productVisibilityActive  productVisibility = "active"
	productVisibilityDeleted productVisibility = "deleted"
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
	return r.listByVisibility(ctx, input, productVisibilityActive)
}

func (r *PostgresRepository) ListDeleted(ctx context.Context, input ListProductsInput) (ListProductsResult, error) {
	return r.listByVisibility(ctx, input, productVisibilityDeleted)
}

func (r *PostgresRepository) Get(ctx context.Context, id string) (Product, error) {
	const query = `
		SELECT
			id::text,
			sku,
			name,
			description,
			price::float8,
			created_at,
			updated_at
		FROM products
		WHERE id = $1::uuid
		  AND deleted_at IS NULL
	`

	var item Product

	err := r.pool.QueryRow(ctx, query, strings.TrimSpace(id)).Scan(
		&item.ID,
		&item.SKU,
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

func (r *PostgresRepository) GetDeleted(ctx context.Context, id string) (Product, error) {
	const query = `
		SELECT
			id::text,
			sku,
			name,
			description,
			price::float8,
			created_at,
			updated_at,
			deleted_at
		FROM products
		WHERE id = $1::uuid
		  AND deleted_at IS NOT NULL
	`

	var item Product

	err := r.pool.QueryRow(ctx, query, strings.TrimSpace(id)).Scan(
		&item.ID,
		&item.SKU,
		&item.Name,
		&item.Description,
		&item.Price,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Product{}, ErrNotFound
		}

		return Product{}, fmt.Errorf("get deleted product: %w", err)
	}

	return item, nil
}

func (r *PostgresRepository) GetBySKU(ctx context.Context, sku string) (Product, error) {
	const query = `
		SELECT
			id::text,
			sku,
			name,
			description,
			price::float8,
			created_at,
			updated_at
		FROM products
		WHERE upper(sku) = upper($1)
		  AND deleted_at IS NULL
	`

	var item Product

	err := r.pool.QueryRow(ctx, query, strings.TrimSpace(sku)).Scan(
		&item.ID,
		&item.SKU,
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

		return Product{}, fmt.Errorf("get product by sku: %w", err)
	}

	return item, nil
}

func (r *PostgresRepository) Create(ctx context.Context, input CreateProductInput) (Product, error) {
	const query = `
		INSERT INTO products (
			sku,
			name,
			description,
			price
		)
		VALUES ($1, $2, $3, $4)
		RETURNING
			id::text,
			sku,
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
		input.SKU,
		input.Name,
		input.Description,
		input.Price,
	).Scan(
		&item.ID,
		&item.SKU,
		&item.Name,
		&item.Description,
		&item.Price,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if isProductSKUUniqueViolation(err) {
			return Product{}, ErrSKUAlreadyExists
		}

		return Product{}, fmt.Errorf("create product: %w", err)
	}

	return item, nil
}

func (r *PostgresRepository) Update(ctx context.Context, id string, input UpdateProductInput) (Product, error) {
	const query = `
		UPDATE products
		SET
			sku = $2,
			name = $3,
			description = $4,
			price = $5,
			updated_at = NOW()
		WHERE id = $1::uuid
		  AND deleted_at IS NULL
		RETURNING
			id::text,
			sku,
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
		input.SKU,
		input.Name,
		input.Description,
		input.Price,
	).Scan(
		&item.ID,
		&item.SKU,
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

		if isProductSKUUniqueViolation(err) {
			return Product{}, ErrSKUAlreadyExists
		}

		return Product{}, fmt.Errorf("update product: %w", err)
	}

	return item, nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id string) error {
	const query = `
		UPDATE products
		SET
			deleted_at = NOW(),
			updated_at = NOW()
		WHERE id = $1::uuid
		  AND deleted_at IS NULL
		RETURNING id
	`

	var deletedID string

	err := r.pool.QueryRow(ctx, query, strings.TrimSpace(id)).Scan(&deletedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}

		return fmt.Errorf("delete product: %w", err)
	}

	return nil
}

func (r *PostgresRepository) Restore(ctx context.Context, id string) (Product, error) {
	const query = `
		UPDATE products
		SET
			deleted_at = NULL,
			updated_at = NOW()
		WHERE id = $1::uuid
		  AND deleted_at IS NOT NULL
		RETURNING
			id::text,
			sku,
			name,
			description,
			price::float8,
			created_at,
			updated_at
	`

	var item Product

	err := r.pool.QueryRow(ctx, query, strings.TrimSpace(id)).Scan(
		&item.ID,
		&item.SKU,
		&item.Name,
		&item.Description,
		&item.Price,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			activeProduct, getErr := r.Get(ctx, id)
			if getErr == nil {
				return activeProduct, nil
			}

			if errors.Is(getErr, ErrNotFound) {
				return Product{}, ErrNotFound
			}

			return Product{}, fmt.Errorf("get product while restoring: %w", getErr)
		}

		if isProductSKUUniqueViolation(err) {
			return Product{}, ErrSKUAlreadyExists
		}

		return Product{}, fmt.Errorf("restore product: %w", err)
	}

	return item, nil
}

func (r *PostgresRepository) HardDelete(ctx context.Context, id string) error {
	const query = `
		DELETE FROM products
		WHERE id = $1::uuid
		  AND deleted_at IS NOT NULL
		RETURNING id
	`

	var deletedID string

	err := r.pool.QueryRow(ctx, query, strings.TrimSpace(id)).Scan(&deletedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			activeProduct, getErr := r.Get(ctx, id)
			if getErr == nil && activeProduct.ID != "" {
				return ErrProductMustBeDeletedFirst
			}

			if errors.Is(getErr, ErrNotFound) {
				return ErrNotFound
			}

			return fmt.Errorf("get product while hard deleting: %w", getErr)
		}

		return fmt.Errorf("hard delete product: %w", err)
	}

	return nil
}

func (r *PostgresRepository) countProducts(
	ctx context.Context,
	input ListProductsInput,
	visibility productVisibility,
) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM products
	`

	whereClause, args := buildProductWhereClause(input, visibility)

	query += whereClause

	var total int

	if err := r.pool.QueryRow(ctx, query, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count products: %w", err)
	}

	return total, nil
}

func (r *PostgresRepository) listProducts(
	ctx context.Context,
	input ListProductsInput,
	visibility productVisibility,
) ([]Product, error) {
	offset := (input.Page - 1) * input.PageSize
	sortExpression := postgresSortExpression(input.Sort)
	sortOrder := postgresSortDirection(input.Order)

	whereClause, args := buildProductWhereClause(input, visibility)

	limitPosition := len(args) + 1
	offsetPosition := len(args) + 2

	args = append(args, input.PageSize, offset)

	query := fmt.Sprintf(`
		SELECT
			id::text,
			sku,
			name,
			description,
			price::float8,
			created_at,
			updated_at,
			deleted_at
		FROM products
		%s
		ORDER BY %s %s, id ASC
		LIMIT $%d OFFSET $%d
	`,
		whereClause,
		sortExpression,
		sortOrder,
		limitPosition,
		offsetPosition,
	)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	defer rows.Close()

	items := make([]Product, 0)

	for rows.Next() {
		var item Product

		if err := rows.Scan(
			&item.ID,
			&item.SKU,
			&item.Name,
			&item.Description,
			&item.Price,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.DeletedAt,
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

func (r *PostgresRepository) listByVisibility(
	ctx context.Context,
	input ListProductsInput,
	visibility productVisibility,
) (ListProductsResult, error) {
	total, err := r.countProducts(ctx, input, visibility)
	if err != nil {
		return ListProductsResult{}, err
	}

	items, err := r.listProducts(ctx, input, visibility)
	if err != nil {
		return ListProductsResult{}, err
	}

	return ListProductsResult{
		Items:       items,
		Total:       total,
		Page:        input.Page,
		PageSize:    input.PageSize,
		TotalPages:  calculateTotalPages(total, input.PageSize),
		Search:      input.Search,
		Sort:        input.Sort,
		Order:       input.Order,
		MinPrice:    input.MinPrice,
		MaxPrice:    input.MaxPrice,
		CreatedFrom: input.CreatedFrom,
		CreatedTo:   input.CreatedTo,
	}, nil
}

func buildProductWhereClause(input ListProductsInput, visibility productVisibility) (string, []any) {
	conditions := make([]string, 0, 6)
	args := make([]any, 0, 5)

	switch visibility {
	case productVisibilityDeleted:
		conditions = append(conditions, "deleted_at IS NOT NULL")
	default:
		conditions = append(conditions, "deleted_at IS NULL")
	}

	if input.Search != "" {
		args = append(args, input.Search)
		position := len(args)

		conditions = append(
			conditions,
			fmt.Sprintf(
				`(sku ILIKE '%%' || $%d || '%%'
				   OR name ILIKE '%%' || $%d || '%%'
				   OR description ILIKE '%%' || $%d || '%%')`,
				position,
				position,
				position,
			),
		)
	}

	if input.MinPrice != nil {
		args = append(args, *input.MinPrice)
		position := len(args)

		conditions = append(
			conditions,
			fmt.Sprintf("price >= $%d", position),
		)
	}

	if input.MaxPrice != nil {
		args = append(args, *input.MaxPrice)
		position := len(args)

		conditions = append(
			conditions,
			fmt.Sprintf("price <= $%d", position),
		)
	}

	if input.CreatedFrom != nil {
		args = append(args, *input.CreatedFrom)
		position := len(args)

		conditions = append(
			conditions,
			fmt.Sprintf("created_at >= $%d", position),
		)
	}

	if input.CreatedTo != nil {
		args = append(args, *input.CreatedTo)
		position := len(args)

		conditions = append(
			conditions,
			fmt.Sprintf("created_at <= $%d", position),
		)
	}

	return "\nWHERE " + strings.Join(conditions, "\n  AND "), args
}

func postgresSortExpression(sortField string) string {
	switch sortField {
	case SortFieldSKU:
		return "lower(sku)"
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

func isProductSKUUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}

	if pgErr.Code != "23505" {
		return false
	}

	return pgErr.ConstraintName == "products_sku_unique" ||
		pgErr.ConstraintName == "idx_products_sku_unique_active"
}
