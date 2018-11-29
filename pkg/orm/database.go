package orm

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"qilin-api/pkg/conf"
	"qilin-api/pkg/model"
)

type Database struct {
	database *gorm.DB
}

func NewDatabase(config *conf.Database) (*Database, error) {
	dsl := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		config.Host, config.Port, config.Database, config.User, config.Password)

	db, err := gorm.Open("postgres", dsl)
	if err != nil {
		return nil, err
	}

	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)

	return &Database{db}, err
}

func (db *Database) Init() {
	db.database.AutoMigrate(&model.Game{}, &model.Prices{})
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
