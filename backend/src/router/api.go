package router

import (
	"backend/src/app"

	"github.com/gofiber/fiber/v3"
)

func InitApiRoutes(fib *fiber.App, a *app.Application) {
	fib.Get("/api/hello", helloApiHandler())

	fib.Get("/api/lol-data/champions", getChampionsApiHandler(a))
	fib.Get("/api/lol-data/items", getItemsApiHandler(a))
}

func helloApiHandler() func(ctx fiber.Ctx) error {
	return func(ctx fiber.Ctx) error {
		return ctx.JSON("hello world")
	}
}

func getChampionsApiHandler(a *app.Application) func(ctx fiber.Ctx) error {
	return func(ctx fiber.Ctx) error {
		champs, err := a.GetChampions(ctx.Context())
		if err != nil {
			return err
		}
		return ctx.JSON(champs)
	}
}

func getItemsApiHandler(a *app.Application) func(ctx fiber.Ctx) error {
	return func(ctx fiber.Ctx) error {
		items, err := a.GetItems(ctx.Context())
		if err != nil {
			return err
		}
		return ctx.JSON(items)
	}
}
