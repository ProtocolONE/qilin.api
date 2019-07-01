package model

// LocalizedString is helper object to hold localized string properties.
type LocalizedString struct {
	// english name
	EN string `json:"en"`

	// russian name
	RU string `json:"ru,omitempty"`

	// other languages
	FR string `json:"fr,omitempty"`
	ES string `json:"es,omitempty"`
	DE string `json:"de,omitempty"`
	IT string `json:"it,omitempty"`
	PT string `json:"pt,omitempty"`
}
