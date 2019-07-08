package orm

import (
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/model"
	"time"
)

type keyListProvider struct {
	streamId uuid.UUID
	db       *Database
}

func NewKeyListProvider(streamId uuid.UUID, db *Database) (model.KeyStreamProvider, error) {
	keyStream := model.KeyStream{}
	err := db.DB().Model(model.KeyStream{}).Where("id = ?", streamId).First(&keyStream).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, NewServiceErrorf(http.StatusNotFound, "Key Stream with id `%s` not found", streamId)
		}
		return nil, NewServiceError(http.StatusInternalServerError, err)
	}

	if keyStream.Type != model.ListKeyStream {
		return nil, NewServiceErrorf(http.StatusBadRequest, "Key Stream with id `%s` is not key list stream", streamId)
	}

	return &keyListProvider{streamId: streamId, db: db}, nil
}

func (provider *keyListProvider) Redeem() (model.Key, error) {
	key := model.Key{}
	err := provider.db.DB().Model(model.Key{}).Where("key_stream_id = ?", provider.streamId).Where("redeem_time IS NULL").First(&key).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return key, NewServiceErrorf(http.StatusBadRequest, "No free keys in the stream `%s`", provider.streamId)
		}
		return key, NewServiceError(http.StatusInternalServerError, err)
	}

	t := time.Now().UTC()
	key.RedeemTime = &t
	err = provider.db.DB().Model(model.Key{}).Update(key).Error
	if err != nil {
		return key, NewServiceError(http.StatusInternalServerError, err)
	}

	return key, nil
}

func (provider *keyListProvider) RedeemList(count int) ([]model.Key, error) {
	transaction := provider.db.DB().Begin()

	var keys []model.Key
	err := transaction.Model(model.Key{}).Where("key_stream_id = ?", provider.streamId).Where("redeem_time IS NULL").Limit(count).Find(&keys).Error
	if err != nil {
		transaction.Rollback()
		return nil, NewServiceError(http.StatusInternalServerError, err)
	}

	if len(keys) != count {
		transaction.Rollback()
		return nil, NewServiceErrorf(http.StatusBadRequest, "Keys not enough. Stream have `%d` keys.", len(keys))
	}
	
	for _, key := range keys {
		t := time.Now().UTC()
		key.RedeemTime = &t
		err := transaction.Model(key).Update(key).Error
		if err != nil {
			transaction.Rollback()
			return nil, NewServiceError(http.StatusInternalServerError, err)
		}
	}

	err = transaction.Commit().Error
	if err != nil {
		return nil, NewServiceError(http.StatusInternalServerError, err)
	}

	return keys, nil
}

func (provider *keyListProvider) AddKeys(codes []string) error {
	transaction := provider.db.DB().Begin()

	for _, code := range codes {
		key := model.Key{
			ActivationCode: code,
			KeyStreamID:    provider.streamId,
		}
		key.ID = uuid.NewV4()
		err := transaction.Model(model.Key{}).Create(&key).Error
		if err != nil {
			transaction.Rollback()
			return NewServiceError(http.StatusBadRequest, err)
		}
	}

	err := transaction.Commit().Error
	if err != nil {
		return NewServiceError(http.StatusBadRequest, err)
	}
	return nil
}
