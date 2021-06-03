package db

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// Note is an FTS5 model
type Note struct {
	ID        string
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt

	ModelID string
	Model   Model

	Key       string
	Data      NoteData
	Generated bool
}

type NoteData struct {
	Raw string
}

func (j *NoteData) Scan(value interface{}) error {
	s, ok := value.(string)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	j.Raw = s
	return nil
}

func (j NoteData) Value() (driver.Value, error) {
	return j.Raw, nil
}

func (j NoteData) Get() (interface{}, error) {
	b := []byte(j.Raw)
	if regexp.MustCompile("^{.+}$").Match(b) {
		var out interface{}
		e := json.Unmarshal(b, &out)
		return out, e
	}

	return j.Raw, nil
}

func (j *NoteData) Set(v interface{}) error {
	switch v := v.(type) {
	case string:
		j.Raw = v
	default:
		b, e := json.Marshal(v)
		j.Raw = string(b)
		return e
	}

	return nil
}

func (NoteData) GormDBDataType(db *gorm.DB, _ *schema.Field) string {
	switch db.Dialector.Name() {
	case "mysql", "sqlite":
		return "JSON"
	case "postgres":
		return "JSONB"
	}
	return "TEXT"
}

func (Note) Init(tx *gorm.DB) error {
	r := tx.Exec(`
	CREATE VIRTUAL TABLE IF NOT EXISTS note USING fts5(
		id,        		-- TEXT NOT NULL
		updated_at,		-- TIMESTAMP
		deleted_at,    	-- TIMESTAMP
		model_id,		-- TEXT NOT NULL REFERENCES model(id)
		"key",        	-- TEXT NOT NULL
		"data",       	-- JSON -- must be JSONified text
		"generated" UNINDEXED
	);
	`)
	if r.Error != nil {
		return r.Error
	}

	return nil
}

func (Note) Tidy(tx *gorm.DB) error {
	if r := tx.
		Where("id IS NULL").
		Or("[key] IS NULL").
		Or("[data] LIKE '{%}' AND NOT json_valid([data])").
		Or("model_id NOT IN (SELECT id FROM model)").
		Or("ROWID NOT IN (SELECT ROWID FROM note GROUP BY id, [key])").
		Or("ROWID NOT IN (SELECT ROWID FROM note GROUP BY model_id)").
		Delete(&Note{}); r.Error != nil {
		return r.Error
	}

	return nil
}
