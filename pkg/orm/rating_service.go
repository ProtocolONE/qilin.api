package orm

import (
	"net/http"
	"qilin-api/pkg/model"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
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
	rating := model.GameRating{}
	game := &model.Game{ID: id}
	if s.db.NewRecord(&game) {
		return &rating, NewServiceError(http.StatusNotFound, "Game not found")
	}

	err := s.db.Model(&game).Related(&rating).Error

	if err != gorm.ErrRecordNotFound {
		return &rating, errors.Wrap(err, "Search related ratings")
	}

	return &rating, nil
}

//SaveRatingsForGame is method for updating game ratings
func (s *RatingService) SaveRatingsForGame(id uuid.UUID, newRating *model.GameRating) error {
	rating := model.GameRating{}
	err := s.db.Model(&model.Game{ID: id}).Related(&rating).Error

	if err != gorm.ErrRecordNotFound {
		return errors.Wrap(err, "search game by id")
	}

	if rating.ID != 0 {
		newRating.ID = rating.ID
	}

	newRating.GameID = id

	err = s.db.Save(newRating).Error
	if err != nil {
		return errors.Wrap(err, "save rating error")
	}

	return nil
}
