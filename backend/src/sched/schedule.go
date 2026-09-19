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
}
