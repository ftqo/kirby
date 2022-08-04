package logger

import (
	"fmt"
	"os"
	"strings"

	"github.com/disgoorg/log"
	"github.com/sirupsen/logrus"

	"github.com/ftqo/kirby/config"
)

type kirbyFormatter struct {
	logrus.TextFormatter
}

func (kf *kirbyFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	level := entry.Level.String()
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
	return []byte(fmt.Sprintf("\u001b[30;1m%s \x1b[%d;1m%s\x1b[0m%s%s\n", entry.Time.Format(kf.TimestampFormat), color, strings.ToUpper(level[:4]), "  ", entry.Message)), nil
}

func GetLogger(c config.LogConfig) log.Logger {
	log := logrus.New()

	switch strings.ToLower(c.Level) {
	case "trace":
		log.SetLevel(logrus.TraceLevel)
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "info":
		log.SetLevel(logrus.InfoLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	case "fatal":
		log.SetLevel(logrus.FatalLevel)
	case "panic":
		log.SetLevel(logrus.PanicLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}

	log.SetFormatter(&kirbyFormatter{logrus.TextFormatter{
		TimestampFormat: c.Timestamp.Format,
		FullTimestamp:   c.Timestamp.Full,
	}})
	log.SetOutput(os.Stdout)
	log.Info("logger initialized")
	return log
}
