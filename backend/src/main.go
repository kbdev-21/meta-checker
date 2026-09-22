package main

import (
	"backend/src/app"
	"backend/src/config"
	"backend/src/router"
	"context"
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
)

func main() {
	cfg := config.LoadConfig()

	a, err := app.NewApplication(context.Background(), cfg)
	if err != nil {
		log.Fatal(err)
	}

	//sched.StartUpdateLolDataSched(a)
	//sched.StartCrawlLolMatchesSched(a)

	fib := fiber.New()

	fib.Use(cors.New())
	fib.Use(logger.New())

	router.InitLolDataApiRoutes(fib, a)
	router.InitLolApiRoutes(fib, a)

	log.Println("Server started")
	log.Fatal(fib.Listen(":" + cfg.Port))
}
