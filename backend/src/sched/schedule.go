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
	upsertLolData(a)
	for range ticker.C {
		fmt.Println("Schedule: EveryHour")
		upsertLolData(a)
	}
}

func upsertLolData(a *app.Application) {
	ctx := context.Background()
	go func() {
		err := a.UpsertChampions(ctx)
		if err != nil {
			log.Println("UpsertChampions error:", err)
		} else {
			log.Println("UpsertChampions success")
		}
		
	}()
	go func() {
		err := a.UpsertItems(ctx)
		if err != nil {
			log.Println("UpsertItems error:", err)
		} else {
			log.Println("UpsertChampions success")
		}
	}()
}
