package orm

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/model"
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
		return result, NewServiceError(404, "Game not found")
	} else if err != nil {
		return result, errors.Wrap(err, "search game by id")
	}

	return result, err
}

func (p *MediaService) Update(id uuid.UUID, media *model.Media) error {

	m := model.Media{}
	err := p.db.Where("ID = ?", id).First(&m).Error

	if err == gorm.ErrRecordNotFound {
		return NewServiceError(http.StatusNotFound, "Game not found")
	} else if err != nil {
		return errors.Wrap(err, "search game by id")
	}

	if media.UpdatedAt.Before(m.UpdatedAt) {
		return NewServiceError(http.StatusConflict, "Game has new changes")
	}

	media.CreatedAt = m.CreatedAt
	media.ID = m.ID

	err = p.db.Save(&media).Error

	if err != nil {
		return errors.Wrap(err, "save media for game")
	}

	return err
}
