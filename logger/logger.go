package logger

import (
	"github.com/sirupsen/logrus"
)

func GetLogger() *logrus.Logger {
	log := logrus.New()
	log.SetLevel(logrus.InfoLevel)
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "01-02--15-04-05",
	})
	log.Info("logger initialized")
	return log
}
