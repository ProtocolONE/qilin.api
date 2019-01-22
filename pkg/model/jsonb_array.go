package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type JSONBArray []JSONB

func (p JSONBArray) Value() (driver.Value, error) {
	j, err := json.Marshal(p)
	return j, err
}

func (p *JSONBArray) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return errors.New("Type assertion .([]byte) failed.")
	}

	var i interface{}
	if err := json.Unmarshal(source, &i); err != nil {
		return err
	}

	*p, ok = i.(JSONBArray)
	if !ok {
		return errors.New("Type assertion .(JSONBArray) failed.")
	}

	return nil
}
