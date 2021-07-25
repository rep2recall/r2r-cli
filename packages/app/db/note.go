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

type Note struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time      `gorm:"index"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Key       string         `gorm:"index:,unique"`
	ModelID   string         `gorm:"index"`
	Attrs     []NoteAttr     `gorm:"constraint:OnDelete:CASCADE"`
}

// NoteAttr contains hooks to FTS5 model
type NoteAttr struct {
	ID        uint
	CreatedAt time.Time
	UpdatedAt time.Time `gorm:"index"`
	NoteID    string    `gorm:"index:idx_note_attr_u,unique"`
	Key       string    `gorm:"index:idx_note_attr_u,unique"`
	Data      NoteData
	Lang      string
}

type NoteData struct {
	Raw string
}

func (j *NoteData) Scan(value interface{}) error {
	s, ok := value.(string)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSON value:", value))
	}

	j.Raw = s
	return nil
}

func (j NoteData) Value() (driver.Value, error) {
	return j.Raw, nil
}

func (j NoteData) Get() (interface{}, error) {
	b := []byte(j.Raw)
	if regexp.MustCompile(`^({.*}|\[.*\])$`).Match(b) {
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
	return "JSON"
}

// GormDataType gorm common data type
func (NoteData) GormDataType() string {
	return "NoteData"
}

func NoteFTSInit(tx *gorm.DB) error {
	r := tx.Exec(`
	CREATE VIRTUAL TABLE IF NOT EXISTS note_fts USING fts5(
		note_id UNINDEXED,
		key,
		data,
		content=note_attr,
		content_rowid=id,
		tokenize=porter
	);

	-- Triggers to keep the FTS index up to date.
	CREATE TRIGGER IF NOT EXISTS t_note_attr_ai AFTER INSERT ON note_attr BEGIN
		INSERT INTO note_fts(rowid, note_id, key, data) VALUES (new.id, new.note_id, new.key, tokenize(new.data, new.lang));
	END;
	CREATE TRIGGER IF NOT EXISTS t_note_attr_ad AFTER DELETE ON note_attr BEGIN
		INSERT INTO note_fts(note_fts, rowid, note_id, key, data) VALUES ('delete', old.id, old.note_id, old.key, tokenize(old.data, old.lang));
	END;
	CREATE TRIGGER IF NOT EXISTS t_note_attr_au AFTER UPDATE ON note_attr BEGIN
		INSERT INTO note_fts(note_fts, rowid, note_id, key, data) VALUES ('delete', old.id, old.note_id, old.key, tokenize(old.data, old.lang));
		INSERT INTO note_fts(rowid, note_id, key, data) VALUES (new.id, new.note_id, new.key, tokenize(new.data, new.lang));
	END;
	`)
	if r.Error != nil {
		return r.Error
	}

	return nil
}

func (Note) Tidy(tx *gorm.DB) error {
	return nil
}
