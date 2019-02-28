package game

import (
	"github.com/satori/go.uuid"
	"qilin-api/pkg/model"
	bto "qilin-api/pkg/model/game"
	"qilin-api/pkg/model/utils"
	"time"
)

type (
	MachineRequirementsDTO struct {
		System           string `json:"system"`
		Processor        string `json:"processor"`
		Graphics         string `json:"graphics"`
		Sound            string `json:"sound"`
		Ram              int    `json:"ram"`
		RamDimension     string `json:"ramdimension"`
		Storage          int    `json:"storage"`
		StorageDimension string `json:"storagedimension"`
		Other            string `json:"other"`
	}

	PlatformRequirementsDTO struct {
		Minimal     MachineRequirementsDTO `json:"minimal" validate:"required,dive"`
		Recommended MachineRequirementsDTO `json:"recommended" validate:"required,dive"`
	}

	GameRequirementsDTO struct {
		Windows PlatformRequirementsDTO `json:"windows" validate:"required,dive"`
		MacOs   PlatformRequirementsDTO `json:"macOs" validate:"required,dive"`
		Linux   PlatformRequirementsDTO `json:"linux" validate:"required,dive"`
	}

	GamePlatformDTO struct {
		Windows bool `json:"windows"`
		MacOs   bool `json:"macOs"`
		Linux   bool `json:"linux"`
	}

	LangsDTO struct {
		Voice     bool `json:"voice"`
		Interface bool `json:"interface"`
		Subtitles bool `json:"subtitles"`
	}

	GameLangsDTO struct {
		EN LangsDTO `json:"en" validate:"dive"`
		RU LangsDTO `json:"ru" validate:"dive"`
		FR LangsDTO `json:"fr" validate:"dive"`
		ES LangsDTO `json:"es" validate:"dive"`
		DE LangsDTO `json:"de" validate:"dive"`
		IT LangsDTO `json:"it" validate:"dive"`
		PT LangsDTO `json:"pt" validate:"dive"`
	}

	GameTagDTO struct {
		Id    int                   `json:"id" validate:"required"`
		Title utils.LocalizedString `json:"title" validate:"dive"`
	}

	RatingDescriptorDTO struct {
		Id     uint                  `json:"id" validate:"required"`
		Title  utils.LocalizedString `json:"title" validate:"dive"`
		System string                `json:"system"`
	}

	GameFeaturesDTO struct {
		Common      []string `json:"common"`
		Controllers string   `json:"controllers"`
	}

	BaseGameDTO struct {
		InternalName         string              `json:"internalName"`
		Title                string              `json:"title"`
		Developers           string              `json:"developers"`
		Publishers           string              `json:"publishers"`
		ReleaseDate          time.Time           `json:"releaseDate" validate:"required"`
		DisplayRemainingTime bool                `json:"displayRemainingTime"`
		AchievementOnProd    bool                `json:"achievementOnProd"`
		Features             GameFeaturesDTO     `json:"features" validate:"required,dive"`
		Platforms            GamePlatformDTO     `json:"platforms" validate:"required,dive"`
		Requirements         GameRequirementsDTO `json:"requirements" validate:"required,dive"`
		Languages            GameLangsDTO        `json:"languages" validate:"required,dive"`
	}

	GameGenreDTO struct {
		Main        int64        `json:"main"`
		Addition    []int64      `json:"addition" validate:"required"`
	}

	GameDTO struct {
		ID uuid.UUID `json:"id"`
		BaseGameDTO
		Genres GameGenreDTO `json:"genres" validate:"required,dive"`
		Tags   []int64      `json:"tags" validate:"required"`
	}

	UpdateGameDTO struct {
		BaseGameDTO
		Genres GameGenreDTO `json:"genres" validate:"required,dive"`
		Tags   []int64      `json:"tags" validate:"required"`
	}

	GamePriceDTO struct {
		Price    float64 `json:"price" validate:"required"`
		Currency string  `json:"currency" validate:"required"`
	}

	ShortGameInfoDTO struct {
		ID           uuid.UUID     `json:"id"`
		InternalName string        `json:"internalName"`
		Icon         string        `json:"icon"`
		Genres       GameGenreDTO  `json:"genres"`
		ReleaseDate  time.Time     `json:"releaseDate"`
		Prices       GamePriceDTO  `json:"prices"`
	}

	DescrReview struct {
		PressName string `json:"pressName"`
		Link      string `json:"link"`
		Score     string `json:"score"`
		Quote     string `json:"quote"`
	}

	Socials struct {
		Facebook string `json:"facebook"`
		Twitter  string `json:"twitter"`
	}

	GameDescrDTO struct {
		Tagline               utils.LocalizedString `json:"tagline" validate:"dive"`
		Description           utils.LocalizedString `json:"description" validate:"dive"`
		Reviews               []DescrReview         `json:"reviews" validate:"required,dive"`
		AdditionalDescription string                `json:"additionalDescription"`
		GameSite              string                `json:"gameSite"`
		Socials               Socials               `json:"socials"`
	}
)

