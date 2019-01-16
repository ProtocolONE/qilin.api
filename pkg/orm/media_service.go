package orm

import (
	"qilin-api/pkg/model"
	"github.com/satori/go.uuid"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
)

// MediaService is service to interact with database and Media object.
type MediaService struct {
	db *gorm.DB
}

// NewMediaService initialize this service.
func NewMediaService(db *Database) (*MediaService, error) {
	return &MediaService{db.database}, nil
}

func (p *MediaService) Get(id uuid.UUID) (*model.Media, error) {

	result := &model.Media{}
	err := p.db.Where("ID = ?", id).First(&result).Error

	if err == gorm.ErrRecordNotFound {
		return result, echo.NewHTTPError(404, "Game not found")
	} else if err != nil {
		return result, errors.Wrap(err, "search game by id")
	}

	return result, err
}

func (p *MediaService) Update(id uuid.UUID, media *model.Media) error {

	m := &model.Media{}
	err := p.db.First(m, id).Error

	if err == gorm.ErrRecordNotFound {
		return echo.NewHTTPError(404, "Game not found")
	} else if err != nil {
		return errors.Wrap(err, "search game by id")
	}

	m.CoverImage = media.CoverImage

	err = p.db.Save(&m).Error

	if err != nil {
		return echo.NewHTTPError(422, "update game by id")
	}

	return err
}
