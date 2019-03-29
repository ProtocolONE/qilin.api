package utils_test

import (
	"github.com/satori/go.uuid"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
	"qilin-api/pkg/orm/utils"
	"qilin-api/pkg/test"
	"testing"
)

func TestCheckExists(t *testing.T) {
	config, err := qilin_test.LoadTestConfig()
	if err != nil {
		t.FailNow()
	}
	db, err := orm.NewDatabase(&config.Database)
	if err != nil {
		t.FailNow()
	}

	if err := db.DropAllTables(); err != nil {
		t.Error(err)
	}
	if err := db.Init(); err != nil {
		t.Error(err)
	}

	game := &model.Game{ID: uuid.NewV4(), Tags: []int64{1, 2}, GenreMain: 1, GenreAddition: []int64{2}, FeaturesCommon: []string{"1", "2"}, VendorID: uuid.NewV4(), CreatorID: uuid.NewV4().String(), Title: "asd"}
	err = db.DB().Create(game).Error
	if err != nil {
		t.FailNow()
	}

	tests := []struct {
		name    string
		id      uuid.UUID
		object  interface{}
		want    bool
		wantErr bool
	}{
		{name: "Game no error", object: &model.Game{}, id: game.ID, want: true, wantErr: false},
		{name: "Game error 1", object: &model.Game{}, id: uuid.NewV4(), want: false, wantErr: false},
		{name: "Game error 2", object: &model.Game{}, id: uuid.Nil, want: false, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := utils.CheckExists(db.DB(), tt.object, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CheckExists() = %v, want %v", got, tt.want)
			}
		})
	}

	_ = db.DropAllTables()
}
