package orm

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/model"
	bto "qilin-api/pkg/model/game"
	utils2 "qilin-api/pkg/orm/utils"
	"qilin-api/pkg/utils"
	"strings"
	"time"
)

// gameService is service to interact with database and Game object.
type gameService struct {
	db *gorm.DB
	packageFactory packageFactory
}

// NewGameService initialize this service.
func NewGameService(db *Database) (model.GameService, error) {
	return &gameService{
		db: db.database,
		packageFactory: packageFactory{db.database},
	}, nil
}

func (p *gameService) verifyVendor(vendorId uuid.UUID) error {
	if exist, err := utils2.CheckExists(p.db, &model.Vendor{}, vendorId); !(exist && err == nil) {
		if err != nil {
			return errors.Wrap(err, "Verify vendor")
		}
		return NewServiceError(404, "Vendor not found")
	}
	return nil
}

func (p *gameService) verifyUserAndVendor(userId string, vendorId uuid.UUID) error {
	if err := p.verifyVendor(vendorId); err != nil {
		return err
	}

	if exist, err := utils2.CheckExists(p.db, &model.User{}, userId); !(exist && err == nil) {
		if err != nil {
			return errors.Wrap(err, "Verify user")
		}
		return NewServiceError(404, "User not found")
	}

	return nil
}

func (p *gameService) GetTags(ids []string) (tags []model.GameTag, err error) {
	stmt := p.db
	if ids != nil && len(ids) > 0 {
		stmt = stmt.Where("ID in (?)", ids)
	}
	err = stmt.Order("id").Find(&tags).Error
	if err != nil {
		return nil, errors.Wrap(err, "While fetch tags")
	}
	return
}

func (p *gameService) GetGenres(ids []string) (genres []model.GameGenre, err error) {
	stmt := p.db
	if ids != nil && len(ids) > 0 {
		stmt = stmt.Where("ID in (?)", ids)
	}
	err = stmt.Order("id").Find(&genres).Error
	if err != nil {
		return nil, errors.Wrap(err, "While fetch genres")
	}
	return
}

func (p *gameService) GetRatingDescriptors(system string) (items []model.Descriptor, err error) {
	query := p.db.Order("title ->> 'en'")

	if system != "" {
		query = query.Where("system = ?", system)
	}

	err = query.Find(&items).Error
	if err != nil {
		return nil, errors.Wrap(err, "Fetch rating descriptors")
	}
	return
}

func (p *gameService) FindTags(userId string, title string, limit, offset int) (tags []model.GameTag, err error) {
	stmt := p.db
	if title != "" {
		user := model.User{}
		err = p.db.Select("id, lang").Where("id = ?", userId).First(&user).Error
		if err != nil {
			return nil, errors.Wrap(err, "while fetch user")
		}
		stmt = stmt.Where("title ->> ? ilike ?", user.Lang, title)
	}
	if limit > 0 {
		stmt = stmt.Limit(limit).Offset(offset)
	}
	err = stmt.Order("id").Find(&tags).Error
	if err != nil {
		return nil, errors.Wrap(err, "While fetch tags")
	}
	return
}
func (p *gameService) FindGenres(userId string, title string, limit, offset int) (genres []model.GameGenre, err error) {
	stmt := p.db
	if title != "" {
		user := model.User{}
		err = p.db.Select("lang").Where("id = ?", userId).First(&user).Error
		if err != nil {
			return nil, errors.Wrap(err, "while fetch user")
		}
		stmt = stmt.Where("title ->> ? ilike ?", user.Lang, title)
	}
	if limit > 0 {
		stmt = stmt.Limit(limit).Offset(offset)
	}
	err = stmt.Order("id").Find(&genres).Error
	if err != nil {
		return nil, errors.Wrap(err, "While fetch genres")
	}
	return
}

