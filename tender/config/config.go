package config

import (
	"github.com/gookit/slog"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type (
	Config struct {
		HTTP
		PG
	}
	HTTP struct {
		ServerAddress string `env:"SERVER_ADDRESS"`
	}
	PG struct {
		PostgresConn string `env:"POSTGRES_CONN"`
	}
)

func NewConfig() *Config {
	cfg := &Config{}
	err := godotenv.Load()
	if err != nil {
		slog.Fatalf("can't load env %s", err.Error())
	}
	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		slog.Fatalf("error reading env %s", err.Error())
	}
	return cfg
}
