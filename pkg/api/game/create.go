package game

import (
	"github.com/labstack/echo"
	"net/http"
	"qilin-api/pkg/model"
	"qilin-api/pkg/model/game"
)

func mapReqs(r* game.MachineRequirements) MachineRequirementsDTO {
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

func mapLangs(src* game.GameLangs) (dst GameLangsDTO) {
	/*dst = make(GameLangsDTO)
	for k, v := range *src {
		dst[k] = LangsDTO{
			Voice: v.Voice,
			Interface: v.Interface,
			Subtitles: v.Subtitles,
		}
	}*/
	return dst
}

func mapTags(tags []model.GameTag) (dst GameTagsDTO) {
	dst = GameTagsDTO{}
	for _, v := range tags {
		dst = append(dst, GameTagDTO{Id: v.ID, Title: v.Title})
	}
	return dst
}

func (api *Router) create(ctx echo.Context) error {
	internalName := ctx.FormValue("internalName")

	game, err := api.gameService.CreateGame(internalName)
	if err != nil {
		return err
	}
	tags, err := api.gameService.GetTags(game.Tags)
	if err != nil {
		return err
	}
	dto := GameDTO{
		ID: game.ID,
		InternalName: game.InternalName,
		Title: game.Title,
		Developers: game.Developers,
		Publishers: game.Publishers,
		ReleaseDate: game.ReleaseDate,
		DisplayRemainingTime: game.DisplayRemainingTime,
		AchievementOnProd: game.AchievementOnProd,
		/*Features: GameFeaturesDTO{Common: game.Features.Common, Controllers: game.Features.Controllers},
		Requirements: GameRequirementsDTO{
			Windows:PlatformRequirementsDTO{
				Minimal: mapReqs(&game.Requirements.Windows.Minimal),
				Recommended: mapReqs(&game.Requirements.Windows.Recommended)},
			Linux:PlatformRequirementsDTO{
				Minimal: mapReqs(&game.Requirements.Windows.Minimal),
				Recommended: mapReqs(&game.Requirements.Windows.Recommended)},
			MacOs:PlatformRequirementsDTO{
				Minimal: mapReqs(&game.Requirements.Windows.Minimal),
				Recommended: mapReqs(&game.Requirements.Windows.Recommended)},
			},
		Languages: mapLangs(&game.Languages),*/
		//Genre: mapTags(&game.Genre),
		Tags: mapTags(tags),
	}
	return ctx.JSON(http.StatusCreated, dto)
}
