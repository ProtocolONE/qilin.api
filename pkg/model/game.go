package model

import (
    "github.com/lib/pq"
    "github.com/satori/go.uuid"
    "qilin-api/pkg/model/game"
    "qilin-api/pkg/model/utils"
    "time"
)

type (
    GameTag struct {
        ID                   string                 `gorm:"primary_key"`
        Title                utils.LocalizedString  `gorm:"column:title; type:JSONB NOT NULL"`
    }

    GameGenre struct {
        GameTag
    }

    Game struct {
        ID                   uuid.UUID             `gorm:"type:uuid; primary_key"`
        InternalName         string                `gorm:"index: UNIQ"`
        Title                string                `gorm:"type:text"`
        Developers           string                `gorm:"type:text"`
        Publishers           string                `gorm:"type:text"`
        ReleaseDate          time.Time             `gorm:"type:timestamp; default:now()"`
        DisplayRemainingTime bool                  `gorm:"type:boolean"`
        AchievementOnProd    bool                  `gorm:"type:boolean"`
        FeaturesCommon       pq.StringArray        `gorm:"type:text[]"`
        FeaturesCtrl         string                `gorm:"type:text"`
        Platforms            game.Platforms        `gorm:"type:JSONB NOT NULL"`
        Requirements         game.GameRequirements `gorm:"type:JSONB NOT NULL"`
        Languages            game.GameLangs        `gorm:"type:JSONB NOT NULL"`
        Genre                pq.StringArray        `gorm:"type:text[]"`
        Tags                 pq.StringArray        `gorm:"type:text[]"`
        Vendor               *User                 `gorm:"foreignkey:UserId; association_foreignkey:Refer"`
        VendorId             uuid.UUID             `gorm:"type:uuid"`

        CreatedAt time.Time
        UpdatedAt time.Time
        DeletedAt *time.Time `sql:"index"`
    }

    // GameService is a helper service class to interact with Game object.
    GameService interface {
        GetTags([]string) ([]GameTag, error)
        GetGenres([]string) ([]GameGenre, error)
        Create(*uuid.UUID, string) (*Game, error)
        GetList(vendorId *uuid.UUID, offset, limit int, internalName, genre, releaseDate, sort string, price float64) ([]*Game, error)
        GetInfo(vendorId *uuid.UUID, gameId *uuid.UUID) (*Game, error)
        Delete(vendorId *uuid.UUID, gameId *uuid.UUID) error
        Update(vendorId *uuid.UUID, game *Game) error
    }
)