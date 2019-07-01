package packages

import (
	"context"
	"github.com/micro/go-micro"
	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"qilin-api/services/packages/internal"
	"qilin-api/services/packages/proto"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

func Init() *cobra.Command {
	return &cobra.Command{
		Use:   "package-service",
		Short: "Run micro service server for packages",
		Run:   runPackageService,
	}
}

func runPackageService(_ *cobra.Command, _ []string) {
	logger, _ := zap.NewProduction()
	zap.ReplaceGlobals(logger)

	config, err := internal.LoadConfig()
	if err != nil {
		zap.L().Fatal("Init config failed", zap.Error(err))
	}

	opts := options.Client().ApplyURI(config.Db.Uri)
	client, err := mongo.NewClient(opts)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	db := client.Database(config.Db.Name)

	var service micro.Service
	options := []micro.Option{
		micro.Name(proto.ServiceName),
		micro.Version("latest"),
	}

	logger.Info("Init service")

	service = micro.NewService(options...)
	service.Init()

	packageService := internal.NewService(db)

	if err := proto.RegisterPackageServiceHandler(service.Server(), packageService); err != nil {
		logger.Fatal("Registration service in micro is failed", zap.Error(err))
	}

	if err := service.Run(); err != nil {
		logger.Fatal("Can`t run service", zap.Error(err))
	}
}
