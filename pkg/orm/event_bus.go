package orm

import (
	"fmt"
	"github.com/ProtocolONE/qilin-common/pkg/proto"
	"github.com/ProtocolONE/rabbitmq/pkg"
	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
	"github.com/satori/go.uuid"
	"go.uber.org/zap"
	"qilin-api/pkg/mapper"
	"qilin-api/pkg/model"
	"qilin-api/pkg/model/game"
	"qilin-api/pkg/model/utils"
	"strconv"
	"strings"
	"time"
)

type eventBus struct {
	broker *rabbitmq.Broker
	db     *gorm.DB
}

func NewEventBus(db *gorm.DB, host string) (model.EventBus, error) {
	broker, err := rabbitmq.NewBroker(host)
	if err != nil {
		return nil, err
	}

	return &eventBus{db: db, broker: broker}, nil
}

func (bus *eventBus) PublishGameChanges(gameId uuid.UUID) error {
	media := model.Media{}
	game := model.Game{}

	var genres []model.GameGenre
	var tags []model.GameTag

	err := bus.db.Model(model.Game{}).Where("id = ?", gameId).First(&game).Error
	if err != nil {
		return err
	}

	err = bus.db.Model(model.Media{}).Where("id = ?", gameId).First(&media).Error
	if err != nil {
		return err
	}

	if len(game.Tags) > 0 {
		tt := toPgArray(game.Tags)
		err = bus.db.Model(model.GameTag{}).Where("id in (?)", tt).Find(&tags).Error
		if err != nil {
			return err
		}
	}

	if len(game.GenreAddition) > 0 || game.GenreMain != 0 {
		filter := game.GenreAddition
		filter = append(filter, game.GenreMain)
		err = bus.db.Model(model.GameGenre{}).Where("id in (?)", toPgArray(filter)).Find(&genres).Error
		if err != nil {
			return err
		}
	}

	var ratings model.GameRating
	if err := bus.db.Model(model.GameRating{}).Where("game_id = ?", gameId).First(&ratings).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return err
	}

	description := model.GameDescr{}
	if err := bus.db.Model(model.GameDescr{}).Where("game_id = ?", gameId).First(&description).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return err
	}

	gameObject := MapGameObject(&game, &media, tags, genres, ratings, description)
	return bus.broker.Publish("game_changed", gameObject, nil)
}

func toPgArray(array pq.Int64Array) string {
	var s []string
	for _, a := range array {
		s = append(s, strconv.FormatInt(a, 10))
	}
	return strings.Join(s, ",")
}

func (bus *eventBus) PublishGameDelete(gameId uuid.UUID) error {
	gameObject := &proto.GameDeleted{ID: gameId.String()}
	return bus.broker.Publish("game_deleted", gameObject, nil)
}

func MapGameObject(game *model.Game, media *model.Media, tags []model.GameTag, genre []model.GameGenre, ratings model.GameRating, descr model.GameDescr) *proto.GameObject {
	return &proto.GameObject{
		ID:                   game.ID.String(),
		Description:          MapLocalizedString(descr.Description),
		Name:                 game.Title,
		Title:                game.Title,
		Developer:            &proto.LinkObject{ID: "", Title: game.Developers},
		ReleaseDate:          game.ReleaseDate.Format(time.RFC3339),
		Tags:                 MapTags(tags),
		DisplayRemainingTime: game.DisplayRemainingTime,
		GenreMain:            MapGenre(game.GenreMain, genre),
		Genres:               MapGenres(game.GenreAddition, genre),
		Languages:            MapLanguages(game.Languages),
		Platforms:            MapPlatforms(game.Platforms),
		Requirements:         MapRequirements(game.Requirements),
		FeaturesControl:      game.FeaturesCtrl,
		Features:             game.FeaturesCommon,
		Media:                MapMedia(media),
		Ratings:              MapRatings(ratings),
		GameSite:             descr.GameSite,
		Reviews:              MapReviews(descr.Reviews),
		Tagline:              MapLocalizedString(descr.Tagline),
		Publisher:			  &proto.LinkObject{ID: "", Title: game.Publishers},
	}
}

func MapReviews(reviews game.GameReviews) []*proto.Review {
	if reviews == nil {
		return nil
	}

	var result []*proto.Review
	for _, review := range reviews {
		result = append(result, &proto.Review{
			Link:      review.Link,
			PressName: review.PressName,
			Quote:     review.Quote,
			Score:     review.Score,
		})
	}

	return result
}

func MapRatings(rating model.GameRating) *proto.Ratings {
	result := &proto.Ratings{}
	err := mapper.Map(rating, result)
	zap.L().Error("Can't map ratings", zap.Error(err))
	return result
}

