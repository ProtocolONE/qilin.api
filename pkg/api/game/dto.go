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
		System              string  `json:"system"`
		Processor           string  `json:"processor"`
		Graphics            string  `json:"graphics"`
		Sound               string  `json:"sound"`
		Ram                 int     `json:"ram"`
		RamDimension        string  `json:"ramdimension"`
		Storage             int     `json:"storage"`
		StorageDimension    string  `json:"storagedimension"`
		Other               string  `json:"other"`
	}

	PlatformRequirementsDTO struct {
		Minimal         MachineRequirementsDTO `json:"minimal"`
		Recommended     MachineRequirementsDTO `json:"recommended"`
	}

	GameRequirementsDTO struct {
		Windows     PlatformRequirementsDTO `json:"windows"`
		MacOs       PlatformRequirementsDTO `json:"macOs"`
		Linux       PlatformRequirementsDTO `json:"linux"`
	}

	GamePlatformDTO struct {
		Windows bool    `json:"windows"`
		MacOs bool      `json:"macOs"`
		Linux bool      `json:"linux"`
	}

	LangsDTO struct {
		Voice bool          `json:"voice"`
		Interface bool      `json:"interface"`
		Subtitles bool      `json:"subtitles"`
	}

	GameLangsDTO struct {
		EN LangsDTO         `json:"en"`
		RU LangsDTO         `json:"ru"`
	}

	GameTagDTO struct {
		Id          string                    `json:"id"`
		Title       utils.LocalizedString     `json:"title"`
	}

	GameTagsDTO     []GameTagDTO

	GameFeaturesDTO struct {
		Common          []string    `json:"common"`
		Controllers     string      `json:"controllers"`
	}

	GameDTO struct {
		ID                   uuid.UUID           `json:"id"`
		InternalName         string              `json:"InternalName"`
		Title                string              `json:"title"`
		Developers           string              `json:"developers"`
		Publishers           string              `json:"publishers"`
		ReleaseDate          time.Time           `json:"releaseDate"`
		DisplayRemainingTime bool                `json:"displayRemainingTime"`
		AchievementOnProd    bool                `json:"achievementOnProd"`
		Features             GameFeaturesDTO     `json:"features"`
		Platforms            GamePlatformDTO     `json:"platforms"`
		Requirements         GameRequirementsDTO `json:"requirements"`
		Languages            GameLangsDTO        `json:"languages"`
		Genre                GameTagsDTO         `json:"genre"`
		Tags                 GameTagsDTO         `json:"tags"`
	}

	GamePriceDTO struct {
		Price               float64
		Currency            string
	}

	GamePricesDTO       []GamePriceDTO

	ShortGameInfoDTO struct {
		ID                  uuid.UUID           `json:"id"`
		InternalName        string              `json:"internalName"`
		Icon                string              `json:"icon"`
		Genre               GameTagsDTO         `json:"genre"`
		ReleaseDate         time.Time           `json:"releaseDate"`
		Prices              GamePricesDTO       `json:"prices"`
	}
)

func mapReqs(r* bto.MachineRequirements) MachineRequirementsDTO {
	return MachineRequirementsDTO{
		System:              r.System,
		Processor:           r.Processor,
		Graphics:            r.Graphics,
		Sound:               r.Sound,
		Ram:                 r.Ram,
		RamDimension:        r.RamDimension,
		Storage:             r.Storage,
		StorageDimension:    r.StorageDimension,
		Other:               r.Other,
	}
}

func mapReqsBTO(r* MachineRequirementsDTO) bto.MachineRequirements {
	return bto.MachineRequirements{
		System:              r.System,
		Processor:           r.Processor,
		Graphics:            r.Graphics,
		Sound:               r.Sound,
		Ram:                 r.Ram,
		RamDimension:        r.RamDimension,
		Storage:             r.Storage,
		StorageDimension:    r.StorageDimension,
		Other:               r.Other,
	}
}

func mapGenres(tags []model.GameGenre) (dst GameTagsDTO) {
	dst = GameTagsDTO{}
	for _, v := range tags {
		dst = append(dst, GameTagDTO{Id: v.ID, Title: v.Title})
	}
	return dst
}

