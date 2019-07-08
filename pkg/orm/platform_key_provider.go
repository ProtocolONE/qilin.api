package orm

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
	"math/rand"
	"net/http"
	"qilin-api/pkg/model"
	"time"
)

type platformKeyProvider struct {
	db       *Database
	streamId uuid.UUID
}

func (provider *platformKeyProvider) Redeem() (model.Key, error) {
	transaction := provider.db.DB().Begin()
	key, err := redeem(provider.streamId, transaction)
	if err != nil {
		transaction.Rollback()
		return key, err
	}

	return key, transaction.Commit().Error
}

func generateCode() string {
	// Example: AAAA-BBBB-CCCC-DDDD
	return fmt.Sprintf("%s-%s-%s-%s", RandStringRunes(4), RandStringRunes(4), RandStringRunes(4), RandStringRunes(4))
}

func redeem(streamId uuid.UUID, transaction *gorm.DB) (model.Key, error) {
	key := model.Key{
		KeyStreamID: streamId,
	}
	key.ID = uuid.NewV4()

	infiniteLoopIndex := 0
	for {
		key.ActivationCode = generateCode()
		t := time.Now().UTC()
		key.RedeemTime = &t
		var keyExist model.Key
		err := transaction.Model(model.Key{}).Where("key_stream_id = ? AND activation_code = ?",
			streamId,
			key.ActivationCode).First(&keyExist).Error
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				break
			}
			return key, NewServiceError(http.StatusInternalServerError, errors.Wrap(err, "Trying to check activation code before exist"))
		}

		if infiniteLoopIndex >= 100 {
			return key, NewServiceError(http.StatusInternalServerError, errors.Wrap(err, "Infinite loop break condition."))
		}
		infiniteLoopIndex++
		zap.L().Info(fmt.Sprintf("KeyStream `%s` already contains activation_code `%s`. Trying again", streamId, key.ActivationCode))
	}

	err := transaction.Model(model.Key{}).Create(&key).Error
	if err != nil {
		return key, err
	}

	return key, nil
}

func (service *platformKeyProvider) RedeemList(count int) ([]model.Key, error) {
	var keys []model.Key
	transaction := service.db.DB().Begin()
	for i := 0; i < count; i++ {
		key, err := redeem(service.streamId, transaction)
		if err != nil {
			transaction.Rollback()
			return keys, err
		}
		keys = append(keys, key)
	}

	return keys, transaction.Commit().Error
}

func (platformKeyProvider) AddKeys(codes []string) error {
	return NewServiceError(http.StatusBadRequest, "Not implemented for PlatformsKey")
}

func NewPlatformKeyProvider(keyStreamId uuid.UUID, database *Database) (model.KeyStreamProvider, error) {
	keyStream := model.KeyStream{}
	err := database.DB().Model(model.KeyStream{}).Where("id = ?", keyStreamId).First(&keyStream).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, NewServiceErrorf(http.StatusNotFound, "Key stream with id `%s` not found", keyStreamId)
		}
		return nil,  NewServiceError(http.StatusInternalServerError, err)
	}

	if keyStream.Type != model.PlatformKeysStream {
		return nil, NewServiceErrorf(http.StatusBadRequest, "Key stream with id `%s` has wrong type", keyStreamId)
	}

	return &platformKeyProvider{db: database, streamId: keyStreamId}, nil
}

var letterRunes = []rune("1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ")

//RandStringRunes creates random string with specified length
func RandStringRunes(n int) string {
	b := make([]rune, n)
	lenRunes := len(letterRunes)
	for i := range b {
		b[i] = letterRunes[rand.Intn(lenRunes)]
	}
	return string(b)
}
