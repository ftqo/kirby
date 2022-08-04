package main

import (
	"context"
	"os/signal"
	"sync"
	"syscall"

	"github.com/ftqo/kirby/assets"
	"github.com/ftqo/kirby/config"
	"github.com/ftqo/kirby/database"
	"github.com/ftqo/kirby/discord"
	"github.com/ftqo/kirby/logger"
)

func main() {
	wg := &sync.WaitGroup{}
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	c, err := config.GetConfig()
	if err != nil {
		panic("failed to get config at startup: " + err.Error())
	}
	log := logger.GetLogger(c.LogConfig)

	db, err := database.Open(ctx, log, c.DBConfig)
	if err != nil {
		log.Fatalf("failed to open database at startup: %v", err)
	}
	defer db.Close()

	a, err := assets.GetAssets(log)
	if err != nil {
		log.Panicf("failed to get assets at startup: %v", err)
	}

	wg.Add(1)
	go discord.Run(ctx, wg, log, c.DiscordConfig, db, a)

	wg.Wait()
}
