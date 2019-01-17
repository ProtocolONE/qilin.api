package game

import (
	"github.com/satori/go.uuid"
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

	GameLangsDTO map[string]LangsDTO

	GameTagDTO struct {
		Id string                   `json:"id"`
		Title map[string]string     `json:"title"`
	}

	GameTagsDTO     []GameTagDTO

	GameFeaturesDTO struct {
		Common          []string    `json:"common"`
		Controllers     string      `json:"controllers"`
	}

	CreateGameDTO struct {
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
)