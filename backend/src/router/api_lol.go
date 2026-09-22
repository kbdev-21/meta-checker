package router

import (
	"log"
	"net/url"
	"strconv"
	"strings"

	"backend/src/app"
	"backend/src/external"
	"backend/src/shared"

	"github.com/gofiber/fiber/v3"
)

const playerSearchLimit = 20

// count mặc định / tối đa cho list match; giới hạn để 1 request không gọi Riot quá nhiều.
const (
	matchListDefaultCount = 10
	matchListMaxCount     = 10
)

func InitLolApiRoutes(fib *fiber.App, a *app.Application) {
	fib.Get("/api/lol/players", searchPlayersApiHandler(a))
	fib.Get("/api/lol/players/by-info/:server/:name/:tag", findPlayerByInfoApiHandler(a))
	fib.Post("/api/lol/players/by-info/:server/:name/:tag/update", updatePlayerByInfoApiHandler(a))

	fib.Get("/api/lol/matches/by-player-info/:server/:name/:tag", getMatchesByPlayerInfoApiHandler(a))

	fib.Get("/api/lol/analytics", getAnalyticsApiHandler(a))
	fib.Get("/api/lol/analytics/champions/:id", getChampionAnalyticsApiHandler(a))
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

		player, err := a.FindPlayerByPlayerInfo(ctx.Context(), server, name, tag)
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

		_, err = a.UpdatePlayerByPlayerInfo(ctx.Context(), server, name, tag)
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

// GET /api/lol/matches/by-player-info/:server/:name/:tag?mode=SOLO|FLEX&start=0&count=20
// mode optional, không phân biệt hoa thường. Không tìm thấy player => 404.
func getMatchesByPlayerInfoApiHandler(a *app.Application) func(ctx fiber.Ctx) error {
	return func(ctx fiber.Ctx) error {
		server, name, tag, err := playerInfoParams(ctx)
		if err != nil {
			return err
		}
		mode, start, count, err := matchListParams(ctx)
		if err != nil {
			return err
		}

		matches, err := a.GetMatchesByPlayerInfo(ctx.Context(), server, name, tag, mode, start, count)
		if err != nil {
			log.Printf("get matches %s/%s#%s: %v", server, name, tag, err)
			return err
		}
		if matches.IsNull {
			return ctx.SendStatus(fiber.StatusNotFound)
		}
		return ctx.JSON(matches.Value)
	}
}

// GET /api/lol/analytics?server=GLOBAL&rankBucket=MASTER_PLUS
// Chưa tổng hợp cho patch hiện tại => 404.
func getAnalyticsApiHandler(a *app.Application) func(ctx fiber.Ctx) error {
	return func(ctx fiber.Ctx) error {
		server, bucket, err := analyticsParams(ctx)
		if err != nil {
			return err
		}

		meta, err := a.GetAnalytics(ctx.Context(), server, bucket)
		if err != nil {
			log.Printf("get analytics %s/%s: %v", server, bucket, err)
			return err
		}
		if meta.IsNull {
			return ctx.SendStatus(fiber.StatusNotFound)
		}
		return ctx.JSON(meta.Value)
	}
}

// GET /api/lol/analytics/champions/:id?server=GLOBAL&rankBucket=MASTER_PLUS
// Trả list ChampionStat mọi position của champion. Meta chưa tổng hợp => 404.
func getChampionAnalyticsApiHandler(a *app.Application) func(ctx fiber.Ctx) error {
	return func(ctx fiber.Ctx) error {
		server, bucket, err := analyticsParams(ctx)
		if err != nil {
			return err
		}
		id, err := strconv.Atoi(ctx.Params("id"))
		if err != nil || id <= 0 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid id")
		}

		stats, err := a.GetChampionStats(ctx.Context(), server, bucket, int32(id))
		if err != nil {
			log.Printf("get champion analytics %s/%s/%d: %v", server, bucket, id, err)
			return err
		}
		if stats.IsNull {
			return ctx.SendStatus(fiber.StatusNotFound)
		}
		return ctx.JSON(stats.Value)
	}
}

// ---------- private ----------

// Đọc ?server&rankBucket. Không truyền => mặc định GLOBAL / MASTER_PLUS. Input sai => *fiber.Error 400.
func analyticsParams(ctx fiber.Ctx) (server string, bucket app.RankBucket, err error) {
	server = strings.ToUpper(ctx.Query("server", app.MetaServerGlobal))
	if !app.IsValidMetaServer(server) {
		return "", "", fiber.NewError(fiber.StatusBadRequest, "invalid server")
	}
	bucket = app.RankBucket(strings.ToUpper(ctx.Query("rankBucket", string(app.RankBucketMasterPlus))))
	if !bucket.IsValid() {
		return "", "", fiber.NewError(fiber.StatusBadRequest, "invalid rankBucket")
	}
	return server, bucket, nil
}

// Đọc :server/:name/:tag. Input sai => *fiber.Error 400.
func playerInfoParams(ctx fiber.Ctx) (server app.Server, name, tag string, err error) {
	server = app.Server(strings.ToUpper(ctx.Params("server")))
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

// Đọc ?mode&start&count. Input sai => *fiber.Error 400.
func matchListParams(ctx fiber.Ctx) (mode shared.Nullable[app.GameMode], start, count int, err error) {
	mode = shared.Nullable[app.GameMode]{IsNull: true}
	if q := ctx.Query("mode"); q != "" {
		m := app.GameMode(strings.ToUpper(q))
		if m != app.GameModeSolo && m != app.GameModeFlex {
			return mode, 0, 0, fiber.NewError(fiber.StatusBadRequest, "invalid mode")
		}
		mode = shared.Nullable[app.GameMode]{Value: m}
	}

	start, err = strconv.Atoi(ctx.Query("start", "0"))
	if err != nil || start < 0 {
		return mode, 0, 0, fiber.NewError(fiber.StatusBadRequest, "invalid start")
	}
	count, err = strconv.Atoi(ctx.Query("count", strconv.Itoa(matchListDefaultCount)))
	if err != nil || count < 1 || count > matchListMaxCount {
		return mode, 0, 0, fiber.NewError(fiber.StatusBadRequest, "invalid count")
	}
	return mode, start, count, nil
}
