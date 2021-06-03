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
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		log.Fatalln(err)
	}

	mi := db.Migrator()
	if !mi.HasTable(Model{}) {
		if e := mi.CreateTable(Model{}); e != nil {
			log.Fatalln(e)
		}
	}
	if !mi.HasTable(Card{}) {
		if e := mi.CreateTable(Card{}); e != nil {
			log.Fatalln(e)
		}
	}
	if !mi.HasTable(Template{}) {
		if e := mi.CreateTable(Template{}); e != nil {
			log.Fatalln(e)
		}
	}

	if err := (Note{}).Init(db); err != nil {
		log.Fatalln(err)
	}

	return db
}
