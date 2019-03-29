package mock

import (
	"github.com/satori/go.uuid"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
)

type gameService struct {

}

func (gameService) GetOwnerForGame(gameId uuid.UUID) (string, error) {
	panic("implement me")
}

func (gameService) CreateTags([]model.GameTag) error {
	panic("implement me")
}

func (gameService) CreateGenres([]model.GameGenre) error {
	return nil
}

func (gameService) GetTags([]string) ([]model.GameTag, error) {
	return []model.GameTag{}, nil
}

func (gameService) GetGenres([]string) ([]model.GameGenre, error) {
	return []model.GameGenre{}, nil
}

func (gameService) GetRatingDescriptors(system string) ([]model.Descriptor, error) {
	return []model.Descriptor{}, nil
}

func (gameService) FindTags(userId string, title string, limit, offset int) ([]model.GameTag, error) {
	return []model.GameTag{}, nil
}

func (gameService) FindGenres(userId string, title string, limit, offset int) ([]model.GameGenre, error) {
	return []model.GameGenre{}, nil
}

func (gameService) Create(userId string, vendorId uuid.UUID, internalName string) (*model.Game, error) {
	return &model.Game{}, nil
}

func (gameService) Delete(userId string, gameId uuid.UUID) error {
	return nil
}

func (gameService) GetList(userId string, vendorId uuid.UUID, offset, limit int, internalName, genre, releaseDate, sort string, price float64) ([]*model.ShortGameInfo, error) {
	return []*model.ShortGameInfo{}, nil
}

func (gameService) GetInfo(gameId uuid.UUID) (*model.Game, error) {
	return &model.Game{}, nil
}

func (gameService) UpdateInfo(game *model.Game) error {
	return nil
}

func (gameService) GetDescr(gameId uuid.UUID) (*model.GameDescr, error) {
	return &model.GameDescr{}, nil
}

func (gameService) UpdateDescr(descr *model.GameDescr) error {
	return nil
}

func (gameService) GetProduct(gameId uuid.UUID) (model.Product, error) {
	return nil, nil
}

func NewGameService(_ *orm.Database) (model.GameService, error) {
	return &gameService{}, nil
}
