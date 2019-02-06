package utils

import (
	"database/sql/driver"
	"encoding/json"
	"github.com/pkg/errors"
)

// LocalizedString is helper object to hold localized string properties.
type LocalizedStringArray struct {
	// english name
	EN []string `json:"en"`

	// russian name
	RU []string `json:"ru"`
}

func (p LocalizedStringArray) Value() (driver.Value, error) {
	j, err := json.Marshal(p)
	return string(j), err
}

func (p *LocalizedStringArray) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return errors.New("Type assertion .([]byte) failed.")
	}
	if err := json.Unmarshal(source, &p); err != nil {
		return err
	}
	return nil
}