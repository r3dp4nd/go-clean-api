package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

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

	if shouldTruncate {
		if _, err := postgresPool.Exec(ctx, "TRUNCATE TABLE products"); err != nil {
			logger.Error("error truncating products table", "error", err)
			os.Exit(1)
		}

		logger.Info("products table truncated")
	}

	productRepository := product.NewPostgresRepository(postgresPool)
	productService := product.NewService(productRepository)

	seedProducts := []product.CreateProductInput{
		{
			Name:        "Laptop Basic",
			Description: "Laptop para oficina y navegación diaria",
			Price:       2500,
		},
		{
			Name:        "Laptop Pro",
			Description: "Laptop para desarrollo backend con Go",
			Price:       4500,
		},
		{
			Name:        "Laptop Air",
			Description: "Laptop ligera para trabajo remoto",
			Price:       3500,
		},
		{
			Name:        "Mouse Wireless",
			Description: "Mouse inalámbrico ergonómico",
			Price:       120,
		},
		{
			Name:        "Mechanical Keyboard",
			Description: "Teclado mecánico para programación",
			Price:       250,
		},
		{
			Name:        "Monitor 27",
			Description: "Monitor de 27 pulgadas para productividad",
			Price:       1200,
		},
		{
			Name:        "USB-C Hub",
			Description: "Hub USB-C con HDMI, USB y lector SD",
			Price:       180,
		},
		{
			Name:        "Headphones",
			Description: "Audífonos para reuniones y concentración",
			Price:       300,
		},
		{
			Name:        "Dock Station",
			Description: "Dock station para setup profesional",
			Price:       700,
		},
		{
			Name:        "Webcam HD",
			Description: "Cámara web HD para videollamadas",
			Price:       220,
		},
	}

	for _, input := range seedProducts {
		createdProduct, err := productService.Create(ctx, input)
		if err != nil {
			logger.Error(
				"error creating seed product",
				"product_name", input.Name,
				"error", err,
			)
			os.Exit(1)
		}

		logger.Info(
			"seed product created",
			"id", createdProduct.ID,
			"name", createdProduct.Name,
			"price", createdProduct.Price,
		)
	}

	logger.Info(
		"products seed completed",
		"total_products", len(seedProducts),
	)
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
