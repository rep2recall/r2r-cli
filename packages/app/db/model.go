package db

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type Model struct {
	ID        string `gorm:"primarykey;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time      `gorm:"index"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Name      string `gorm:"index"`
	Front     string
	Back      string
	Shared    string
	Generated JSONObject
}

type JSONObject map[string]interface{}

// Scan scan value into JSON, implements sql.Scanner interface
func (j *JSONObject) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	s, ok := value.(string)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	r := make(map[string]interface{})
	err := json.Unmarshal([]byte(s), &r)
	*j = r
	return err
}

// Value return json value, implement driver.Value interface
func (j JSONObject) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}

	b, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// GormDBDataType represents driver's JSON data type
func (JSONObject) GormDBDataType(db *gorm.DB, _ *schema.Field) string {
	switch db.Dialector.Name() {
	case "mysql", "sqlite":
		return "JSON"
	case "postgres":
		return "JSONB"
	}
	return "TEXT"
}

// GormDataType gorm common data type
func (JSONObject) GormDataType() string {
	return "jsonObject"
}

func (Model) Tidy(tx *gorm.DB) error {
	return nil
}
