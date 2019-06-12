package orm

import (
	"fmt"
	"net/http"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm/utils"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
)

// RatingService is service to interact with database and Rating object.
type RatingService struct {
	db *gorm.DB
}

// NewRatingService initialize this service.
func NewRatingService(db *Database) (*RatingService, error) {
	return &RatingService{db.database}, nil
}

// GetRatingsForGame is method for getting ratings for game
func (s *RatingService) GetRatingsForGame(id uuid.UUID) (*model.GameRating, error) {
	if err := checkGameExist(s.db, id); err != nil {
		return nil, err
	}

	rating := model.GameRating{}
	err := s.db.Model(&model.Game{ID: id}).Related(&rating).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, NewServiceError(http.StatusInternalServerError, errors.Wrapf(err, "Search related ratings with game id %s", id))
	}

	return &rating, nil
}

//SaveRatingsForGame is method for updating game ratings
func (s *RatingService) SaveRatingsForGame(id uuid.UUID, newRating *model.GameRating) error {
	if err := checkGameExist(s.db, id); err != nil {
		return err
	}

	rating := model.GameRating{}

	err := s.db.Model(&model.Game{}).Where("id = ?", id).Related(&rating).Error

	if err == gorm.ErrRecordNotFound {
		rating.CreatedAt = time.Now()
	} else if err != nil {
		return errors.Wrap(err, "search game by id")
	}

	if err := checkDescriptorsForGameRating(s.db, newRating); err != nil {
		return err
	}

	newRating.ID = rating.ID
	newRating.GameID = id
	newRating.UpdatedAt = time.Now()
	newRating.CreatedAt = rating.CreatedAt

	err = s.db.Save(newRating).Error
	if err != nil {
		return NewServiceError(http.StatusInternalServerError, errors.Wrap(err, "save rating error"))
	}

	return nil
}

func checkDescriptorsForGameRating(db *gorm.DB, rating *model.GameRating) error {
	if err := checkDescriptorsForRating(db, rating.BBFC, "BBFC"); err != nil {
		return err
	}
	if err := checkDescriptorsForRating(db, rating.CERO, "CERO"); err != nil {
		return err
	}
	if err := checkDescriptorsForRating(db, rating.ESRB, "ESRB"); err != nil {
		return err
	}
	if err := checkDescriptorsForRating(db, rating.USK, "USK"); err != nil {
		return err
	}
	if err := checkDescriptorsForRating(db, rating.PEGI, "PEGI"); err != nil {
		return err
	}

	return nil
}

func checkDescriptorsForRating(db *gorm.DB, rating model.JSONB, name string) error {
	descriptors := rating[model.DescriptorsField]
	if descriptors == nil {
		return nil
	}

	if check, err := checkDescriptors(db, descriptors.([]uint), name); !check || err != nil {
		if err != nil {
			return NewServiceError(http.StatusInternalServerError, errors.Wrap(err, "Check descriptors error"))
		}
		return NewServiceError(http.StatusUnprocessableEntity, fmt.Sprintf("Descriptors for `%s` failed. Some of them does not exists in database", name))
	}
	return nil
}

func checkDescriptors(db *gorm.DB, descriptors []uint, name string) (bool, error) {
	if descriptors != nil && len(descriptors) > 0 {
		count := 0
		err := db.Model(model.Descriptor{}).Where("ID in (?) AND system = ?", descriptors, name).Count(&count).Error
		return count == len(descriptors), err
	}
	return true, nil
}

func checkGameExist(db *gorm.DB, id uuid.UUID) error {
	if exist, err := utils.CheckExists(db, &model.Game{}, id); !(exist && err == nil) {
		if err != nil {
			return NewServiceError(http.StatusInternalServerError, errors.Wrap(err, "Object exist checking failed"))
		}
		return NewServiceError(http.StatusNotFound, "Game not found")
	}
	return nil
}
