package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"go.uber.org/zap"

	"github.com/CssHammer/go-template/cache/redis"
	"github.com/CssHammer/go-template/config"
	"github.com/CssHammer/go-template/healthcheck"
	"github.com/CssHammer/go-template/http"
	"github.com/CssHammer/go-template/prometheus"
	"github.com/CssHammer/go-template/service"
	"github.com/CssHammer/go-template/storage/mongo"
	"github.com/CssHammer/go-template/storage/postgres"
)

func main() {
	// read config from os env and flags
	cfg, err := config.ReadEnv()
	if err != nil {
		fmt.Printf("failed to init config: %v", err)
		return
	}

	// init logger
	log, err := initLogger(cfg)
	if err != nil {
		fmt.Printf("failed to init logger: %v", err)
		return
	}

	log.Info("starting service...")

	// register metrics
	prometheus.Register()

	// prepare main context
	ctx, cancel := context.WithCancel(context.Background())
	setupGracefulShutdown(log, cancel)
	wg := new(sync.WaitGroup)

	// build postgres storage
	pgStorage, err := postgres.New(ctx, wg, log, cfg.Postgres)
	if err != nil {
		log.Error("postgres init", zap.Error(err))
		cancel()
		wg.Wait()
		return
	}

	// build mongo storage
	mongoStorage, err := mongo.New(ctx, wg, log, cfg.Mongo)
	if err != nil {
		log.Error("mongo init", zap.Error(err))
		cancel()
		wg.Wait()
		return
	}

	// build cache
	cache, err := redis.New(ctx, wg, log, cfg.CacheDSN)
	if err != nil {
		log.Error("cache init", zap.Error(err))
		cancel()
		wg.Wait()
		return
	}

	// build main service
	srv := service.New(pgStorage, cache)

	// build health check service
	healthSrv := healthcheck.New(
		pgStorage.HealthCheck,
		mongoStorage.HealthCheck,
		cache.HealthCheck,
	)

	// build prometheus middlewares
	middlewares := prometheus.NewMiddlewares(nil, nil)

	// build http service
	httpSrv := http.New(cfg.HTTPListen, log, srv, healthSrv, middlewares)

	// run http service
	httpSrv.Run(ctx, wg)

	// wait while services work
	wg.Wait()
	log.Info("shutdown")
}

func initLogger(config *config.Config) (*zap.Logger, error) {
	var log *zap.Logger
	var err error

	if config.Debug {
		log, err = zap.NewDevelopment()
	} else {
		log, err = zap.NewProduction()
	}

	if err != nil {
		return nil, fmt.Errorf("create logger (debug: %t): %w", config.Debug, err)
	}

	return log, nil
}

func setupGracefulShutdown(log *zap.Logger, stop func()) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signalChan
		log.Info("got termination signal")
		stop()
	}()
}
