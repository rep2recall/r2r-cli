package db

import (
	"database/sql"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/mattn/go-sqlite3"
	"github.com/rep2recall/r2r/shared"
	gormSqlite "gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func tokenize(s string, lang string) string {
	segmenter := shared.Config.Segmenter[lang]
	if len(segmenter.Command) > 0 {
		args := segmenter.Command[1:]
		args = append(args, s)

		dir := filepath.Join(shared.UserDataDir, "plugins", "app")
		cmd := exec.Command(filepath.Join(dir, segmenter.Command[0]), args...)
		cmd.Dir = dir
		cmd.Stderr = os.Stderr

		if b, e := cmd.Output(); e == nil {
			return string(b)
		}
	}

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
		shared.Fatalln(err)
	}

	if err := db.AutoMigrate(
		&Model{},
		&Template{},
		&Note{},
		&NoteAttr{},
		&Card{},
	); err != nil {
		shared.Fatalln(err)
	}

	if err := NoteFTSInit(db); err != nil {
		shared.Fatalln(err)
	}

	return db
}
