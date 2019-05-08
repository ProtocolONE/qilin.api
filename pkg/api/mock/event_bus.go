package mock

import (
	"github.com/satori/go.uuid"
	"qilin-api/pkg/model"
)

type eventBus struct {
}

func (eventBus) PublishGameChanges(gameId uuid.UUID) error {
	return nil
}

func (eventBus) PublishGameDelete(gameId uuid.UUID) error {
	return nil
}

func NewEventBus() model.EventBus {
	return &eventBus{}
}
