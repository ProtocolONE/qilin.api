package orm

import (
	"flag"
	"qilin-api/pkg/conf"
	"qilin-api/pkg/model"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type Database struct {
	database *gorm.DB
}

func NewDatabase(config *conf.Database) (*Database, error) {
	db, err := gorm.Open("postgres", config.DSN)
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
		&model.DocumentsInfo{},
		&model.Notification{},
	)
}

//DropAllTables is method for clearing DB. WARNING: Use it only for testing purposes
func (db *Database) DropAllTables() error {
	if flag.Lookup("test.v") != nil {
		return db.database.DropTableIfExists(
			model.GameRating{},
			model.Descriptor{},
			model.GameDescr{},
			model.GameGenre{},
			model.GameTag{},
			model.Discount{},
			model.Price{},
			model.Game{},
			model.Vendor{},
			model.User{},
			"vendor_users",
			model.DocumentsInfo{},
			model.Notification{},
		).Error
	}
	return nil
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