func mapReqs(r *bto.MachineRequirements) MachineRequirementsDTO {
	return MachineRequirementsDTO{
		System:           r.System,
		Processor:        r.Processor,
		Graphics:         r.Graphics,
		Sound:            r.Sound,
		Ram:              r.Ram,
		RamDimension:     r.RamDimension,
		Storage:          r.Storage,
		StorageDimension: r.StorageDimension,
		Other:            r.Other,
	}
}

func mapReqsBTO(r *MachineRequirementsDTO) bto.MachineRequirements {
	return bto.MachineRequirements{
		System:           r.System,
		Processor:        r.Processor,
		Graphics:         r.Graphics,
		Sound:            r.Sound,
		Ram:              r.Ram,
		RamDimension:     r.RamDimension,
		Storage:          r.Storage,
		StorageDimension: r.StorageDimension,
		Other:            r.Other,
	}
}

func mapGameInfo(game *model.Game, service model.GameService) (dst *GameDTO, err error) {
	return &GameDTO{
		ID: game.ID,
		BaseGameDTO: BaseGameDTO{
			InternalName:         game.InternalName,
			Title:                game.Title,
			Developers:           game.Developers,
			Publishers:           game.Publishers,
			ReleaseDate:          game.ReleaseDate,
			DisplayRemainingTime: game.DisplayRemainingTime,
			AchievementOnProd:    game.AchievementOnProd,
			Features:             GameFeaturesDTO{Common: game.FeaturesCommon, Controllers: game.FeaturesCtrl},
			Platforms: GamePlatformDTO{
				Windows: game.Platforms.Windows,
				MacOs:   game.Platforms.MacOs,
				Linux:   game.Platforms.Linux,
			},
			Requirements: GameRequirementsDTO{
				Windows: PlatformRequirementsDTO{
					Minimal:     mapReqs(&game.Requirements.Windows.Minimal),
					Recommended: mapReqs(&game.Requirements.Windows.Recommended)},
				Linux: PlatformRequirementsDTO{
					Minimal:     mapReqs(&game.Requirements.Linux.Minimal),
					Recommended: mapReqs(&game.Requirements.Linux.Recommended)},
				MacOs: PlatformRequirementsDTO{
					Minimal:     mapReqs(&game.Requirements.MacOs.Minimal),
					Recommended: mapReqs(&game.Requirements.MacOs.Recommended)},
			},
			Languages: GameLangsDTO{
				EN: LangsDTO{game.Languages.EN.Voice, game.Languages.EN.Interface, game.Languages.EN.Subtitles},
				RU: LangsDTO{game.Languages.RU.Voice, game.Languages.RU.Interface, game.Languages.RU.Subtitles},
			},
		},
		Genres: GameGenreDTO{
			Main: game.GenreMain,
			Addition: game.GenreAddition,
		},
		Tags:  game.Tags,
	}, nil
}

func mapGameInfoBTO(game *UpdateGameDTO) (dst model.Game) {
	return model.Game{
		InternalName:         game.InternalName,
		Title:                game.Title,
		Developers:           game.Developers,
		Publishers:           game.Publishers,
		ReleaseDate:          game.ReleaseDate,
		DisplayRemainingTime: game.DisplayRemainingTime,
		AchievementOnProd:    game.AchievementOnProd,
		FeaturesCommon:       game.Features.Common,
		FeaturesCtrl:         game.Features.Controllers,
		Platforms: bto.Platforms{
			Windows: game.Platforms.Windows,
			MacOs:   game.Platforms.MacOs,
			Linux:   game.Platforms.Linux,
		},
		Requirements: bto.GameRequirements{
			Windows: bto.PlatformRequirements{
				Minimal:     mapReqsBTO(&game.Requirements.Windows.Minimal),
				Recommended: mapReqsBTO(&game.Requirements.Windows.Recommended)},
			Linux: bto.PlatformRequirements{
				Minimal:     mapReqsBTO(&game.Requirements.Linux.Minimal),
				Recommended: mapReqsBTO(&game.Requirements.Linux.Recommended)},
			MacOs: bto.PlatformRequirements{
				Minimal:     mapReqsBTO(&game.Requirements.MacOs.Minimal),
				Recommended: mapReqsBTO(&game.Requirements.MacOs.Recommended)},
		},
		Languages: bto.GameLangs{
			EN: bto.Langs{game.Languages.EN.Voice, game.Languages.EN.Interface, game.Languages.EN.Subtitles},
			RU: bto.Langs{game.Languages.RU.Voice, game.Languages.RU.Interface, game.Languages.RU.Subtitles},
		},
		GenreMain: game.Genres.Main,
		GenreAddition: game.Genres.Addition,
		Tags:  game.Tags,
	}
}
