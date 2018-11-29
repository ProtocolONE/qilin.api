package cmd

import (
	"qilin-api/pkg/api"
	"qilin-api/pkg/orm"

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
		logger.Fatal("Failed to start mongo session: " + err.Error())
	}

	db.Init()

	defer func() {
		logger.Fatal(db.Close())
	}()

	serverConfig := api.ServerConfig{
		Log:          logger,
		Jwt:          &config.Jwt,
		ServerConfig: &config.Server,
		Database:     db,
	}

	server, err := api.NewServer(&serverConfig)
	if err != nil {
		logger.Fatal("Failed to create server: " + err.Error())
	}

	logger.Infof("Starting up server")
	err = server.Start()
	if err != nil {
		logger.Fatal("Failed to start server: " + err.Error())
	}
}
