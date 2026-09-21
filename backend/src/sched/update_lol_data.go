package sched

import (
	"backend/src/app"
	"context"
	"fmt"
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
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	fmt.Println("Schedule: updateLolData")
	updateLolData(a)
	for range ticker.C {
		fmt.Println("Schedule: updateLolData")
		updateLolData(a)
	}
}

func updateLolData(a *app.Application) {
	ctx := context.Background()
	go func() {
		err := a.UpdateChampions(ctx)
		if err != nil {
			log.Println("UpdateChampions error:", err)
		} else {
			log.Println("UpdateChampions success")
		}

	}()
	go func() {
		err := a.UpdateItems(ctx)
		if err != nil {
			log.Println("UpdateItems error:", err)
		} else {
			log.Println("UpdateItems success")
		}
	}()
	go func() {
		err := a.UpdateSpells(ctx)
		if err != nil {
			log.Println("UpdateSpells error:", err)
		} else {
			log.Println("UpdateSpells success")
		}
	}()
	go func() {
		err := a.UpdateRunes(ctx)
		if err != nil {
			log.Println("UpdateRunes error:", err)
		} else {
			log.Println("UpdateRunes success")
		}
	}()
}
