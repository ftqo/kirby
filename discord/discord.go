package discord

import (
	"context"
	"database/sql"
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

	"github.com/ftqo/kirby/assets"
	"github.com/ftqo/kirby/config"
	"github.com/ftqo/kirby/database/queries"
)

type kirby struct {
	db     *sql.DB
	assets *assets.Assets

	commands map[string]command
}

func Run(ctx context.Context, wg *sync.WaitGroup, log log.Logger, config config.DiscordConfig, db *sql.DB, assets *assets.Assets) {
	log.Info("running discord service")
	defer wg.Done()

	k := kirby{db: db, assets: assets}
	q := queries.New(db)

	// get and parse old session and sequence
	var sequence int
	var sessionID string
	dbSequence, err := q.GetV(ctx, "sequence")
	if err != nil {
		log.Warnf("failed to get sequence from database: %v", err)
	}
	dbSessionID, err := q.GetV(ctx, "session")
	if err != nil {
		log.Warnf("failed to get session from database: %v", err)
	}
	if len(dbSequence) != 0 && len(dbSessionID) != 0 {
		log.Info("sessionID and sequence detected, attempting to resume")
		sessionID = dbSessionID
		sequence, err = strconv.Atoi(dbSequence)
		if err != nil {
			log.Error("failed to convert sequence from type string to int: ", err)
			sessionID = ""
		}
	} else {
		log.Warn("no sessionID and/or sequence detected")
	}

	// create bot client
	client, err := disgo.New(config.Token,
		bot.WithGatewayConfigOpts(
			gateway.WithIntents(
				gateway.IntentGuildMembers,
				gateway.IntentGuilds,
			),
			gateway.WithSequence(sequence),
			gateway.WithSessionID(sessionID),
		),
		bot.WithCacheConfigOpts(cache.WithCacheFlags(cache.FlagGuilds), cache.WithGuildCachePolicy(cache.DefaultConfig().GuildCachePolicy)),
		bot.WithEventListeners(&events.ListenerAdapter{
			OnReady:                         k.onReady,
			OnGuildMemberJoin:               k.onGuildMemberJoin,
			OnApplicationCommandInteraction: k.onApplicationCommandInteractionCreate,
			OnResumed:                       k.onResume,
		}),
		bot.WithLogger(log),
	)
	if err != nil {
		log.Panicf("failed to create disgo client: %v", err)
	}

	k.commands = k.getCommands()
	commands := []discord.ApplicationCommandCreate{}
	for _, c := range k.commands {
		commands = append(commands, c.def)
	}
	_, err = client.Rest().SetGlobalCommands(client.ApplicationID(), commands)
	if err != nil {
		log.Panicf("failed to set application commands: %v", err)
	}

	err = client.OpenGateway(ctx)
	if err != nil {
		log.Panicf("failed to connect to gateway: %v", err)
	}

	<-ctx.Done()

	log.Info("gracefully shutting down discord service")
	err = q.UpsertKV(context.Background(), queries.UpsertKVParams{
		K: "session", V: *client.Gateway().SessionID(),
	})
	if err != nil {
		log.Errorf("failed to insert session into database: %v", err)
	}

	err = q.UpsertKV(context.Background(), queries.UpsertKVParams{
		K: "sequence", V: strconv.Itoa(*client.Gateway().LastSequenceReceived()),
	})
	if err != nil {
		log.Errorf("failed to insert sequence into database: %v", err)
	}

	client.Gateway().CloseWithCode(context.Background(), websocket.CloseServiceRestart, "Restarting")
}
