package orm

import (
	"net/http"
	"qilin-api/pkg/model"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"

	uuid "github.com/satori/go.uuid"
)

// DiscountService is service to interact with database and Discount object.
type DiscountService struct {
	db *gorm.DB
}

// NewDiscountService initialize this service.
func NewDiscountService(db *Database) (*DiscountService, error) {
	return &DiscountService{db.database}, nil
}

//GetDiscountsForGame is method for getting all discounts for game
func (s *DiscountService) GetDiscountsForGame(id uuid.UUID) ([]model.Discount, error) {

	var result []model.Discount
	game := model.Game{ID: id}
	err := s.db.Model(&game).Related(&result).Error

	if err == gorm.ErrRecordNotFound {
		return nil, NewServiceError(http.StatusNotFound, "Game not found")
	} else if err != nil {
		return nil, errors.Wrap(err, "search game by id")
	}

	return result, nil
}

//AddDiscountForGame is method for creating new discount for game
func (s *DiscountService) AddDiscountForGame(id uuid.UUID, discount *model.Discount) (uuid.UUID, error) {
	game := model.Game{ID: id}
	count := 0
	err := s.db.Model(&game).Count(&count).Error

	if err == gorm.ErrRecordNotFound || count == 0 {
		return uuid.Nil, NewServiceError(http.StatusNotFound, "Game not found")
	} else if err != nil {
		return uuid.Nil, errors.Wrap(err, "search game by id")
	}

	if discount.Rate <= 0 {
		return uuid.Nil, NewServiceError(http.StatusUnprocessableEntity, "Rate should be more than 0")
	}

	discount.GameID = id
	discount.ID = uuid.NewV4()

	err = s.db.Create(discount).Error
	if err != nil {
		return uuid.Nil, errors.Wrap(err, "Insert discount")
	}

	return discount.ID, nil
}

//UpdateDiscountForGame method for update existing discount
func (s *DiscountService) UpdateDiscountForGame(discount *model.Discount) error {
	discountInDb := model.Discount{}
	discountInDb.ID = discount.ID

	err := s.db.Model(&discountInDb).First(&discount).Error

	if err == gorm.ErrRecordNotFound {
		return NewServiceError(http.StatusNotFound, "Discount not found")
	} else if err != nil {
		return errors.Wrap(err, "search game by id")
	}

	if discount.Rate <= 0 {
		return NewServiceError(http.StatusUnprocessableEntity, "Rate should be more than 0")
	}

	discount.GameID = discountInDb.GameID

	err = s.db.Model(&discountInDb).Update(discount).Error
	if err != nil {
		return errors.Wrap(err, "Update discount")
	}

	return nil
}

//RemoveDiscountForGame is method for removing discount for game
func (s *DiscountService) RemoveDiscountForGame(id uuid.UUID) error {
	discountInDb := model.Discount{}
	discountInDb.ID = id

	err := s.db.Model(&discountInDb).Delete(&discountInDb).Error
	if err == gorm.ErrRecordNotFound {
		return NewServiceError(http.StatusNotFound, "Discount not found")
	} else if err != nil {
		return errors.Wrap(err, "search game by id")
	}

	return nil
}
