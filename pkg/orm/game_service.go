package orm

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"qilin-api/pkg/model"
	"qilin-api/pkg/model/game"
	"strings"
)

// GameService is service to interact with database and Game object.
type GameService struct {
	db *gorm.DB
}

// NewGameService initialize this service.
func NewGameService(db *Database) (*GameService, error) {
	return &GameService{db.database}, nil
}

func (p *GameService) GetTags(tag_ids []string) (tags []model.GameTag, err error) {
	tags = []model.GameTag{}
	if len(tag_ids) > 0 {
		err = p.db.Where("ID in (?)", tag_ids).Find(&tags).Error
		if err != nil {
			return nil, errors.Wrap(err, "While fetch tags")
		}
	}
	return
}

// CreateGame creates new Game object in database
func (p *GameService) CreateGame(internalName string) (item *model.Game, err error) {

	item = &model.Game{}
	errE := p.db.First(item, `"internalName" = ?`, internalName).Error
	if errE == nil {
		return nil, NewServiceError(400, "Name already in use")
	}

	item.ID = uuid.NewV4()
	item.InternalName = internalName
	item.FeaturesCtrl = ""
	item.FeaturesCommon = []string{}
	item.Platforms = game.Platforms{}
	item.Requirements = game.GameRequirements{}
	item.Languages = game.GameLangs{}
	item.FeaturesCommon = []string{}
	//item.Genre = []model.GameTag{}
	item.Tags = []string{}

	err = p.db.Create(item).Error
	if err != nil {
		return nil, errors.Wrap(err, "While create new game")
	}

	return
}

func (p *GameService) GetList(offset, limit int, technicalName, genre, price, releaseDate, sort string) (list []*model.Game, err error) {

	conds := []string{}
	vals := []interface{}{}
	if technicalName != "" {
		conds = append(conds, "technicalName like ?")
		vals = append(vals, technicalName)
	}
	if genre != "" {
		conds = append(conds, "genre->Title->ru like ?")
		vals = append(vals, genre)
	}

	err = p.db.Limit(limit).Offset(offset).Where(strings.Join(conds, " and "), vals...).Find(&list).Error
	if err != nil {
		return nil, err
	}

	return
}
