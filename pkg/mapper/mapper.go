package mapper

import (
	"github.com/mitchellh/mapstructure"
	maper "gopkg.in/jeevatkm/go-model.v1"
)

//Map is mapping function from DTO (Data trasport object) to other (for example domain) model
func Map(from interface{}, to interface{}) error {
	input, err := maper.Map(from)
	if err != nil {
		return err
	}

	err = mapstructure.Decode(input, to)
	if err != nil {
		return err
	}

	return nil
}
