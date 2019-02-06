package main

import (
	"github.com/kelseyhightower/envconfig"
	"log"
	"qilin-api/pkg/api"
	"qilin-api/pkg/conf"
	"qilin-api/pkg/orm"
	"qilin-api/pkg/sys"
	"runtime/debug"
)

func main() {
	config := &conf.Config{}

	if err := envconfig.Process("QAPI", config); err != nil {
		log.Fatalf("Config init failed with error: %s\n", err)
	}

	logger, err := conf.ConfigureLogging(&config.Log)
	if err != nil {
		log.Fatal("Failed to configure logging: " + err.Error())
	}

	logger.Debugf("Config accepted")

	db, err := orm.NewDatabase(&config.Database)
	if err != nil {
		logger.Fatal("Failed to make Postgres connection: " + err.Error())
	}

	db.Init()

	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
			logger.Error(string(debug.Stack()))
		}
		logger.Fatal(db.Close())
	}()

	mailer := sys.NewMailer(config.Mailer)

	serverOptions := api.ServerOptions{
		Log:          logger,
		Jwt:          &config.Jwt,
		ServerConfig: &config.Server,
		Database:     db,
		Mailer:       mailer,
	}

	server, err := api.NewServer(&serverOptions)
	if err != nil {
		logger.Fatal("Failed to create server: " + err.Error())
	}

	logger.Infof("Starting up server")
	err = server.Start()
	if err != nil {
		logger.Fatal("Failed to start server: " + err.Error())
	}
}
