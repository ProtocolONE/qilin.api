package utils

import (
	"database/sql/driver"
	"encoding/json"
	"github.com/pkg/errors"
	"gopkg.in/go-playground/validator.v9"
	"reflect"
)

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

func (p LocalizedString) Value() (driver.Value, error) {
	j, err := json.Marshal(p)
	return string(j), err
}

func (p *LocalizedString) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return errors.New("Type assertion .([]byte) failed.")
	}
	if err := json.Unmarshal(source, &p); err != nil {
		return err
	}
	return nil
}

func ValidateUrls(loc *LocalizedString) error {
	validate := validator.New()
	val := reflect.ValueOf(loc)
	elem := val.Elem()
	for i := 0; i < elem.NumField(); i++ {
		url := elem.Field(i).String()
		err := validate.Var(url, "omitempty,url")
		if err != nil {
			return errors.Wrap(err, "Validate localized URLs")
		}
	}
	return nil
}