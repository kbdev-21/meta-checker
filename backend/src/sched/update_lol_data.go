package sched

import (
	"backend/src/app"
	"context"
	"log"
	"time"
)

// Sync champions / items / spells / runes từ ddragon, chạy ngay 1 lần rồi lặp mỗi giờ.
// Trả về ngay, sched chạy nền.
func StartUpdateLolDataSched(a *app.Application) {
	go startUpdateLolData(a)
}

// ---------- private ----------

func startUpdateLolData(a *app.Application) {
	ticker := time.NewTicker(2 * time.Hour)
	defer ticker.Stop()

	updateLolData(a)
	for range ticker.C {
		updateLolData(a)
	}
}

func updateLolData(a *app.Application) {
	ctx := context.Background()
	go func() {
		err := a.UpdateChampions(ctx)
		if err != nil {
			log.Println("Update champions data error:", err)
		} else {
			log.Println("Update champions data success")
		}

	}()
	go func() {
		err := a.UpdateItems(ctx)
		if err != nil {
			log.Println("Update items data error:", err)
		} else {
			log.Println("Update items data success")
		}
	}()
	go func() {
		err := a.UpdateSpells(ctx)
		if err != nil {
			log.Println("Update spells data error:", err)
		} else {
			log.Println("Update spells data success")
		}
	}()
	go func() {
		err := a.UpdateRunes(ctx)
		if err != nil {
			log.Println("Update runes data error:", err)
		} else {
			log.Println("Update runes data success")
		}
	}()
}