// Creates new Game object in database
func (p *gameService) Create(userId string, vendorId uuid.UUID, internalName string) (item *model.Game, err error) {
	if err := p.verifyUserAndVendor(userId, vendorId); err != nil {
		return nil, err
	}

	internalName = strings.Trim(internalName, " \r\n\t")
	internalName = strings.Replace(internalName, " ", "_", -1)
	if len(internalName) < 2 {
		return nil, NewServiceError(400, "Incorrect internalName")
	}
	item = &model.Game{}
	errE := p.db.First(item, `internal_name ilike ?`, internalName).Error
	if errE == nil {
		return nil, NewServiceError(400, "Name already in use")
	}

	item.ID = uuid.NewV4()
	item.InternalName = internalName
	item.FeaturesCtrl = ""
	item.FeaturesCommon = []string{}
	item.Platforms = bto.Platforms{}
	item.Requirements = bto.GameRequirements{}
	item.Languages = bto.GameLangs{}
	item.FeaturesCommon = []string{}
	item.GenreMain = 0
	item.GenreAddition = []int64{}
	item.Tags = []int64{}
	item.VendorID = vendorId
	item.CreatorID = userId
	item.Product.EntryID = item.ID
	item.DefPackageID = uuid.NewV4()

	err = p.db.Create(item).Error
	if err != nil {
		return nil, errors.Wrap(err, "While create new game")
	}

	err = p.db.Create(&model.GameDescr{
		Game:    item,
		Reviews: []bto.GameReview{},
	}).Error
	if err != nil {
		return nil, errors.Wrap(err, "Create descriptions for game")
	}

	err = p.packageFactory.Create(item.DefPackageID, vendorId, userId, item.InternalName, []uuid.UUID{item.ID})
	if err != nil {
		return nil, err
	}

	return
}

func (p *gameService) GetList(userId string, vendorId uuid.UUID,
	offset, limit int, internalName, genre, releaseDate, sort string, price float64) (list []*model.ShortGameInfo, err error) {

	if err := p.verifyUserAndVendor(userId, vendorId); err != nil {
		return nil, err
	}

	user := model.User{}
	err = p.db.Select("lang, currency").Where("id = ?", userId).First(&user).Error
	if err != nil {
		return nil, errors.Wrap(err, "while fetch user")
	}

	conds := []string{}
	vals := []interface{}{}

	if internalName != "" {
		conds = append(conds, `internal_name ilike ?`)
		vals = append(vals, "%"+internalName+"%")
	}

	if genre != "" {
		genres := []model.GameGenre{}
		/// title[user.Lang] === genre or title.en === genre
		err = p.db.Where("(title ->> ? ilike ? or title ->> 'en' ilike ?)", user.Lang, genre, genre).
			Limit(1).Find(&genres).Error
		if err != nil {
			return nil, errors.Wrap(err, "while fetch genres")
		}
		if len(genres) == 0 {
			return // 200: No any genre found
		}
		conds = append(conds, "(genre_main = ? or ? = ANY(genre_addition))")
		vals = append(vals, genres[0].ID, genres[0].ID)
	}

	if releaseDate != "" {
		rdate, err := time.Parse(time.RFC3339, releaseDate)
		if err != nil {
			return nil, NewServiceError(400, "Invalid date")
		}
		conds = append(conds, `date(release_date) = ?`)
		vals = append(vals, rdate)
	}

	var orderBy interface{}
	orderBy = "created_at ASC"
	if sort != "" {
		switch sort {
		case "-genre":
			orderBy = "game_genres.title ->> 'en' DESC, created_at DESC"
		case "+genre":
			orderBy = "game_genres.title ->> 'en' ASC, created_at ASC"
		case "-releaseDate":
			orderBy = "release_date DESC"
		case "+releaseDate":
			orderBy = "release_date ASC"
		case "-internalName":
			orderBy = "internal_name DESC"
		case "+internalName":
			orderBy = "internal_name ASC"
		}
	}

	err = p.db.
		Model(model.Game{}).
		Select("games.*").
		Joins("LEFT JOIN game_genres on game_genres.id = games.genre_main").
		Where(`vendor_id = ?`, vendorId).
		Where(strings.Join(conds, " or "), vals...).
		Order(orderBy).
		Limit(limit).
		Offset(offset).
		Find(&list).Error
	if err != nil {
		return nil, errors.Wrap(err, "Fetch games list")
	}

	return
}

func (p *gameService) GetInfo(gameId uuid.UUID) (game *model.Game, err error) {

	game = &model.Game{}
	err = p.db.First(game, `id = ?`, gameId).Error
	if err == gorm.ErrRecordNotFound {
		return nil, NewServiceError(404, "Game not found")
	} else if err != nil {
		return nil, errors.Wrap(err, "Fetch game info")
	}

	return game, nil
}

