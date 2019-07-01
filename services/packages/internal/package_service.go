package internal

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"qilin-api/services/packages/internal/mapper"
	"qilin-api/services/packages/internal/model"
	"qilin-api/services/packages/proto"
)

type packageService struct {
	db *mongo.Database
}

func (service *packageService) Publish(ctx context.Context, req *proto.PublishRequest, res *proto.PublishResponse) error {
	zap.L().Info(fmt.Sprintf("Publish package `%s`", req.Package.Id))

	p := mapper.FromProto(req.Package)

	_, err := service.db.Collection("packages").UpdateOne(ctx, bson.M{"package_id": req.Package.Id}, p)
	if err != nil {
		zap.S().Error("Fail to update package", "id", req.Package.Id, "error", err)
		return err
	}

	return nil
}

func (service *packageService) Get(ctx context.Context, req *proto.GetPackageRequest, res *proto.GetPackageResponse) error {
	zap.L().Info("Get")

	p := &model.Package{}
	if err := service.db.Collection("packages").FindOne(ctx, bson.M{"package_id": req.Id}).Decode(p); err != nil {
		zap.S().Error("Fail to get package", "id", req.Id, "error", err)
		return err
	}

	res.Package = mapper.ToProto(p)

	return nil
}

func (packageService) GetPrice(context.Context, *proto.GetPackagePriceRequest, *proto.GetPackagePriceResponse) error {
	panic("implement me")
}

func NewService(db *mongo.Database) *packageService {
	return &packageService{db: db}
}
