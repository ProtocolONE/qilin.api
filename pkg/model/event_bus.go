package model

import "github.com/satori/go.uuid"

type EventBus interface {
	PublishGameChanges(gameId uuid.UUID) error
	PublishGameDelete(gameId uuid.UUID) error
}
