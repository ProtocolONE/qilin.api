package orm

import (
	"fmt"
	"qilin-api/pkg/conf"
	"qilin-api/pkg/model"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type Database struct {
	database *gorm.DB
}

func NewDatabase(config *conf.Database) (*Database, error) {
	dsl := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.User, config.Password, config.Host, config.Port, config.Database)

	db, err := gorm.Open("postgres", dsl)
	if err != nil {
		return nil, err
	}

	db.LogMode(config.LogMode)
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)

	return &Database{db}, err
}

func (db *Database) Init() {
	db.database.AutoMigrate(
		&model.User{},
		&model.Vendor{},
		&model.Game{},
		&model.BasePrice{},
		&model.Price{},
		&model.Media{},
		&model.Discount{},
		&model.GameTag{},
		&model.GameGenre{},
		&model.GameDescr{},
		&model.Descriptor{},
		&model.GameRating{},
	)
}

func (db *Database) DB() *gorm.DB {
	return db.database
}

func (db *Database) Close() error {
	if db.database == nil {
		return nil
	}

	return db.database.Close()
}
