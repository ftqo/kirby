package discord

import (
	"context"
	"math"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/rest/route"
	"github.com/disgoorg/snowflake/v2"
	"github.com/ftqo/kirby/database/queries"
)

func (k *kirby) onGuildMemberJoin(e *events.GuildMemberJoin) {
	log := e.Client().Logger()
	q := queries.New(k.db)

	g, ok := e.Client().Caches().Guilds().Get(e.GuildID)
	if !ok {
		rg, err := e.Client().Rest().GetGuild(e.GuildID, true)
		if err != nil {
			log.Errorf("failed to get guild from api for simulation: %v", err)
		}
		g = rg.Guild
		g.MemberCount = g.ApproximateMemberCount
	}

	w, err := q.GetWelcome(context.Background(), e.GuildID.String())
	if err != nil {
		log.Warnf("failed to get guild welcome from database: %v", err)
		err = q.InsertWelcome(context.Background(), defaultWelcome(g.ID.String()))
		if err != nil {
			log.Errorf("failed to insert welcome into database: %v", err)
		}
		return
	}
	if len(w.ChannelID) == 0 {
		return
	}
	wr := welcomeReplace{
		mention:   e.Member.User.Mention(),
		nickname:  e.Member.User.Username,
		username:  e.Member.User.Tag(),
		avatarURL: e.Member.User.EffectiveAvatarURL(discord.WithSize(512), discord.WithFormat(route.PNG)),
		members:   g.MemberCount,
		guildName: g.Name,
	}
	wc, err := snowflake.Parse(w.ChannelID)
	if err != nil {
		log.Error("failed to parse channel ID: ", err)
		return
	}
	go func() {
		welcome := generateWelcomeMessage(log, welcome(w), wr, k.assets)
		_, err = e.Client().Rest().CreateMessage(wc, welcome)
		if err != nil {
			log.Error("failed to send welcome message: ", err)
		}
	}()
}

func (k *kirby) onReady(e *events.Ready) {
	log := e.Client().Logger()
	log.Info("kirby connected to discord")

	var min int64 = math.MinInt64
	err := e.Client().SetPresence(context.Background(), gateway.MessageDataPresenceUpdate{
		Since: &min,
		Activities: []discord.Activity{
			{
				Name: "the stars",
				Type: discord.ActivityTypeWatching,
			},
		},
		Status: discord.OnlineStatusOnline,
		AFK:    false,
	})
	if err != nil {
		log.Errorf("failed to set status on ready")
	}
}

func (k *kirby) onResume(e *events.Resumed) {
	e.Client().Logger().Debug("resumed")
}

func (k *kirby) onApplicationCommandInteractionCreate(e *events.ApplicationCommandInteractionCreate) {
	if c, ok := k.commands[e.Data.CommandName()]; ok {
		c.handler(e)
	}
}
