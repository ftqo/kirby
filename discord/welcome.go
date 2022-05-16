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
	"github.com/ftqo/kirby/database"

	"github.com/anthonynsimon/bild/transform"
	"github.com/fogleman/gg"
)

const (
	PfpSize = 256

	width  = 848
	height = 477
	margin = 15
)

type welcomeMessageInfo struct {
	mention   string
	nickname  string
	username  string
	guildName string
	avatarURL string
	members   int
}

func generateWelcomeMessage(ctx context.Context, log log.Logger, gw database.GuildWelcome, wi welcomeMessageInfo) discord.MessageCreate {
	log.Info("generating welcome message")
	var msg discord.MessageCreate

	r := strings.NewReplacer("%mention%", wi.mention, "%nickname%", wi.nickname, "%username%", wi.username, "%guild%", wi.guildName)
	gw.Text = r.Replace(gw.Text)
	gw.ImageText = r.Replace(gw.ImageText)

	msg.Content = gw.Text

	switch gw.Type {
	case "embed":
		log.Error("embedded welcome messages not implemented; sending plain")
	case "image":
		imageCtx := gg.NewContextForImage(assets.Images[gw.Image])
		req, err := http.NewRequestWithContext(ctx, "GET", wi.avatarURL, nil)
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
		if rawPfp.Bounds().Max.X != PfpSize {
			pfp = image.Image(transform.Resize(rawPfp, PfpSize, PfpSize, transform.Linear))
		} else {
			pfp = rawPfp
		}
		// draw colored rectangle over image
		imageCtx.SetColor(color.RGBA{52, 45, 50, 130})
		imageCtx.DrawRectangle(margin, margin, width-(2*margin), height-(2*margin))
		imageCtx.Fill()
		imageCtx.ClearPath()
		// draw outline circle and pfp
		imageCtx.SetColor(color.White)
		imageCtx.DrawCircle(width/2, height*(44.0/100.0), PfpSize/2+3)
		imageCtx.SetLineWidth(5)
		imageCtx.Stroke()
		imageCtx.DrawCircle(width/2, height*(44.0/100.0), PfpSize/2)
		imageCtx.Clip()
		imageCtx.DrawImage(pfp, width/2-PfpSize/2, height*44/100-PfpSize/2)
		imageCtx.ResetClip()
		// write title and subtitle
		fontLarge := assets.Fonts["coolveticaLarge"]
		fontSmall := assets.Fonts["coolveticaSmall"]
		imageCtx.SetFontFace(fontLarge)
		imageCtx.DrawStringAnchored(gw.ImageText, width/2, height*78/100, 0.5, 0.5)
		imageCtx.SetFontFace(fontSmall)
		imageCtx.DrawStringAnchored("member #"+strconv.Itoa(wi.members), width/2, height*85/100, 0.5, 0.5)
		buf := bytes.Buffer{}
		enc := png.Encoder{
			CompressionLevel: png.NoCompression,
		}
		err = enc.Encode(&buf, imageCtx.Image())
		if err != nil {
			log.Error("failed to encode image into bytes buffer")
		}
		f := &discord.File{
			Name:   "welcome_" + wi.nickname + ".jpg",
			Reader: &buf,
		}
		msg.Files = append(msg.Files, f)
	}

	return msg
}
