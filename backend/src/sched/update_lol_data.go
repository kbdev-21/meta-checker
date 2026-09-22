package sched

import (
	"backend/src/app"
	"context"
	"log"
	"sync"
	"time"
)

// Sync champions / items / spells / runes từ ddragon, chạy ngay 1 lần rồi lặp mỗi giờ.
// Trả về ngay, sched chạy nền.
func StartUpdateLolDataSched(a *app.Application) {
	go startUpdateLolData(a)
}

// ---------- private ----------

func startUpdateLolData(a *app.Application) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	updateLolData(a)
	for range ticker.C {
		updateLolData(a)
	}
}

// Sync 4 loại data ddragon song song, chờ xong hết (WaitGroup) rồi mới tổng hợp analytics
// (build items cần lol_items mới nhất, patch lấy từ ddragon).
func updateLolData(a *app.Application) {
	ctx := context.Background()

	var wg sync.WaitGroup
	wg.Add(4)
	go func() {
		defer wg.Done()
		err := a.UpdateChampions(ctx)
		if err != nil {
			log.Println("Update champions data error:", err)
		} else {
			log.Println("Update champions data success")
		}
	}()
	go func() {
		defer wg.Done()
		err := a.UpdateItems(ctx)
		if err != nil {
			log.Println("Update items data error:", err)
		} else {
			log.Println("Update items data success")
		}
	}()
	go func() {
		defer wg.Done()
		err := a.UpdateSpells(ctx)
		if err != nil {
			log.Println("Update spells data error:", err)
		} else {
			log.Println("Update spells data success")
		}
	}()
	go func() {
		defer wg.Done()
		err := a.UpdateRunes(ctx)
		if err != nil {
			log.Println("Update runes data error:", err)
		} else {
			log.Println("Update runes data success")
		}
	}()
	wg.Wait()

	err := a.UpdateAnalytics(ctx)
	if err != nil {
		log.Println("Update analytics error:", err)
	} else {
		log.Println("Update analytics success")
	}
}