func mapTags(tags []model.GameTag) (dst GameTagsDTO) {
	dst = GameTagsDTO{}
	for _, v := range tags {
		dst = append(dst, GameTagDTO{Id: v.ID, Title: v.Title})
	}
	return dst
}

func mapTagsBTO(tags []GameTagDTO) (bto []string) {
	bto = []string{}
	for _, v := range tags {
		bto = append(bto, v.Id)
	}
	return bto
}

func mapGameInfo(game *model.Game, service model.GameService) (dst *GameDTO, err error) {

	tags, err := service.GetTags(game.Tags)
	if err != nil {
		return nil, err
	}
	ganres, err := service.GetGenres(game.Genre)
	if err != nil {
		return nil, err
	}
	return &GameDTO{
		ID: game.ID,
		InternalName: game.InternalName,
		Title: game.Title,
		Developers: game.Developers,
		Publishers: game.Publishers,
		ReleaseDate: game.ReleaseDate,
		DisplayRemainingTime: game.DisplayRemainingTime,
		AchievementOnProd: game.AchievementOnProd,
		Features: GameFeaturesDTO{Common: game.FeaturesCommon, Controllers: game.FeaturesCtrl},
		Platforms: GamePlatformDTO{
			Windows: game.Platforms.Windows,
			MacOs: game.Platforms.MacOs,
			Linux: game.Platforms.Linux,
		},
		Requirements: GameRequirementsDTO{
			Windows: PlatformRequirementsDTO{
				Minimal: mapReqs(&game.Requirements.Windows.Minimal),
				Recommended: mapReqs(&game.Requirements.Windows.Recommended)},
			Linux: PlatformRequirementsDTO{
				Minimal: mapReqs(&game.Requirements.Linux.Minimal),
				Recommended: mapReqs(&game.Requirements.Linux.Recommended)},
			MacOs: PlatformRequirementsDTO{
				Minimal: mapReqs(&game.Requirements.MacOs.Minimal),
				Recommended: mapReqs(&game.Requirements.MacOs.Recommended)},
		},
		Languages: GameLangsDTO{
			EN: LangsDTO{game.Languages.EN.Voice, game.Languages.EN.Interface, game.Languages.EN.Subtitles},
			RU: LangsDTO{game.Languages.RU.Voice, game.Languages.RU.Interface, game.Languages.RU.Subtitles},
		},
		Genre: mapGenres(ganres),
		Tags: mapTags(tags),
	}, nil
}

func mapGameInfoBTO(game *GameDTO) (dst model.Game) {

	return model.Game{
		ID: game.ID,
		InternalName: game.InternalName,
		Title: game.Title,
		Developers: game.Developers,
		Publishers: game.Publishers,
		ReleaseDate: game.ReleaseDate,
		DisplayRemainingTime: game.DisplayRemainingTime,
		AchievementOnProd: game.AchievementOnProd,
		FeaturesCommon: game.Features.Common,
		FeaturesCtrl: game.Features.Controllers,
		Platforms: bto.Platforms{
			Windows: game.Platforms.Windows,
			MacOs: game.Platforms.MacOs,
			Linux: game.Platforms.Linux,
		},
		Requirements: bto.GameRequirements{
			Windows: bto.PlatformRequirements{
				Minimal: mapReqsBTO(&game.Requirements.Windows.Minimal),
				Recommended: mapReqsBTO(&game.Requirements.Windows.Recommended)},
			Linux: bto.PlatformRequirements{
				Minimal: mapReqsBTO(&game.Requirements.Linux.Minimal),
				Recommended: mapReqsBTO(&game.Requirements.Linux.Recommended)},
			MacOs: bto.PlatformRequirements{
				Minimal: mapReqsBTO(&game.Requirements.MacOs.Minimal),
				Recommended: mapReqsBTO(&game.Requirements.MacOs.Recommended)},
		},
		Languages: bto.GameLangs{
			EN: bto.Langs{game.Languages.EN.Voice, game.Languages.EN.Interface, game.Languages.EN.Subtitles},
			RU: bto.Langs{game.Languages.RU.Voice, game.Languages.RU.Interface, game.Languages.RU.Subtitles},
		},
		Genre: mapTagsBTO(game.Genre),
		Tags: mapTagsBTO(game.Tags),
	}
}