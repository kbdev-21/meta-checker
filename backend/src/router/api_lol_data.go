package router

import (
	"log"

	"backend/src/app"

	"github.com/gofiber/fiber/v3"
)

func InitLolDataApiRoutes(fib *fiber.App, a *app.Application) {
	fib.Get("/api/lol-data/champions", getChampionsApiHandler(a))
	fib.Get("/api/lol-data/items", getItemsApiHandler(a))
	fib.Get("/api/lol-data/spells", getSpellsApiHandler(a))
	fib.Get("/api/lol-data/runes", getRunesApiHandler(a))
}

func getChampionsApiHandler(a *app.Application) func(ctx fiber.Ctx) error {
	return func(ctx fiber.Ctx) error {
		champs, err := a.GetChampions(ctx.Context())
		if err != nil {
			log.Printf("get champions: %v", err)
			return err
		}
		return ctx.JSON(champs)
	}
}

func getItemsApiHandler(a *app.Application) func(ctx fiber.Ctx) error {
	return func(ctx fiber.Ctx) error {
		items, err := a.GetItems(ctx.Context())
		if err != nil {
			log.Printf("get items: %v", err)
			return err
		}
		return ctx.JSON(items)
	}
}

func getSpellsApiHandler(a *app.Application) func(ctx fiber.Ctx) error {
	return func(ctx fiber.Ctx) error {
		spells, err := a.GetSpells(ctx.Context())
		if err != nil {
			log.Printf("get spells: %v", err)
			return err
		}
		return ctx.JSON(spells)
	}
}

// Trả cả cây rune lẫn rune: dòng có styleId = null là cây.
func getRunesApiHandler(a *app.Application) func(ctx fiber.Ctx) error {
	return func(ctx fiber.Ctx) error {
		runes, err := a.GetRunes(ctx.Context())
		if err != nil {
			log.Printf("get runes: %v", err)
			return err
		}
		return ctx.JSON(runes)
	}
}
