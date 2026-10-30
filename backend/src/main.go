package main

import (
	"backend/src/app"
	"backend/src/config"
	"backend/src/router"
	"backend/src/sched"
	"context"
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
)

func main() {
	cfg := config.LoadConfig()

	app, err := app.NewApplication(context.Background(), cfg)
	if err != nil {
		log.Fatal(err)
	}

	sched.StartUpdateLolDataSched(app)
	sched.StartCrawlLolMatchesSched(app)
	sched.StartCleanupLolMatchesSched(app)

	fib := fiber.New()

	fib.Use(cors.New())
	fib.Use(logger.New())
	fib.Use(compress.New())

	fib.Get("/", func(ctx fiber.Ctx) error {
		return ctx.JSON("Welcome to MetaChecker API")
	})

	router.InitLolDataApiRoutes(fib, app)
	router.InitLolApiRoutes(fib, app)

	log.Println("Server started")
	log.Fatal(fib.Listen(":" + cfg.Port))
}
