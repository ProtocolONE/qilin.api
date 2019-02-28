package sys

import (
	"context"
	"encoding/json"
	"github.com/centrifugal/gocent"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Notifier interface {
	Publish(channel string, payload []byte) error
	SendMessage(channel string, message NotifyMessage) error
}

type notifierImpl struct {
	secret  string
	address string
	client  *gocent.Client
}

type NotifyMessage struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Body     string `json:"body"`
	DateTime string `json:"dateTime"`
}

func NewNotifier(secret string, addr string) (Notifier, error) {
	if secret == "" {
		return nil, errors.New("ApiKey couldn't be empty")
	}

	client := gocent.New(gocent.Config{
		Addr: addr,
		Key:  secret,
	})

	notifier := &notifierImpl{secret: secret, client: client}
	return notifier, nil
}

func (n *notifierImpl) Publish(channel string, payload []byte) error {
	return n.client.Publish(context.TODO(), channel, payload)
}

func (n *notifierImpl) SendMessage(channel string, message NotifyMessage) error {
	bytes, err := json.Marshal(message)
	if err != nil {
		zap.L().Error("[SendMessage] Can't marshal json", zap.Error(err))
		return err
	}
	err = n.Publish(channel, bytes)

	if err != nil {
		zap.L().Error("[SendMessage] Error during publish", zap.Error(err))
	}
	return err
}
