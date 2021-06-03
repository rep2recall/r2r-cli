package db

import (
	"database/sql"
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
	CreatedAt TimeString
	UpdatedAt TimeString
	DeletedAt TimeString

	ModelID string

	Key  string
	Data NoteData
}

type TimeString sql.NullTime

func (j *TimeString) Scan(value interface{}) error {
	s, ok := value.(string)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal timestamp value:", value))
	}

	if s == "" {
		return nil
	}

	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}

	j.Time = t
	return nil
}

func (j TimeString) Value() (driver.Value, error) {
	if j.Time.IsZero() {
		return nil, nil
	}

	b, e := j.Time.MarshalText()
	if e != nil {
		return nil, e
	}

	return string(b), nil
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
		created_at,		-- TIMESTAMP
		updated_at,		-- TIMESTAMP
		deleted_at,    	-- TIMESTAMP
		model_id,		-- TEXT NOT NULL REFERENCES model(id)
		"key",        	-- TEXT NOT NULL
		"data",       	-- JSON or TEXT
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
