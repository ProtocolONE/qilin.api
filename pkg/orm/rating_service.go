package orm

import (
	"net/http"
	"qilin-api/pkg/model"
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
	game := model.Game{ID: id}
	count := 0

	if err := s.db.Model(&game).Where("ID = ?", id).Limit(1).Count(&count).Error; err != nil {
		return nil, NewServiceError(http.StatusInternalServerError, errors.Wrapf(err, "Search game with id %s", id))
	}

	if count == 0 {
		return nil, NewServiceError(http.StatusNotFound, "Game not found")
	}

	rating := model.GameRating{}
	err := s.db.Model(&game).Related(&rating).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, NewServiceError(http.StatusInternalServerError, errors.Wrapf(err, "Search related ratings with game id %s", id))
	}

	return &rating, nil
}

//SaveRatingsForGame is method for updating game ratings
func (s *RatingService) SaveRatingsForGame(id uuid.UUID, newRating *model.GameRating) error {
	rating := model.GameRating{}
	game := model.Game{ID: id}
	if s.db.NewRecord(&game) {
		return NewServiceError(http.StatusNotFound, "Game not found")
	}

	err := s.db.Model(&model.Game{ID: id}).Related(&rating).Error

	if err == gorm.ErrRecordNotFound {
		rating.CreatedAt = time.Now()
	} else if err != nil {
		return errors.Wrap(err, "search game by id")
	}

	newRating.ID = rating.ID
	newRating.GameID = id
	newRating.UpdatedAt = time.Now()
	newRating.CreatedAt = rating.CreatedAt

	err = s.db.Save(newRating).Error
	if err != nil {
		return errors.Wrap(err, "save rating error")
	}

	return nil
}
