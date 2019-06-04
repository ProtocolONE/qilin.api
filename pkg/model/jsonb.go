package model

import (
	"database/sql/driver"

	"encoding/json"
)

type JSONB map[string]interface{}

//Value is marshaling function
func (j JSONB) Value() (driver.Value, error) {
	valueString, err := json.Marshal(j)
	return string(valueString), err
}

//Scan is unmarshaling function
func (j *JSONB) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &j); err != nil {
		return err
	}
	return nil
}

func (j JSONB) GetString(key string) string {
	return j[key].(string)
}

func (j JSONB) GetStringArray(key string) []string {
	if j == nil {
		return nil
	}
	return j[key].([]string)
}
