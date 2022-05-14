package database

type GuildWelcome struct {
	GuildID   string
	ChannelID string
	Type      string
	Text      string
	Image     string
	ImageText string
}

type Session struct {
	SessionID string
	Sequence  string
}

func NewDefaultGuildWelcome() GuildWelcome {
	return GuildWelcome{
		ChannelID: "",
		Type:      "image",
		Text:      "hi %mention%, welcome to %guild% :)",
		Image:     "original",
		ImageText: "%username% joined the server",
	}
}
