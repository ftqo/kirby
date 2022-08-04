package assets

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	"path"
	"strings"

	"github.com/disgoorg/log"
	"github.com/golang/freetype/truetype"
)

//go:embed fonts
var fontsFS embed.FS

//go:embed images
var imagesFS embed.FS

type Assets struct {
	Images map[string]image.Image
	Fonts  map[string]truetype.Font
}

func GetAssets(log log.Logger) (*Assets, error) {
	a := &Assets{
		Images: make(map[string]image.Image),
		Fonts:  make(map[string]truetype.Font),
	}
	err := a.LoadImages(log)
	if err != nil {
		return nil, fmt.Errorf("failed to load images: %v", err)
	}
	err = a.LoadFonts(log)
	if err != nil {
		return nil, fmt.Errorf("failed to load fonts: %v", err)
	}
	return a, nil
}

func (a *Assets) LoadImages(log log.Logger) error {
	log.Info("loading images into memory")
	files, err := imagesFS.ReadDir("images")
	if err != nil {
		return fmt.Errorf("failed to read embedded images directory: %v", err)
	}
	for _, file := range files {
		fname := file.Name()
		raw, err := imagesFS.ReadFile(path.Join("images", fname))
		if err != nil {
			return fmt.Errorf("failed to read file %s: %v", fname, err)
		}
		name := fname[strings.LastIndex(fname, "-")+1 : strings.Index(fname, ".")]
		var format string
		a.Images[name], format, err = image.Decode(bytes.NewReader(raw))
		if err != nil {
			return fmt.Errorf("failed to decode image %s: %v", fname, err)
		}
		if format != "jpeg" {
			return fmt.Errorf("failed to decode %s as jpeg (formatted as %s)", fname, format)
		}
		log.Debugf("loaded %s", fname)
	}
	return nil
}

func (a *Assets) LoadFonts(log log.Logger) error {
	log.Info("loading images into memory")
	files, err := fontsFS.ReadDir("fonts")
	if err != nil {
		return fmt.Errorf("failed to read embedded fonts directory: %v", err)
	}
	for _, file := range files {
		fname := file.Name()
		raw, err := fontsFS.ReadFile(path.Join("fonts", fname))
		if err != nil {
			log.Panicf("failed to read file %s: %v", fname, err)
		}
		name := fname[strings.LastIndex(file.Name(), "-")+1 : strings.Index(fname, ".")]
		font, err := truetype.Parse(raw)
		if err != nil {
			return fmt.Errorf("failed to parse font %s: %v", fname, err)
		}
		a.Fonts[name] = *font

		log.Debugf("loaded %s", fname)
	}
	return nil
}
