package model

import "time"

type KeyStream struct {
	// Unique key identifier in qilin system
	ID string `json:"id" validate:"required"`

	// Unique game ID in qilin system
	GameID string `bson:"gameId"`

	Source int32 `bson:"source"`

	Restrictions int32 `bson:"restrictions"`

	ActivationIP string `bson:"activation_ip"`

	// date of create key in system
	CreatedAt time.Time `json:"created_at"`

	// date of last update key in system
	UpdatedAt time.Time `json:"updated_at"`

	// date of activation key in system
	ActivatedAt time.Time `json:"activated_at"`

	// date of deletion key in system
	DeletedAt time.Time `json:"deleted_at"`
}
