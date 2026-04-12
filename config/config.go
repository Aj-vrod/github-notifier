package config

import (
	"Aj-vrod/github-notifier/internal/poller"
	"Aj-vrod/github-notifier/pkg/github"
	"log"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	GithubCfg github.GithubConfig
	PollerCfg poller.Config
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err == nil {
		log.Println("Loading .env file")
	}
	cfg := &Config{}
	if err := envconfig.Process("", cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
