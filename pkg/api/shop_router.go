package api

/// TEMPORARY FOR TESTING

import (
	"qilin-api/pkg/model"
	"qilin-api/pkg/sys"
)

type ShopRouter struct {
	service model.UserService
}

func InitShopRoutes(s *Server, mailer sys.Mailer) error {

	/*user := model.User{
		Nickname:   "User X",
		ID:         uuid.NewV4(),
		ExternalID: uuid.NewV4().String(),
		Lang:       "ru-Ru",
	}
	err := s.db.DB().Create(&user).Error
	if err != nil {
		return errors.Wrap(err, "by insert new user")
	}

	vendorService, err := orm.NewVendorService(s.db)
	if err != nil {
		return errors.Wrap(err, "create vendor service")
	}
	vendor, err := vendorService.Create(&model.Vendor{
		ManagerID: user.ID,
	})
	if err != nil {
		return errors.Wrap(err, "create vendor")
	}

	gameService, err := orm.NewGameService(s.db)
	if err != nil {
		return errors.Wrap(err, "create game service")
	}
	game, err := gameService.Create(user.ID, vendor.ID, "Game Y")
	if err != nil {
		return errors.Wrap(err, "create game")
	}

	err = s.db.DB().Create(model.Package{
		Name: "A",
		Products: []model.Product{game},
	}).Error
	if err != nil {
		return errors.Wrap(err, "create package A")
	}
	err = s.db.DB().Create(model.Package{
		Name: "B",
		Products: []model.Product{
			game,
		},
	}).Error
	if err != nil {
		return errors.Wrap(err, "create package B")
	}

	//api.Router.GET("/me", userRouter.getAppState)*/

	return nil
}

