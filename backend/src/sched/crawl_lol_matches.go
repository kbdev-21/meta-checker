package sched

import (
	"backend/src/app"
	"context"
	"time"
)

// Mỗi server một goroutine tự lặp: crawl xong lượt này mới nghỉ CrawlFreq rồi chạy lượt kế,
// nên hai lượt của cùng một crawler không bao giờ chồng nhau (ticker nhịp cố định thì lượt
// chưa xong đã có lượt mới cùng ghi crawler.idx / crawler.puuids).
// Trả về ngay, các crawler chạy nền.
func StartCrawlLolMatchesSched(a *app.Application) {
	for _, cr := range app.Crawlers {
		go startCrawler(a, cr)
	}
}

// ---------- private ----------

func startCrawler(a *app.Application, cr *app.Crawler) {
	for {
		_ = a.CrawlMatches(context.Background(), cr)
		time.Sleep(10 * time.Second)
	}
}
