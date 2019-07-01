package cmd

import (
	"fmt"
	"github.com/ProtocolONE/rbac"
	"github.com/casbin/redis-adapter"
	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"qilin-api/pkg/api"
	"qilin-api/pkg/conf"
	"qilin-api/pkg/orm"
	"qilin-api/pkg/sys"
)

func init() {
	runServerCommand := &cobra.Command{
		Use:   "server",
		Short: "Run Qilin server",
		Run:   runServer,
	}
	command.AddCommand(runServerCommand)
}

func runServer (_ *cobra.Command, _ []string) {
	logger, _ := zap.NewProduction()
	zap.ReplaceGlobals(logger)

	config := &conf.Config{}

	if err := envconfig.Process("QILINAPI", config); err != nil {
		zap.L().Fatal("Config init failed", zap.Error(err))
	}

	logger.Debug("Config accepted")

	db, err := orm.NewDatabase(&config.Database)
	if err != nil {
		logger.Fatal("Failed to make Postgres connection", zap.Error(err))
	}

	if err := db.Init(); err != nil {
		logger.Fatal("Failed to migrate database", zap.Error(err))
	}

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

	adapter := redisadapter.NewAdapter("tcp", fmt.Sprintf("%s:%d", config.Enforcer.Host, config.Enforcer.Port))

	enf := rbac.NewEnforcer(adapter)

	serverOptions := api.ServerOptions{
		Auth1:            &config.Auth1,
		ServerConfig:     &config.Server,
		Database:         db,
		Mailer:           mailer,
		Notifier:         notifier,
		CentrifugoSecret: config.Notifier.Secret,
		Enforcer:         enf,
		EventBus:         &config.EventBus,
		Imaginary:        &config.Imaginary,
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
