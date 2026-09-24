package app

import (
	"backend/src/external"
	"backend/src/shared"
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

const crawlFreq = 10 * time.Second
const crawlBatch = 10
const crawlMatchesPerPlayer = 10

type Crawler struct {
	Server Server
	puuids []string
	idx    int
}

var Crawlers = []*Crawler{
	{Server: ServerVN},
	{Server: ServerKR},
	{Server: ServerEUW},
	{Server: ServerEUNE},
	{Server: ServerNA},
	{Server: ServerBR},
}

func (a *Application) CrawlMatches(ctx context.Context, crawler *Crawler) error {
	// Crawler chạy nền nên xếp hàng sau request của user ở rate limiter.
	ctx = external.WithPriority(ctx, external.PriorityLow)

	if len(crawler.puuids) == 0 {
		err := a.resetCrawlPuuids(ctx, crawler)
		if err != nil {
			return err
		}
	}

	// Lượt cuối của một vòng ngắn hơn crawlBatch khi len(puuids) không chia hết.
	success := 0
	end := min(crawler.idx+crawlBatch, len(crawler.puuids))
	for i := crawler.idx; i < end; i++ {
		// 1 player lỗi (account bị ban, đổi region, Riot hỏng...) chỉ bỏ qua player đó chứ
		// không chặn cả crawler: idx vẫn tiến nên lượt sau crawl tiếp người kế tiếp.
		err := a.crawlPlayer(ctx, crawler.Server, crawler.puuids[i])
		if err == nil {
			success++
		}
		time.Sleep(crawlFreq)
	}

	log.Printf("Crawler server %s: %d/%d players\n", crawler.Server, success, crawlBatch)
	crawler.idx = end

	// Hết một vòng thì lấy lại ladder. Reset lỗi thì puuids cũ được giữ nguyên và idx vẫn bằng
	// len(puuids), nên lượt sau không crawl ai mà thử reset lại.
	if crawler.idx >= len(crawler.puuids) {
		err := a.resetCrawlPuuids(ctx, crawler)
		if err != nil {
			return err
		}
	}

	return nil
}

// ---------- private ----------

// Cập nhật player rồi lưu các match gần nhất của 1 puuid.
func (a *Application) crawlPlayer(ctx context.Context, server Server, puuid string) error {
	_, err := a.UpdatePlayerByPuuid(ctx, server, puuid)
	if err != nil {
		return err
	}
	_, err = a.GetMatchesByPuuid(ctx, server, puuid, shared.Nullable[GameMode]{Value: GameModeSolo}, 0, crawlMatchesPerPlayer)
	return err
}

func (a *Application) resetCrawlPuuids(ctx context.Context, crawler *Crawler) error {
	var (
		chal   []string
		gm     []string
		master []string
		wg     sync.WaitGroup
	)
	wg.Add(3)

	go func() {
		defer wg.Done()
		entries, err := a.riot.FetchChallengerLeague(ctx, serverRiotRoutings[crawler.Server].platform, "RANKED_SOLO_5x5")
		if err != nil {
			return
		}
		for _, e := range entries.Entries {
			chal = append(chal, e.Puuid)
		}
	}()

	go func() {
		defer wg.Done()
		entries, err := a.riot.FetchGrandmasterLeague(ctx, serverRiotRoutings[crawler.Server].platform, "RANKED_SOLO_5x5")
		if err != nil {
			return
		}
		for _, e := range entries.Entries {
			gm = append(gm, e.Puuid)
		}
	}()

	go func() {
		defer wg.Done()
		entries, err := a.riot.FetchMasterLeague(ctx, serverRiotRoutings[crawler.Server].platform, "RANKED_SOLO_5x5")
		if err != nil {
			return
		}
		for _, e := range entries.Entries {
			master = append(master, e.Puuid)
		}
	}()

	wg.Wait()

	if len(chal) == 0 || len(gm) == 0 || len(master) == 0 {
		log.Printf("Crawler server %s: reset puuids failed\n", crawler.Server)
		return fmt.Errorf("fetch puuids failed for crawler %s", string(crawler.Server))
	}

	crawler.puuids = []string{}
	crawler.idx = 0
	crawler.puuids = append(crawler.puuids, chal...)
	crawler.puuids = append(crawler.puuids, gm...)
	crawler.puuids = append(crawler.puuids, master...)

	log.Printf("Crawler server %s: reset %d puuids\n", crawler.Server, len(crawler.puuids))

	return nil
}
