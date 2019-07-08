package orm

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/model"
)

type keyStreamService struct {
	db *Database
}

var allowedStreamProviders = []model.KeyStreamType{model.ListKeyStream, model.PlatformKeysStream}

func NewKeyStreamService(db *Database) model.KeyStreamService {
	return &keyStreamService{db}
}

func (service *keyStreamService) Get(streamId uuid.UUID) (model.KeyStreamProvider, error) {
	stream := &model.KeyStream{}
	err := service.db.DB().Model(model.KeyStream{}).Where("id = ?", streamId).First(stream).Error

	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, NewServiceError(http.StatusNotFound, err)
		}
		return nil, NewServiceError(http.StatusInternalServerError, err)
	}

	switch stream.Type {
	case model.ListKeyStream:
		return NewKeyListProvider(stream.ID, service.db)
	case model.PlatformKeysStream:
		return NewPlatformKeyProvider(stream.ID, service.db)
	default:
		return nil, NewServiceError(http.StatusBadRequest, errors.New(fmt.Sprintf("Unknown stream provider `%s`", stream.Type)))
	}
}

func (service *keyStreamService) Create(keyProviderType model.KeyStreamType) (uuid.UUID, error) {
	if checkTypeIsAllowed(keyProviderType) == false {
		return uuid.Nil, NewServiceErrorf(http.StatusUnprocessableEntity, "Type `%s` is not allowed", keyProviderType)
	}

	stream := &model.KeyStream{
		Type: keyProviderType,
	}

	stream.ID = uuid.NewV4()
	err := service.db.DB().Model(model.KeyStream{}).Create(stream).Error

	if err != nil {
		return uuid.Nil, NewServiceError(http.StatusInternalServerError, err)
	}

	return stream.ID, nil
}

func checkTypeIsAllowed(streamType model.KeyStreamType) bool {
	for _, s := range allowedStreamProviders {
		if s == streamType {
			return true
		}
	}

	return false
}
