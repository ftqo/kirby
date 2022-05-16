package logger

import (
	"fmt"
	"os"
	"strings"

	"github.com/disgoorg/log"
	"github.com/sirupsen/logrus"
)

type kirbyFormatter struct {
	logrus.TextFormatter
}

func (kf *kirbyFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var color int
	switch entry.Level {
	case logrus.TraceLevel:
		color = 30 // grey
	case logrus.DebugLevel:
		color = 37 // white
	case logrus.InfoLevel:
		color = 36 // blue
	case logrus.WarnLevel:
		color = 33 // yellow
	case logrus.ErrorLevel:
		color = 31 // red
	case logrus.FatalLevel:
		color = 31
	case logrus.PanicLevel:
		color = 31
	default:
		color = 32
	}
	return []byte(fmt.Sprintf("%s \x1B[1m\x1b[%d;1m%s\x1b[0m%s%s\n", entry.Time.Format(kf.TimestampFormat), color, strings.ToUpper(entry.Level.String()), strings.Repeat(" ", 7-len(strings.ToUpper(entry.Level.String()))), entry.Message)), nil
}

func GetLogger() log.Logger {
	log := logrus.New()
	log.SetLevel(logrus.TraceLevel)
	log.SetFormatter(&kirbyFormatter{logrus.TextFormatter{
		TimestampFormat: "2006/01/02 15:04:05",
		FullTimestamp:   true,
	}})
	log.SetOutput(os.Stdout)
	log.Info("logger initialized")
	return log
}
