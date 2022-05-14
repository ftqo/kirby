package assets

import (
	"bytes"
	"image"
	"io/ioutil"
	"path"
	"runtime"
	"strings"

	"github.com/golang/freetype/truetype"
	"github.com/sirupsen/logrus"
	"golang.org/x/image/font"
)

var (
	Fonts  map[string]font.Face
	Images map[string]image.Image
)

func LoadAssets(log *logrus.Logger) {
	log.Info("loading assets into memory")
	Images = make(map[string]image.Image)
	Fonts = make(map[string]font.Face)
	_, b, _, _ := runtime.Caller(0)
	d := path.Join(path.Dir(b))
	imagesPath := path.Join(d, "images")
	fontsPath := path.Join(d, "fonts")
	imgs, err := ioutil.ReadDir(imagesPath)
	if err != nil {
		log.Panicf("failed to read directory %s: %v", imagesPath, err)
	}
	for _, file := range imgs {
		fp := path.Join(imagesPath, file.Name())
		bts, err := ioutil.ReadFile(fp)
		if err != nil {
			log.Panicf("failed to read file %s: %v", fp, err)
		}
		fn := file.Name()
		noPre := fn[strings.LastIndex(file.Name(), "-")+1:]
		noExt := noPre[:strings.Index(noPre, ".")]
		Images[noExt], _, err = image.Decode(bytes.NewReader(bts))
		if err != nil {
			log.Panicf("failed to decode %s", fp)
		}
		log.Debugf("loaded %s", fp)
	}
	fonts, err := ioutil.ReadDir(fontsPath)
	if err != nil {
		log.Panicf("failed to read directory %s", fonts)
	}
	for _, file := range fonts {
		fp := path.Join(fontsPath, file.Name())
		bts, err := ioutil.ReadFile(fp)
		if err != nil {
			log.Panicf("failed to read file %s", fp)
		}
		fn := file.Name()
		noPre := fn[strings.LastIndex(file.Name(), "-")+1:]
		noExt := noPre[:strings.Index(noPre, ".")]
		fnt, err := truetype.Parse(bts)
		if err != nil {
			log.Panicf("failed to parse font %s", fp)
		}
		large := truetype.NewFace(fnt, &truetype.Options{Size: 40})
		small := truetype.NewFace(fnt, &truetype.Options{Size: 25})
		Fonts[noExt+"Large"] = large
		Fonts[noExt+"Small"] = small

		log.Debugf("loaded %s", fp)
	}
}
