package sched

import (
	"backend/src/app"
	"context"
	"fmt"
	"log"
	"time"
)

func StartEveryHour(a *app.Application) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	fmt.Println("Schedule: EveryHour")
	updateLolData(a)
	for range ticker.C {
		fmt.Println("Schedule: EveryHour")
		updateLolData(a)
	}
}

func updateLolData(a *app.Application) {
	ctx := context.Background()
	go func() {
		err := a.UpdateLolChampions(ctx)
		if err != nil {
			log.Println("UpdateLolChampions error:", err)
		} else {
			log.Println("UpdateLolChampions success")
		}
		
	}()
	go func() {
		err := a.UpdateLolItems(ctx)
		if err != nil {
			log.Println("UpdateLolItems error:", err)
		} else {
			log.Println("UpdateLolItems success")
		}
	}()
}
