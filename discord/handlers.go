package discord

import (
	"context"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/disgo/rest/route"
	"github.com/disgoorg/snowflake/v2"
	"github.com/ftqo/kirby/database"
	"github.com/sirupsen/logrus"
)

func createOnGuildMemberJoin(ctx context.Context, log *logrus.Logger, db database.DB) func(*events.GuildMemberJoinEvent) {
	log.Info("creating OnGuildMemberJoin function")
	return func(e *events.GuildMemberJoinEvent) {
		g, ok := e.Client().Caches().Guilds().Get(e.GuildID)
		if !ok {
			log.Error("guild not in cache")
			return
		}
		gw, err := db.GetGuildWelcome(ctx, log, e.GuildID.String())
		if err != nil {
			log.Error("failed to get guild welcome from database, retrying: ", err)
			// RETRY. THERE IS A CHANCE THE BOT JOINED A SERVER WHILE THE BOT WAS DOWN.
			db.InsertGuild(ctx, log, e.GuildID.String())
			gw, err = db.GetGuildWelcome(ctx, log, e.GuildID.String())
			if err != nil {
				log.Error("failed to get guild welcome from database: ", err)
				return
			}
		}
		if gw.ChannelID == "" {
			return
		}
		wi := welcomeMessageInfo{
			mention:   e.Member.User.Mention(),
			nickname:  e.Member.User.Username,
			username:  e.Member.User.Tag(),
			avatarURL: e.Member.User.EffectiveAvatarURL(discord.WithSize(512), discord.WithFormat(route.PNG)),
			members:   g.MemberCount,
			guildName: g.Name,
		}
		welcomeChannel, err := snowflake.Parse(gw.ChannelID)
		if err != nil {
			log.Error("failed to parse channel ID: ", err)
			return
		}
		welcome := generateWelcomeMessage(ctx, log, gw, wi)
		_, err = e.Client().Rest().CreateMessage(welcomeChannel, welcome, rest.WithCtx(ctx))
		if err != nil {
			log.Error("failed to send welcome message: ", err)
		}
	}
}
