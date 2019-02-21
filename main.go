package main

import (
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
	"log"
	"qilin-api/pkg/api"
	"qilin-api/pkg/conf"
	"qilin-api/pkg/orm"
	"qilin-api/pkg/sys"
)

func main() {
	logger, _ := zap.NewProduction()
	zap.ReplaceGlobals(logger)

	config := &conf.Config{}

	if err := envconfig.Process("QILINAPI", config); err != nil {
		log.Fatalf("Config init failed with error: %s\n", err)
	}

	logger.Debug("Config accepted")

	db, err := orm.NewDatabase(&config.Database)
	if err != nil {
		logger.Fatal("Failed to make Postgres connection", zap.Error(err))
	}

	db.Init()

	defer func() {
		if err := recover(); err != nil {
			logger.Sugar().Error("Failed on main recover", "error", err)
		}

		err := db.Close()
		if err != nil {
			logger.Error("Closing database error", zap.Error(err))
		}
	}()

	mailer := sys.NewMailer(config.Mailer)

	notifier, err := sys.NewNotifier(config.Notifier.ApiKey, config.Notifier.Host)
	if err != nil {
		logger.Fatal("Failed to create notifier", zap.Error(err))
	}

	serverOptions := api.ServerOptions{
		Jwt:          &config.Jwt,
		ServerConfig: &config.Server,
		Database:     db,
		Mailer:       mailer,
		Notifier:     notifier,
	}

	server, err := api.NewServer(&serverOptions)
	if err != nil {
		logger.Fatal("Failed to create server", zap.Error(err))
	}

	logger.Info("Starting up server")
	err = server.Start()
	if err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
