package discord

import (
	"bytes"
	"context"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"net/http"
	"strconv"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/log"
	"github.com/ftqo/kirby/assets"
	"github.com/ftqo/kirby/database/queries"
	"github.com/golang/freetype/truetype"

	"github.com/anthonynsimon/bild/transform"
	"github.com/fogleman/gg"
)

const (
	avatarSize = 256
	width      = 848
	height     = 477
	margin     = 15
)

type welcome = queries.InsertWelcomeParams

type welcomeReplace struct {
	mention   string
	nickname  string
	username  string
	guildName string
	avatarURL string
	members   int
}

func generateWelcomeMessage(log log.Logger, w welcome, wr welcomeReplace, a *assets.Assets) discord.MessageCreate {
	log.Trace("generating welcome message")
	var msg discord.MessageCreate

	r := strings.NewReplacer("%mention%", wr.mention, "%nickname%", wr.nickname,
		"%username%", wr.username, "%guild%", wr.guildName, "%members%", strconv.Itoa(wr.members))
	w.MessageText = r.Replace(w.MessageText)
	w.ImageTitle = r.Replace(w.ImageTitle)
	w.ImageSubtitle = r.Replace(w.ImageSubtitle)

	msg.Content = w.MessageText

	switch w.MessageType {
	case "embed":
		log.Error("embedded welcome messages not implemented; sending plain")
	case "image":
		imageCtx := gg.NewContextForImage(a.Images[w.ImageName])
		req, err := http.NewRequestWithContext(context.Background(), "GET", wr.avatarURL, nil)
		if err != nil {
			log.Error("failed to generate request for user profile pic: ", err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Error("failed to get profile picture response", err)
		}
		defer resp.Body.Close()
		rawPfp, _, err := image.Decode(resp.Body)
		if err != nil {
			log.Error("failed to decode profile picture", err)
		}

		// resize if necessary
		var pfp image.Image
		if rawPfp.Bounds().Max.X != avatarSize {
			pfp = image.Image(transform.Resize(rawPfp, avatarSize, avatarSize, transform.Linear))
		} else {
			pfp = rawPfp
		}

		// draw colored rectangle over image
		imageCtx.SetColor(color.RGBA{50, 45, 50, 130})
		imageCtx.DrawRectangle(margin, margin, width-(2*margin), height-(2*margin))
		imageCtx.Fill()

		// draw outline circle
		imageCtx.SetColor(color.White)
		imageCtx.DrawCircle(width/2, height*44/100, avatarSize/2+3)
		imageCtx.SetLineWidth(5)
		imageCtx.Stroke()

		// draw pfp in a circle
		imageCtx.DrawCircle(width/2, height*44/100, avatarSize/2)
		imageCtx.Clip()
		imageCtx.DrawImage(pfp, width/2-avatarSize/2, height*44/100-(avatarSize/2))
		imageCtx.ResetClip()

		// write title and subtitle
		font := a.Fonts["coolvetica"]
		face := truetype.NewFace(&font, &truetype.Options{Size: 40})
		imageCtx.SetFontFace(face)
		imageCtx.DrawStringAnchored(w.ImageTitle, width/2, height*78/100, 0.5, 0.5)
		face = truetype.NewFace(&font, &truetype.Options{Size: 25})
		imageCtx.SetFontFace(face)
		imageCtx.DrawStringAnchored(w.ImageSubtitle, width/2, height*85/100, 0.5, 0.5)

		// encode and add file to message
		buf := bytes.Buffer{}
		enc := png.Encoder{
			CompressionLevel: png.NoCompression,
		}
		err = enc.Encode(&buf, imageCtx.Image())
		if err != nil {
			log.Error("failed to encode image into bytes buffer")
		}
		f := &discord.File{
			Name:   "welcome_" + wr.nickname + ".jpg",
			Reader: &buf,
		}
		msg.Files = append(msg.Files, f)
	}

	return msg
}

func defaultWelcome(gid string) welcome {
	return welcome{
		GuildID:       gid,
		ChannelID:     "",
		MessageType:   "image",
		MessageText:   "hi %mention%, welcome to %guild% :)",
		ImageName:     "original",
		ImageTitle:    "%username% joined the server",
		ImageSubtitle: "member #%members%",
	}
}
