package game

import (
	"database/sql/driver"
	"encoding/json"
	"github.com/pkg/errors"
)

func (p Features) Value() (driver.Value, error) {
	j, err := json.Marshal(p)
	return j, err
}

func (p *Features) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return errors.New("Type assertion .([]byte) failed.")
	}
	if err := json.Unmarshal(source, &p); err != nil {
		return err
	}
	return nil
}



func (p Platforms) Value() (driver.Value, error) {
	j, err := json.Marshal(p)
	return j, err
}

func (p *Platforms) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return errors.New("Type assertion .([]byte) failed.")
	}
	if err := json.Unmarshal(source, &p); err != nil {
		return err
	}
	return nil
}




func (p GameRequirements) Value() (driver.Value, error) {
	j, err := json.Marshal(p)
	return j, err
}

func (p *GameRequirements) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return errors.New("Type assertion .([]byte) failed.")
	}
	if err := json.Unmarshal(source, &p); err != nil {
		return err
	}
	return nil
}



func (p GameLangs) Value() (driver.Value, error) {
	j, err := json.Marshal(p)
	return j, err
}

func (p *GameLangs) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return errors.New("Type assertion .([]byte) failed.")
	}
	if err := json.Unmarshal(source, &p); err != nil {
		return err
	}
	return nil
}




func (p GameTags) Value() (driver.Value, error) {
	j, err := json.Marshal(p)
	return j, err
}

func (p *GameTags) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return errors.New("Type assertion .([]byte) failed.")
	}
	if err := json.Unmarshal(source, &p); err != nil {
		return err
	}
	return nil
}