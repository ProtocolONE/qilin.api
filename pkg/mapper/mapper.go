package mapper

import (
	"github.com/mitchellh/mapstructure"
)

//Map is mapping function from DTO (Data trasport object) to other (for example domain) model
func Map(from interface{}, to interface{}) error {
	err := mapstructure.WeakDecode(from, to)

	return err
}
