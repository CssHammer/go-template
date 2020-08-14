package config

import (
	"github.com/CssHammer/go-template/storage/mongo"
	"github.com/CssHammer/go-template/storage/postgres"
)

type Config struct {
	Debug      bool
	HTTPListen string
	CacheDSN   string
	Postgres   postgres.Config
	Mongo      mongo.Config
}
