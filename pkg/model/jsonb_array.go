package model

import (
	"database/sql/driver"
	"encoding/json"
	"strings"
)

type JSONBArray []JSONB

//Value is marshaling function
func (j JSONBArray) Value() (driver.Value, error) {
	valueString, err := json.Marshal(j)
	js := string(valueString)
	js = strings.Replace(js, "[", "{", -1)
	js = strings.Replace(js, "]", "}", -1)
	return js, err
}

//Scan is unmarshaling function
func (j *JSONBArray) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &j); err != nil {
		return err
	}
	return nil
}
