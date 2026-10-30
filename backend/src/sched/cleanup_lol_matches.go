package sched

import (
	"backend/src/app"
	"context"
	"log"
	"time"
)

// Chạy mỗi 24 giờ, lần đầu sau 24 giờ kể từ khi khởi động backend.
func StartCleanupLolMatchesSched(a *app.Application) {
	go startCleanupLolMatches(a)
}

// ---------- private ----------

func startCleanupLolMatches(a *app.Application) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		started := time.Now()
		patch, deleted, err := a.CleanupMatches(context.Background())
		if err != nil {
			log.Printf("Cleanup LoL matches error: patch=%q duration=%s: %v", patch, time.Since(started), err)
			continue
		}
		log.Printf("Cleanup LoL matches success: patch=%s deleted=%d duration=%s", patch, deleted, time.Since(started))
	}
}