func (p *gameService) GetProduct(gameId uuid.UUID) (model.Product, error) {

	game := model.ProductGameImpl{}
	err := p.db.
		Where("id = ?", gameId).
		First(&game).Error
	if err == gorm.ErrRecordNotFound {
		return nil, NewServiceError(404, "Game not found")
	} else if err != nil {
		return nil, errors.Wrap(err, "Fetch game info")
	}

	return &game, nil
}

func (p *gameService) Delete(userId string, gameId uuid.UUID) (err error) {
	game, err := p.GetInfo(gameId)
	if err != nil {
		return err
	}

	err = p.db.Delete(game).Error
	if err != nil {
		return errors.Wrap(err, "Delete game")
	}

	return nil
}

func (p *gameService) UpdateInfo(game *model.Game) (err error) {
	gameSrc, err := p.GetInfo(game.ID)
	if err != nil {
		return err
	}
	game.CreatorID = gameSrc.CreatorID
	game.VendorID = gameSrc.VendorID
	game.CreatedAt = gameSrc.CreatedAt
	game.UpdatedAt = time.Now()
	game.InternalName = gameSrc.InternalName
	game.DefPackageID = gameSrc.DefPackageID

	if game.GenreAddition == nil {
		game.GenreAddition = []int64{}
	}
	tempGenres := game.GenreAddition
	if game.GenreMain > 0 {
		tempGenres = append(tempGenres, int64(game.GenreMain))
	}
	if len(tempGenres) > 0 {
		foundGenres := 0
		err = p.db.Model(&model.GameGenre{}).Where("id in (" + utils.JoinInt(tempGenres, ",") + ")").Count(&foundGenres).Error
		if err != nil {
			return errors.Wrap(err, "Fetch genres")
		}
		if foundGenres != len(tempGenres) {
			return NewServiceError(http.StatusUnprocessableEntity, "Invalid genre")
		}
	}
	if game.Tags == nil {
		game.Tags = []int64{}
	}
	if len(game.Tags) > 0 {
		foundTags := 0
		err = p.db.Model(&model.GameTag{}).Where("id in (" + utils.JoinInt(game.Tags, ",") + ")").Count(&foundTags).Error
		if err != nil {
			return errors.Wrap(err, "Fetch genres")
		}
		if foundTags != len(game.Tags) {
			return NewServiceError(http.StatusUnprocessableEntity, "Invalid tag")
		}
	}

	err = p.db.Save(game).Error
	if err != nil && strings.Index(err.Error(), "duplicate key value") > -1 {
		return NewServiceError(http.StatusConflict, "Invalid internal_name")
	} else if err != nil {
		return errors.Wrap(err, "Update game")
	}

	return nil
}

func (p *gameService) GetDescr(gameId uuid.UUID) (descr *model.GameDescr, err error) {
	game, err := p.GetInfo(gameId)
	if err != nil {
		return nil, err
	}
	descr = &model.GameDescr{
		Reviews: []bto.GameReview{},
	}
	err = p.db.Model(game).Related(descr).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, NewServiceError(http.StatusNotFound, "Fetch game descr")
		}
		return nil, errors.Wrap(err, "Fetch game descr")
	}
	return descr, nil
}

func (p *gameService) UpdateDescr(descr *model.GameDescr) (err error) {
	game, err := p.GetInfo(descr.GameID)
	if err != nil {
		return err
	}
	update := *descr
	if update.ID == 0 {
		found := model.GameDescr{}
		err = p.db.Model(game).Related(&found).Error
		if err == gorm.ErrRecordNotFound {
			update.CreatedAt = time.Now()
		} else if err != nil {
			return errors.Wrap(err, "Get game descr")
		} else {
			update.ID = found.ID
			update.CreatedAt = found.CreatedAt
		}
	}
	update.UpdatedAt = time.Now()
	err = p.db.Save(&update).Error
	if err != nil {
		return errors.Wrap(err, "Update game descr")
	}
	return
}

func (p *gameService) CreateTags(tags []model.GameTag) (err error) {
	for _, t := range tags {
		err = p.db.Create(&t).Error
		if err != nil {
			return errors.Wrap(err, "Create game tag")
		}
	}
	return
}

func (p *gameService) CreateGenres(genres []model.GameGenre) (err error) {
	for _, g := range genres {
		err = p.db.Create(&g).Error
		if err != nil {
			return errors.Wrap(err, "Create game tag")
		}
	}
	return
}