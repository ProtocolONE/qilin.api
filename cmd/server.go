package cmd

import (
	"qilin-api/pkg/api"
	"qilin-api/pkg/orm"
	"qilin-api/pkg/sys"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run Qilin api server with given configuration",
	Run:   runServer,
}

func init() {
	rootCmd.AddCommand(serverCmd)
}

func runServer(cmd *cobra.Command, args []string) {
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
