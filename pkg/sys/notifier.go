package sys

import (
	"context"
	"encoding/json"
	"github.com/centrifugal/gocent"
)

type Notifier interface {
	Publish(channel string, payload []byte) error
	SendMessage(message NotifyMessage) error
}

type NotifierImpl struct {
	secret  string
	address string
	client  *gocent.Client
}

type NotifyMessage struct {
	Title    string `json:"title"`
	Body     string `json:"body"`
	DateTime string `json:"dateTime"`
}

const OnboardingNotify string = "onboarding:messages"

func NewNotifier(secret string, addr string) (notifier Notifier) {
	client := gocent.New(gocent.Config{
		Addr: addr,
		Key:  secret,
	})
	notifier = &NotifierImpl{secret: secret, client: client}
	return
}

func (n *NotifierImpl) Publish(channel string, payload []byte) error {
	return n.client.Publish(context.TODO(), channel, payload)
}

func (n *NotifierImpl) SendMessage(message NotifyMessage) error {
	bytes, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return n.Publish(OnboardingNotify, bytes)
}
