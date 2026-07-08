package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/r3dp4nd/go-clean-api/internal/config"
	"github.com/r3dp4nd/go-clean-api/internal/database"
	applogger "github.com/r3dp4nd/go-clean-api/internal/logger"
	"github.com/r3dp4nd/go-clean-api/internal/product"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	logger := applogger.New(cfg.Log)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	logger.Info(
		"starting products seed",
		"database_host", cfg.Database.Host,
		"database_port", cfg.Database.Port,
		"database_name", cfg.Database.Name,
	)

	postgresPool, err := database.OpenPostgresPool(ctx, cfg.Database.DSN)
	if err != nil {
		logger.Error("error connecting to postgres", "error", err)
		os.Exit(1)
	}
	defer postgresPool.Close()

	shouldTruncate := getEnvAsBool("SEED_PRODUCTS_TRUNCATE", true)

	logger.Info("seed config loaded", "truncate_products", shouldTruncate)

	if shouldTruncate {
		if _, err := postgresPool.Exec(ctx, "TRUNCATE TABLE products"); err != nil {
			logger.Error("error truncating products table", "error", err)
			os.Exit(1)
		}

		logger.Info("products table truncated")
	}

	seedProducts := []product.CreateProductInput{
		{
			SKU:         "LAPTOP-BASIC",
			Name:        "Laptop Basic",
			Description: "Laptop para oficina y navegación diaria",
			Price:       2500,
		},
		{
			SKU:         "LAPTOP-PRO",
			Name:        "Laptop Pro",
			Description: "Laptop para desarrollo backend con Go",
			Price:       4500,
		},
		{
			SKU:         "LAPTOP-AIR",
			Name:        "Laptop Air",
			Description: "Laptop ligera para trabajo remoto",
			Price:       3500,
		},
		{
			SKU:         "MOUSE-WIRELESS",
			Name:        "Mouse Wireless",
			Description: "Mouse inalámbrico ergonómico",
			Price:       120,
		},
		{
			SKU:         "KEYBOARD-MECH",
			Name:        "Mechanical Keyboard",
			Description: "Teclado mecánico para programación",
			Price:       250,
		},
		{
			SKU:         "MONITOR-27",
			Name:        "Monitor 27",
			Description: "Monitor de 27 pulgadas para productividad",
			Price:       1200,
		},
		{
			SKU:         "USB-C-HUB",
			Name:        "USB-C Hub",
			Description: "Hub USB-C con HDMI, USB y lector SD",
			Price:       180,
		},
		{
			SKU:         "HEADPHONES",
			Name:        "Headphones",
			Description: "Audífonos para reuniones y concentración",
			Price:       300,
		},
		{
			SKU:         "DOCK-STATION",
			Name:        "Dock Station",
			Description: "Dock station para setup profesional",
			Price:       700,
		},
		{
			SKU:         "WEBCAM-HD",
			Name:        "Webcam HD",
			Description: "Cámara web HD para videollamadas",
			Price:       220,
		},
	}

	for _, input := range seedProducts {
		createdProduct, err := upsertSeedProduct(ctx, postgresPool, input)
		if err != nil {
			logger.Error(
				"error upserting seed product",
				"sku", input.SKU,
				"product_name", input.Name,
				"error", err,
			)
			os.Exit(1)
		}

		logger.Info(
			"seed product upserted",
			"id", createdProduct.ID,
			"sku", createdProduct.SKU,
			"name", createdProduct.Name,
			"price", createdProduct.Price,
		)
	}

	logger.Info(
		"products seed completed",
		"total_products", len(seedProducts),
	)
}

func upsertSeedProduct(
	ctx context.Context,
	pool *pgxpool.Pool,
	input product.CreateProductInput,
) (product.Product, error) {
	const query = `
		INSERT INTO products (
			sku,
			name,
			description,
			price
		)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (sku)
		DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			price = EXCLUDED.price,
			updated_at = NOW()
		RETURNING
			id::text,
			sku,
			name,
			description,
			price::float8,
			created_at,
			updated_at
	`

	var item product.Product

	err := pool.QueryRow(
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
		return product.Product{}, err
	}

	return item, nil
}

func getEnvAsBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsedValue, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return parsedValue
}
