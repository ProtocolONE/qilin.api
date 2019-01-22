package orm

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"qilin-api/pkg/model"
	"qilin-api/pkg/model/game"
	"strings"
	"time"
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

func (p *GameService) GetGenres(ganre_ids []string) (genres []model.GameGenre, err error) {

	// In this case return all possible genres
	if ganre_ids == nil {
		err = p.db.Find(&genres).Order("id").Error
		if err != nil {
			return nil, errors.Wrap(err, "While fetch genres")
		}
		return
	}

	genres = []model.GameGenre{}
	if len(ganre_ids) > 0 {
		err = p.db.Where("ID in (?)", ganre_ids).Find(&genres).Error
		if err != nil {
			return nil, errors.Wrap(err, "While fetch genres")
		}
	}
	return
}

// CreateGame creates new Game object in database
func (p *GameService) Create(vendorId *uuid.UUID, internalName string) (item *model.Game, err error) {

	item = &model.Game{}
	errE := p.db.First(item, `internal_name ilike ?`, internalName).Error
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
	item.Genre = []string{}
	item.Tags = []string{}
	item.VendorId = *vendorId

	err = p.db.Create(item).Error
	if err != nil {
		return nil, errors.Wrap(err, "While create new game")
	}

	return
}

func (p *GameService) GetList(vendorId *uuid.UUID,
	offset, limit int, internalName, genre, releaseDate, sort string, price float64) (list []*model.Game, err error) {

	vendor := model.User{}
	err = p.db.Select("lang, currency").Where("id = ?", vendorId).First(&vendor).Error
	if err != nil {
		return nil, errors.Wrap(err, "while fetch vendor")
	}

	conds := []string{}
	vals := []interface{}{}

	if internalName != "" {
		conds = append(conds, `internal_name ilike ?`)
		vals = append(vals, internalName)
	}

	if genre != "" {
		genres := []model.GameGenre{}
		err = p.db.Where("title ->> ? ilike ?", vendor.Lang, genre).Limit(1).Find(&genres).Error
		if err != nil {
			return nil, errors.Wrap(err, "while fetch genres")
		}
		if len(genres) == 0 {
			return // 200: No any genre found
		}
		conds = append(conds, "? = ANY(genre)")
		vals = append(vals, genres[0].ID)
	}

	if releaseDate != "" {
		rdate, err := time.Parse("2006-01-02", releaseDate)
		if err != nil {
			return nil, NewServiceError(400, "Invalid date")
		}
		conds = append(conds, `date(release_date) = ?`)
		vals = append(vals, rdate)
	}

	if price > 0 {
		conds = append(conds, `game_prices.value = ?`)
		vals = append(vals, price)
	}

	conds = append(conds, `vendor_id = ?`)
	vals = append(vals, vendorId)

	var orderBy interface{}
	orderBy = "created_at ASC"
	if sort != "" {
		switch sort {
		case "-genre": orderBy = "created_at DESC"
		case "+genre": orderBy = "genre ASC"
		case "-releaseDate": orderBy = "release_date DESC"
		case "+releaseDate": orderBy = "release_date ASC"
		case "-price": orderBy = "game_prices.value DESC"
		case "+price": orderBy = "game_prices.value ASC"
		case "-name": orderBy = "internal_name DESC"
		case "+name": orderBy = "internal_name ASC"
		}
	}

	err = p.db.
		Model(model.Game{}).
		//Joins("LEFT JOIN game_prices on game_prices.game_id = game.id and game_prices.currency = ?", vendor.Currency).
		Where(strings.Join(conds, " and "), vals...).
		Order(orderBy).
		Limit(limit).
		Offset(offset).
		Find(&list).Error
	if err != nil {
		return nil, errors.Wrap(err, "Fetch games list")
	}

	return
}

func (p *GameService) GetInfo(vendorId *uuid.UUID, gameId *uuid.UUID) (game *model.Game, err error) {

	game = &model.Game{}
	err = p.db.First(game, `id = ? and vendor_id = ?`, gameId, vendorId).Error
	if err == gorm.ErrRecordNotFound {
		return nil, NewServiceError(404, "Game not found")
	} else if err != nil {
		return nil, errors.Wrap(err, "Fetch game info")
	}

	return game, nil
}

func (p *GameService) Delete(vendorId *uuid.UUID, gameId *uuid.UUID) (err error) {

	game := &model.Game{}
	err = p.db.First(game, `id = ? and vendor_id = ?`, gameId, vendorId).Error
	if err == gorm.ErrRecordNotFound {
		return  NewServiceError(404, "Game not found")
	} else if err != nil {
		return errors.Wrap(err, "Fetch game info")
	}

	err = p.db.Delete(game).Error
	if err != nil {
		return errors.Wrap(err, "Delete game")
	}

	return nil
}


func (p *GameService) Update(vendorId *uuid.UUID, game *model.Game) (err error) {

	err = p.db.Model(game).Where(`vendor_id = ?`, vendorId).Update(game).Error
	if err == gorm.ErrRecordNotFound {
		return  NewServiceError(404, "Game not found")
	} else if err != nil {
		return errors.Wrap(err, "Update game")
	}

	return nil
}
