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
	// Concurrency !
	wg := &sync.WaitGroup{}
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	// Logging and config !
	log := logger.GetLogger()
	c := config.GetConfig(log)

	// Database and assets !
	db := database.OpenDB(ctx, log, c.DBConfig)
	defer db.Close(log)
	db.InitDatabase(ctx, log)
	assets.LoadAssets(log)

	// Services !
	wg.Add(1)
	go discord.Run(ctx, wg, log, db, c.DiscordConfig)

	wg.Wait()
}
