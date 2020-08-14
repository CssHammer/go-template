package config

import (
	"fmt"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/CssHammer/go-template/storage/mongo"
	"github.com/CssHammer/go-template/storage/postgres"
)

const (
	FlagDebug                = "debug"
	FlagListenHTTP           = "http_listen"
	FlagCacheDSN             = "cache_dsn"
	FlagPostgresDSN          = "postgres_dsn"
	FlagPostgresMaxOpenConns = "postgres_max_open_conns"
	FlagMongoDSN             = "mongo_dsn"
	FlagMongoDB              = "mongo_db"
)

func ReadEnv() (*Config, error) {
	pflag.Bool(FlagDebug, false, "debug mode")
	pflag.String(FlagListenHTTP, ":80", "http listen address")
	pflag.String(FlagCacheDSN, "http://localhost:6379", "cache dsn")
	pflag.String(FlagPostgresDSN, "postgres://user:password@localhost:5432/db?sslmode=disable", "postgres dsn")
	pflag.Int(FlagPostgresMaxOpenConns, 5, "postgres max open connections")
	pflag.String(FlagMongoDSN, "mongodb://localhost:27017", "mongo dsn")
	pflag.String(FlagMongoDB, "db", "mongo db name")

	pflag.Parse()
	viper.AutomaticEnv()
	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		return nil, fmt.Errorf("bind flags: %w", err)
	}

	return &Config{
		Debug:      viper.GetBool(FlagDebug),
		HTTPListen: viper.GetString(FlagListenHTTP),
		CacheDSN:   viper.GetString(FlagCacheDSN),
		Postgres: postgres.Config{
			DSN:          viper.GetString(FlagPostgresDSN),
			MaxOpenConns: viper.GetInt(FlagPostgresMaxOpenConns),
		},
		Mongo: mongo.Config{
			DSN:    viper.GetString(FlagMongoDSN),
			DBName: viper.GetString(FlagMongoDB),
		},
	}, nil
}
