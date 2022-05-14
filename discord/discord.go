package discord

import (
	"context"
	"sync"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/sirupsen/logrus"

	"github.com/ftqo/kirby/config"
	"github.com/ftqo/kirby/database"
)

func Run(ctx context.Context, wg *sync.WaitGroup, log *logrus.Logger, db database.DB, config config.DiscordConfig) {
	log.Info("running discord service")
	defer wg.Done()

	client, err := disgo.New(config.Token,
		bot.WithGatewayConfigOpts(
			gateway.WithGatewayIntents(
				discord.GatewayIntentGuildMembers,
				discord.GatewayIntentGuilds,
			),
			// gateway.WithAutoReconnect(true),
			// gateway.WithSequence(),
			// gateway.WithSessionID(),
		),
		bot.WithCacheConfigOpts(cache.WithCacheFlags(cache.FlagsDefault)),
		bot.WithEventListeners(&events.ListenerAdapter{
			OnGuildMemberJoin: createOnGuildMemberJoin(ctx, log, db),
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

	client.Close(ctx)
}
