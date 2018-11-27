package mongo

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"qilin-api/pkg"
	"time"
)

type game struct {
	ID bson.ObjectId `bson:"_id,omitempty"  validate:"required"`

	ExternalID *map[string]string `bson:"external_id"`

	Name string `bson:"name"`

	//UNDONE field
	//Description *localizedString `bson:"description"`

	// date of create merchant in system
	CreatedAt time.Time `bson:"created_at"`

	// date of last update merchant in system
	UpdatedAt time.Time `bson:"updated_at"`
}

func gameModelIndex() mgo.Index {
	return mgo.Index{
		Key:        []string{"name"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
}

func newGameModel(u *qilin.Game) (*game, error) {
	var id bson.ObjectId
	if u.ID == "" {
		id = bson.NewObjectId()
	} else {
		if !bson.IsObjectIdHex(u.ID) {
			return nil, fmt.Errorf("Given `%s` is not the ObjectId Hex", id)
		}
		id = bson.ObjectIdHex(u.ID)
	}

	game := &game{
		ID:         id,
		ExternalID: u.ExternalID,
		Name:       u.Name,
		//Description: u.Description,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}

	return game, nil
}

func (u *game) toQilinGame() *qilin.Game {
	return &qilin.Game{
		ID:         u.ID.Hex(),
		ExternalID: u.ExternalID,
		Name:       u.Name,
		//Description: u.Description,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
