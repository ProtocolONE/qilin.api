package model

import uuid "github.com/satori/go.uuid"

type KeyStream struct {
	Model
	Type         KeyStreamType
}

type KeyStreamProvider interface {
	Redeem() (Key, error)
	RedeemList(count int) ([]Key, error)
	AddKeys(codes []string) error
}

type KeyStreamService interface {
	Get(streamId uuid.UUID) (KeyStreamProvider, error)
	Create(keyProviderType KeyStreamType) (uuid.UUID, error)
}
