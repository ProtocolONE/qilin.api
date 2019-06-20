package model

import uuid "github.com/satori/go.uuid"

type KeyPackage struct {
	Model
	Name          string
	KeyStreamID   uuid.UUID
	PackageID     uuid.UUID
	KeyStreamType KeyStreamType
}

type KeyStreamType string

func (s KeyStreamType) String() string {
	return string(s)
}

const (
	ListKeyStream      KeyStreamType = "key_list"
	PlatformKeysStream KeyStreamType = "key_platform"
)

type KeyPackageService interface {
	Create(packageId uuid.UUID, name string, providerType KeyStreamType) (*KeyPackage, error)
	Update(keyPackageId uuid.UUID, name string) (*KeyPackage, error)
	List(packageId uuid.UUID) ([]KeyPackage, error)
	Get(keyPackageId uuid.UUID) (*KeyPackage, error)
}

type KeyListService interface {
	AddKeys(keyPackageId uuid.UUID, keys []string) error
}