func MapMedia(media *model.Media) *proto.Media {
	if media == nil {
		return nil
	}

	return &proto.Media{
		CoverImage:  MapJsonbToLocalizedString(media.CoverImage),
		CoverVideo:  MapJsonbToLocalizedString(media.CoverVideo),
		Trailers:    MapJsonbToLocalizedStringArray(media.Trailers),
		Screenshots: MapJsonbToLocalizedStringArray(media.Screenshots),
	}
}

func MapJsonbToLocalizedStringArray(jsonb model.JSONB) *proto.LocalizedStringArray {
	if jsonb == nil {
		return nil
	}
	return &proto.LocalizedStringArray{
		EN: jsonb.GetStringArray("en"),
		RU: jsonb.GetStringArray("ru"),
		FR: jsonb.GetStringArray("fr"),
		DE: jsonb.GetStringArray("de"),
		ES: jsonb.GetStringArray("es"),
		IT: jsonb.GetStringArray("it"),
		PT: jsonb.GetStringArray("pt"),
	}
}

func MapJsonbToLocalizedString(jsonb model.JSONB) *proto.LocalizedString {
	if jsonb == nil {
		return nil
	}
	return &proto.LocalizedString{
		EN: jsonb.GetString("en"),
		RU: jsonb.GetString("ru"),
		FR: jsonb.GetString("fr"),
		DE: jsonb.GetString("de"),
		ES: jsonb.GetString("es"),
		IT: jsonb.GetString("it"),
		PT: jsonb.GetString("pt"),
	}
}

func MapRequirements(requirements game.GameRequirements) *proto.Requirements {
	return &proto.Requirements{
		Windows: MapPlatformRequirements(requirements.Windows),
		Linux:   MapPlatformRequirements(requirements.Linux),
		MacOs:   MapPlatformRequirements(requirements.MacOs),
	}
}

func MapPlatformRequirements(requirements game.PlatformRequirements) *proto.PlatformRequirements {
	return &proto.PlatformRequirements{
		Minimal:     MapMachineRequirements(requirements.Minimal),
		Recommended: MapMachineRequirements(requirements.Recommended),
	}
}

func MapMachineRequirements(requirements game.MachineRequirements) *proto.MachineRequirements {
	return &proto.MachineRequirements{
		System:           requirements.System,
		Graphics:         requirements.Graphics,
		Other:            requirements.Other,
		Processor:        requirements.Processor,
		Ram:              requirements.Ram,
		RamDimension:     requirements.RamDimension,
		Sound:            requirements.Sound,
		Storage:          requirements.Storage,
		StorageDimension: requirements.StorageDimension,
	}
}

func MapPlatforms(platforms game.Platforms) *proto.Platforms {
	return &proto.Platforms{
		Linux:   platforms.Linux,
		MacOs:   platforms.MacOs,
		Windows: platforms.Windows,
	}
}

func MapLanguages(langs game.GameLangs) *proto.Languages {
	return &proto.Languages{
		EN: MapLanguage(langs.EN),
		RU: MapLanguage(langs.RU),
		DE: MapLanguage(langs.DE),
		ES: MapLanguage(langs.ES),
		FR: MapLanguage(langs.FR),
		IT: MapLanguage(langs.IT),
		PT: MapLanguage(langs.PT),
	}
}

func MapLanguage(lang game.Langs) *proto.Language {
	return &proto.Language{
		Interface: lang.Interface,
		Subtitles: lang.Subtitles,
		Voice:     lang.Voice,
	}
}

func MapGenres(genreIds []int64, genreModels []model.GameGenre) []*proto.TagObject {
	var result []*proto.TagObject
	for _, gId := range genreIds {
		for _, genre := range genreModels {
			if gId == genre.ID {
				result = append(result, &proto.TagObject{
					ID:   string(genre.ID),
					Name: MapLocalizedString(genre.Title),
				})
			}
		}
	}
	return result
}

func MapGenre(genre int64, genreModels []model.GameGenre) *proto.TagObject {
	for _, g := range genreModels {
		if g.ID == genre {
			return &proto.TagObject{
				ID:   string(g.ID),
				Name: MapLocalizedString(g.Title),
			}
		}
	}

	zap.L().Error(fmt.Sprintf("can't find genre id `%d` in arrays of genres `%v`", genre, genreModels))
	return nil
}

func MapTags(tags []model.GameTag) []*proto.TagObject {
	var result []*proto.TagObject
	for _, tag := range tags {
		obj := MapTag(tag)
		result = append(result, &obj)
	}
	return result
}

func MapTag(tag model.GameTag) proto.TagObject {
	return proto.TagObject{Name: MapLocalizedString(tag.Title), ID: string(tag.ID)}
}

func MapLocalizedString(s utils.LocalizedString) *proto.LocalizedString {
	return &proto.LocalizedString{RU: s.RU, EN: s.EN, DE: s.DE, ES: s.ES, FR: s.FR, IT: s.IT, PT: s.PT}
}
