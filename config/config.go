package config

import (
	_ "embed"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type DiscordConfig struct {
	TestGuild uint64 `yaml:"testGuild"`
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

type LogConfig struct {
	Level     string `yaml:"level"`
	Timestamp struct {
		Format string `yaml:"format"`
		Full   bool   `yaml:"full"`
	} `yaml:"timestamp"`
}

type Config struct {
	APIConfig     `yaml:"api"`
	DBConfig      `yaml:"db"`
	DiscordConfig `yaml:"discord"`
	LogConfig     `yaml:"log"`
}

func GetConfig() (Config, error) {
	b, err := os.ReadFile("config.yaml")
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file: %v", err)
	}

	c := Config{APIConfig{}, DBConfig{}, DiscordConfig{}, LogConfig{}}
	err = yaml.Unmarshal(b, &c)
	if err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal config into struct: %v", err)
	}
	return c, nil
}
