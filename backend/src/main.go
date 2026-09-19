package main

import (
	"backend/src/config"
	"backend/src/db"
	"context"
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.LoadConfig()

	dbPool, err := pgxpool.New(context.Background(), cfg.PostgresConnectionUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer dbPool.Close()

	queries := db.New(dbPool)
	if _, err := queries.Ping(context.Background()); err != nil {
		log.Fatal(err)
	}

	fib := fiber.New()

	fib.Use(cors.New())
	fib.Use(logger.New())

	fib.Get("/", func(ctx fiber.Ctx) error {
		return ctx.JSON("Welcome to Meta Checker Backend")
	})

	log.Fatal(fib.Listen(":" + cfg.Port))
}
