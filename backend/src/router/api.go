package router

import (
	"log"
	"net/url"
	"strings"

	"backend/src/app"
	"backend/src/external"

	"github.com/gofiber/fiber/v3"
)

const playerSearchLimit = 20

func InitApiRoutes(fib *fiber.App, a *app.Application) {
	fib.Get("/api/hello", helloApiHandler())

	fib.Get("/api/lol-data/champions", getChampionsApiHandler(a))
	fib.Get("/api/lol-data/items", getItemsApiHandler(a))

	fib.Get("/api/lol/players", searchPlayersApiHandler(a))
	fib.Get("/api/lol/players/by-info/:server/:name/:tag", findPlayerByInfoApiHandler(a))
	fib.Post("/api/lol/players/by-info/:server/:name/:tag/update", updatePlayerByInfoApiHandler(a))
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

// GET /api/lol/players?q=...
func searchPlayersApiHandler(a *app.Application) func(ctx fiber.Ctx) error {
	return func(ctx fiber.Ctx) error {
		players, err := a.SearchPlayers(ctx.Context(), ctx.Query("q"), playerSearchLimit)
		if err != nil {
			log.Printf("search players: %v", err)
			return err
		}
		return ctx.JSON(players)
	}
}

// GET /api/lol/players/by-info/:server/:name/:tag
// Server không phân biệt hoa thường. Không tìm thấy player => 404.
func findPlayerByInfoApiHandler(a *app.Application) func(ctx fiber.Ctx) error {
	return func(ctx fiber.Ctx) error {
		server, name, tag, err := playerInfoParams(ctx)
		if err != nil {
			return err
		}

		player, err := a.FindPlayerByNameAndTag(ctx.Context(), server, name, tag)
		if err != nil {
			log.Printf("find player %s/%s#%s: %v", server, name, tag, err)
			return err
		}
		if player.IsNull {
			return ctx.SendStatus(fiber.StatusNotFound)
		}
		return ctx.JSON(player.Value)
	}
}

// POST /api/lol/players/by-info/:server/:name/:tag/update
// Luôn gọi Riot. Thành công => 204. Riot không tìm thấy player => 404.
func updatePlayerByInfoApiHandler(a *app.Application) func(ctx fiber.Ctx) error {
	return func(ctx fiber.Ctx) error {
		server, name, tag, err := playerInfoParams(ctx)
		if err != nil {
			return err
		}

		_, err = a.UpdatePlayerByNameAndTag(ctx.Context(), server, name, tag)
		if external.IsRiotNotFound(err) {
			return ctx.SendStatus(fiber.StatusNotFound)
		}
		if err != nil {
			log.Printf("update player %s/%s#%s: %v", server, name, tag, err)
			return err
		}
		return ctx.SendStatus(fiber.StatusNoContent)
	}
}

// ---------- private ----------

// Đọc :server/:name/:tag. Input sai => *fiber.Error 400.
func playerInfoParams(ctx fiber.Ctx) (server app.RiotServer, name, tag string, err error) {
	server = app.RiotServer(strings.ToUpper(ctx.Params("server")))
	if !server.IsValid() {
		return "", "", "", fiber.NewError(fiber.StatusBadRequest, "invalid server")
	}
	// Fiber không tự decode path param (UnescapePath = false), name có dấu / khoảng trắng sẽ bị encode.
	name, err = url.PathUnescape(ctx.Params("name"))
	if err != nil {
		return "", "", "", fiber.NewError(fiber.StatusBadRequest, "invalid name")
	}
	tag, err = url.PathUnescape(ctx.Params("tag"))
	if err != nil {
		return "", "", "", fiber.NewError(fiber.StatusBadRequest, "invalid tag")
	}
	return server, name, tag, nil
}
