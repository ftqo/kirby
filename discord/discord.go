package discord

import (
	"context"
	"strconv"
	"sync"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/log"
	"github.com/gorilla/websocket"

	"github.com/ftqo/kirby/config"
	"github.com/ftqo/kirby/database"
)

func Run(ctx context.Context, wg *sync.WaitGroup, log log.Logger, db database.DB, config config.DiscordConfig) {
	log.Info("running discord service")
	defer wg.Done()

	var sequence int
	var sessionID string
	s, err := db.GetKV(ctx, log, "session")
	if err != nil {
		log.Warn("no sessionID or sequence detected")
	} else {
		log.Info("sessionID and sequence detected, attempting to resume")
		sequence, err = strconv.Atoi(s["sequence"])
		if err != nil {
			log.Error("failed to convert sequence from type string to int: ", err)
		}
		sessionID = s["sessionID"]
	}

	client, err := disgo.New(config.Token,
		bot.WithGatewayConfigOpts(
			gateway.WithGatewayIntents(
				discord.GatewayIntentGuildMembers,
				discord.GatewayIntentGuilds,
			),
			gateway.WithSequence(sequence),
			gateway.WithSessionID(sessionID),
		),
		bot.WithCacheConfigOpts(cache.WithCacheFlags(cache.FlagsDefault)),
		bot.WithEventListeners(&events.ListenerAdapter{
			OnGuildMemberJoin: createOnGuildMemberJoin(ctx, db),
		}),
		bot.WithLogger(log),
	)
	if err != nil {
		log.Panic("failed to build disgo: ", err)
	}

	if err = client.ConnectGateway(ctx); err != nil {
		log.Panic("failed to connect to gateway: ", err)
	}

	<-ctx.Done()

	log.Info("gracefully shutting down discord service")
	db.UpsertKV(context.Background(), log, "session", map[string]string{
		"sessionID": *client.Gateway().SessionID(),
		"sequence":  strconv.Itoa(*client.Gateway().LastSequenceReceived()),
	})

	client.Gateway().CloseWithCode(context.Background(), websocket.CloseServiceRestart, "Restarting")
}
