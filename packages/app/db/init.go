package db

import (
	"log"
	"path/filepath"

	"github.com/rep2recall/rep2recall/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func Connect() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(filepath.Join(shared.UserDataDir(), "data.db")), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		log.Fatalln(err)
	}

	if err := db.AutoMigrate(
		Model{},
		Template{},
		Card{},
	); err != nil {
		log.Fatalln(err)
	}

	if err := (Note{}).Init(db); err != nil {
		log.Fatalln(err)
	}

	return db
}
