package game

type (
	MachineRequirements struct {
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

	PlatformRequirements struct {
		Minimal         MachineRequirements `json:"minimal"`
		Recommended     MachineRequirements `json:"recommended"`
	}

	GameRequirements struct {
		Windows     PlatformRequirements `json:"windows"`
		MacOs       PlatformRequirements `json:"macOs"`
		Linux       PlatformRequirements `json:"linux"`
	}

	Platforms struct {
		Windows bool    `json:"windows"`
		MacOs bool      `json:"macOs"`
		Linux bool      `json:"linux"`
	}

	Langs struct {
		Voice bool          `json:"voice"`
		Interface bool      `json:"interface"`
		Subtitles bool      `json:"subtitles"`
	}

	GameLangs struct {
		EN  Langs   `json:"en"`
		RU  Langs   `json:"ru"`
	}

	GameReviews []GameReview
	GameReview struct {
		PressName   string          `json:"pressName"`
		Link        string          `json:"link"`
		Score       string          `json:"score"`
		Quote       string          `json:"quote"`
	}

	Socials struct {
		Facebook    string          `json:"facebook"`
		Twitter     string          `json:"twitter"`
	}
)

