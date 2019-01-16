package orm

import (
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
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
func (p *GameService) CreateGame(internalName string) (game *model.Game, err error) {

	game = &model.Game{}
	errE := p.db.First(game, "internalName = ?", internalName).Error
	if errE == nil {
		return nil, echo.NewHTTPError(400, "Name already in use")
	}

	game.ID = uuid.NewV4()
	game.InternalName = internalName
	err = p.db.Create(&game).Error
	if err != nil {
		return nil, errors.Wrap(err, "While create new game")
	}

	return
}

func (p *GameService) UpdateGame(u *model.Game) error {
	return p.db.Update(u).Error
}

// FindByID return Game object by given id
func (p *GameService) FindByID(id uuid.UUID) (game model.Game, err error) {
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
