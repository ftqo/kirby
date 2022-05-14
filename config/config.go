package config

import (
	_ "embed"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

//go:embed config.yaml
var config []byte

type DiscordConfig struct {
	TestGuild int64  `yaml:"testGuild"`
	Token     string `yaml:"token"`
}

type DBConfig struct {
	Host     string `yaml:"host"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	Port     string `yaml:"port"`
}

type APIConfig struct {
	Port int `yaml:"port"`
}

type Config struct {
	APIConfig     `yaml:"api"`
	DBConfig      `yaml:"db"`
	DiscordConfig `yaml:"discord"`
}

func GetConfig(log *logrus.Logger) Config {
	log.Info("getting config")
	c := Config{}
	err := yaml.Unmarshal(config, &c)
	if err != nil {
		log.Panic("failed to unmarshal config into struct: ", err)
	}
	return c
}
