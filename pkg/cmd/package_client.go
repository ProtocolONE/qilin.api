// USING IN TESTING PURPOSE ONLY

package cmd

import (
	"context"
	"github.com/micro/go-micro"
	"github.com/spf13/cobra"
	"qilin-api/services/packages/proto"
)

func init() {
	runServerCommand := &cobra.Command{
		Use:   "test-micro",
		Short: "Run micro service client",
		Run:   runClientService,
	}
	command.AddCommand(runServerCommand)
}

func runClientService (_ *cobra.Command, _ []string) {
	service := micro.NewService(micro.Name("packages.service.client"))
	service.Init()

	packageMicroService := proto.NewPackageService(proto.ServiceName, service.Client())
	_, err := packageMicroService.Publish(context.Background(), &proto.PublishRequest{Package: &proto.Package{
		Name: &proto.LocalizedString{
			EN: "Test",
		},
	}})

	if err != nil {
		panic(err)
	}
}
