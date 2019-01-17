package orm

import (
	"qilin-api/pkg/model"
	"github.com/satori/go.uuid"
	"github.com/jinzhu/gorm"
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
	err := p.db.First(&result, model.Media{ID: id}).Error

	if err == gorm.ErrRecordNotFound {
		return result, NewServiceError(404, "Game not found")
	} else if err != nil {
		return result, errors.Wrap(err, "search game by id")
	}

	return result, err
}

func (p *MediaService) Update(id uuid.UUID, media *model.Media) error {

	m := model.Media{}
	err := p.db.First(&m, model.Media{ID: id}).Error

	if err == gorm.ErrRecordNotFound {
		return NewServiceError(404, "Game not found")
	} else if err != nil {
		return errors.Wrap(err, "search game by id")
	}

	if media.UpdatedAt.Before(m.UpdatedAt) {
		// return NewServiceError(409, "Game has new changes")
	}

	media.UpdatedAt = m.UpdatedAt
	media.CreatedAt = m.CreatedAt
	media.ID = m.ID

	err = p.db.Save(&media).Error

	if err != nil {
		return errors.Wrap(err, "search game by id")
	}

	return err
}
