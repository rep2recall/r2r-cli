package db

import (
	"database/sql"
	"log"
	"path/filepath"

	"github.com/mattn/go-sqlite3"
	"github.com/rep2recall/r2r-cli/shared"
	gormSqlite "gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// TODO: check language and parse accordingly
func tokenize(s string, lang string) string {
	return s
}

func init() {
	sql.Register("sqlite3_custom", &sqlite3.SQLiteDriver{
		ConnectHook: func(conn *sqlite3.SQLiteConn) error {
			if err := conn.RegisterFunc("tokenize", tokenize, true); err != nil {
				return err
			}
			return nil
		},
	})
}

func Connect() *gorm.DB {
	db, err := gorm.Open(gormSqlite.Dialector{
		DriverName: "sqlite3_custom",
		DSN:        filepath.Join(shared.UserDataDir, shared.Config.DB),
	}, &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		log.Fatalln(err)
	}

	if err := db.AutoMigrate(
		&Model{},
		&Template{},
		&Note{},
		&NoteAttr{},
		&Card{},
	); err != nil {
		log.Fatalln(err)
	}

	if err := NoteFTSInit(db); err != nil {
		log.Fatalln(err)
	}

	return db
}
