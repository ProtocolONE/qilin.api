package orm

import (
	"github.com/jinzhu/gorm"
	"qilin-api/pkg/model"
)

// GameService is service to interact with database and Game object.
type GameService struct {
	db *gorm.DB
}

// NewGameService initialize this service.
func NewGameService(db *Database) (*GameService, error) {
	return &GameService{db.database}, nil
}

// CreateGame creates new Game object in database
func (p *GameService) CreateGame(u *model.Game) error {
	return p.db.Create(u).Error
}

func (p *GameService) UpdateGame(u *model.Game) error {
	return p.db.Update(u).Error
}

// FindByID return Game object by given id
func (p *GameService) FindByID(id uint) (game model.Game, err error) {
	err = p.db.First(&game, model.Game{ID: id}).Error
	return
}

func (p *GameService) GetAll() ([]*model.Game, error) {
	var games []*model.Game
	err := p.db.Find(&games).Error

	return games, err
}

func (p *GameService) FindByName(name string) ([]*model.Game, error) {
	var games []*model.Game
	err := p.db.Where("name LIKE ?", name).Find(&games).Error

	return games, err
}
