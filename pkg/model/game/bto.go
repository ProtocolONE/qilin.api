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

	GameLangs map[string]Langs

	LocalizedString map[string]string

	Tag struct {
		Id      string              `json:"id"`
		Title   LocalizedString     `json:"title"`
	}

	GameTags []Tag

	Features struct {
		Common          []string    `json:"common"`
		Controllers     string      `json:"controllers"`
	}

)

