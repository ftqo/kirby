package discord

import (
	"context"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/rest/route"
	"github.com/disgoorg/snowflake/v2"
	"github.com/ftqo/kirby/database/queries"
)

type command struct {
	def     discord.ApplicationCommandCreate
	handler func(*events.ApplicationCommandInteractionCreate)
}

func (k *kirby) getCommands() map[string]command {
	// experimental structure for organizing commands...
	return map[string]command{
		"ping": {
			def: discord.SlashCommandCreate{
				CommandName: "ping",
				Description: "a simple command to test if the bot is online",
			},
			handler: func(e *events.ApplicationCommandInteractionCreate) {
				err := e.CreateMessage(discord.NewMessageCreateBuilder().SetContent("pong!").SetEphemeral(true).Build())
				if err != nil {
					e.Client().Logger().Errorf("failed to create pong message response: %v", err)
				}
			},
		},
		"welcome": {
			def: discord.SlashCommandCreate{
				CommandName:              "welcome",
				Description:              "several commands for setting up welcome messages",
				DefaultMemberPermissions: discord.PermissionManageServer,
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionSubCommand{
						CommandName: "set",
						Description: "set welcome message options. placeholders: %guild%, %mention%, %username%, and %nickname%",
						Options: []discord.ApplicationCommandOption{
							discord.ApplicationCommandOptionChannel{
								OptionName:  "channel",
								Description: "the channel to send welcome messages in",
								Required:    false,
							},
							discord.ApplicationCommandOptionString{
								OptionName:  "message",
								Description: "the contents of the message",
								Required:    false,
							},
							discord.ApplicationCommandOptionString{
								OptionName:  "image_title",
								Description: "the message in the top row of the image",
								Required:    false,
							},
							discord.ApplicationCommandOptionString{
								OptionName:  "image_subtitle",
								Description: "the message in the bottom row of the image",
								Required:    false,
							},
							discord.ApplicationCommandOptionString{
								OptionName:  "type",
								Description: "the type of message (plain, embed, or image) for the welcome message",
								Required:    false,
								Choices: []discord.ApplicationCommandOptionChoiceString{
									{
										Name:  "image",
										Value: "image",
									}, {
										Name:  "plain",
										Value: "plain",
									},
								},
							},
							discord.ApplicationCommandOptionString{
								OptionName:  "image",
								Description: "the background image for the welcome message",
								Choices: []discord.ApplicationCommandOptionChoiceString{
									{
										Name:  "original",
										Value: "original",
									},
									{
										Name:  "beach",
										Value: "beach",
									},
									{
										Name:  "sleepy",
										Value: "sleepy",
									},
									{
										Name:  "friends",
										Value: "friends",
									},
									{
										Name:  "melon",
										Value: "melon",
									},
									{
										Name:  "sky",
										Value: "sky",
									},
								},
							},
						},
					},
					discord.ApplicationCommandOptionSubCommand{
						CommandName: "simulate",
						Description: "simulate a welcome message",
					},
					discord.ApplicationCommandOptionSubCommand{
						CommandName: "reset",
						Description: "reset all welcome settings to default",
					},
				},
			},
			handler: func(e *events.ApplicationCommandInteractionCreate) {
				log := e.Client().Logger()
				data := e.SlashCommandInteractionData()
				q := queries.New(k.db)

				switch *data.SubCommandName {
				case "set":
					err := e.CreateMessage(discord.NewMessageCreateBuilder().SetContent("setting welcome config!").SetEphemeral(true).Build())
					if err != nil {
						e.Client().Logger().Errorf("failed to set send message responding to welcome set")
					}

					tx, err := k.db.Begin()
					if err != nil {
						e.Client().Logger().Errorf("failed to begin transaction for welcome set: %v", err)
					}
					q = q.WithTx(tx)
					if channel, ok := data.OptChannel("channel"); ok {
						err = q.SetWelcomeChannel(context.Background(), queries.SetWelcomeChannelParams{GuildID: e.GuildID().String(), ChannelID: channel.ID.String()})
						if err != nil {
							e.Client().Logger().Errorf("failed to set channel for welcome set: %v", err)
						}
					}

					if message, ok := data.OptString("message"); ok {
						err = q.SetWelcomeMessageText(context.Background(), queries.SetWelcomeMessageTextParams{GuildID: e.GuildID().String(), MessageText: message})
						if err != nil {
							e.Client().Logger().Errorf("failed to set channel for welcome set: %v", err)
						}
					}
					if title, ok := data.OptString("image_title"); ok {
						err = q.SetWelcomeImageTitle(context.Background(), queries.SetWelcomeImageTitleParams{GuildID: e.GuildID().String(), ImageTitle: title})
						if err != nil {
							e.Client().Logger().Errorf("failed to set channel for welcome set: %v", err)
						}
					}
					if subtitle, ok := data.OptString("image_subtitle"); ok {
						err = q.SetWelcomeImageSubtitle(context.Background(), queries.SetWelcomeImageSubtitleParams{GuildID: e.GuildID().String(), ImageSubtitle: subtitle})
						if err != nil {
							e.Client().Logger().Errorf("failed to set channel for welcome set: %v", err)
						}
					}
					if image, ok := data.OptString("image"); ok {
						err = q.SetWelcomeImageName(context.Background(), queries.SetWelcomeImageNameParams{GuildID: e.GuildID().String(), ImageName: image})
						if err != nil {
							e.Client().Logger().Errorf("failed to set channel for welcome set: %v", err)
						}
					}
					if typ, ok := data.OptString("type"); ok {
						err = q.SetWelcomeMessageType(context.Background(), queries.SetWelcomeMessageTypeParams{GuildID: e.GuildID().String(), MessageType: typ})
						if err != nil {
							e.Client().Logger().Errorf("failed to set channel for welcome set: %v", err)
						}
					}
					err = tx.Commit()
					if err != nil {
						e.Client().Logger().Errorf("failed to commit transaction for welcome set: %v", err)
					}

				case "simulate":
					w, err := q.GetWelcome(context.Background(), e.GuildID().String())
					if err != nil {
						e.CreateMessage(discord.NewMessageCreateBuilder().SetContent("welcome channel not set, use `/welcome set` and pick a channel!").SetEphemeral(true).Build())
						q.InsertWelcome(context.Background(), defaultWelcome(e.GuildID().String()))
						return
					}
					if len(w.ChannelID) == 0 {
						e.CreateMessage(discord.NewMessageCreateBuilder().SetContent("welcome channel not set, use `/welcome set` and pick a channel!").SetEphemeral(true).Build())
						return
					}

					err = e.CreateMessage(discord.NewMessageCreateBuilder().SetContent("simulating welcome!").SetEphemeral(true).Build())
					if err != nil {
						e.Client().Logger().Errorf("failed to set send message responding to welcome simulate")
					}

					g, ok := e.Client().Caches().Guilds().Get(*e.GuildID())
					if !ok {
						rg, err := e.Client().Rest().GetGuild(*e.GuildID(), true)
						if err != nil {
							log.Errorf("failed to get guild from api for simulation: %v", err)
						}
						g = rg.Guild
						g.MemberCount = g.ApproximateMemberCount
					}

					wr := welcomeReplace{
						mention:   e.Member().Mention(),
						nickname:  e.Member().User.Username,
						username:  e.Member().User.Tag(),
						avatarURL: e.Member().User.EffectiveAvatarURL(discord.WithSize(512), discord.WithFormat(route.PNG)),
						members:   g.MemberCount,
						guildName: g.Name,
					}

					message := generateWelcomeMessage(e.Client().Logger(), welcome(w), wr, k.assets)
					channel, err := snowflake.Parse(w.ChannelID)
					if err != nil {
						log.Errorf("failed to parse channel snowflake from channel id: %v", err)
					}
					_, err = e.Client().Rest().CreateMessage(channel, message)
					if err != nil {
						log.Error("failed to send simulated welcome message: %v", err)
					}
				case "reset":
					err := e.CreateMessage(discord.NewMessageCreateBuilder().SetContent("not implemented!").SetEphemeral(true).Build())
					if err != nil {
						e.Client().Logger().Errorf("failed to set send message responding to welcome reset")
					}
				}
			},
		},
	}
}
