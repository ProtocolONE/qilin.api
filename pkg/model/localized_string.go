package model

// LocalizedString is helper object to hold localized string properties.
type LocalizedString struct {
	// english name
	EN string `json:"en"`

	// russian name
	RU string `json:"ru"`
}
