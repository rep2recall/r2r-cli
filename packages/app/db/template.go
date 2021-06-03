package db

import (
	"time"

	"gorm.io/gorm"
)

type Template struct {
	ID        string `gorm:"primarykey;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time      `gorm:"index"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	ModelID string `gorm:"index"`
	Model   Model  `gorm:"constraint:OnDelete:CASCADE"`

	Name   string `gorm:"index"`
	Front  string
	Back   string
	Shared string
}

func (Template) Tidy(tx *gorm.DB) error {
	if r := tx.
		Where(`([name] IS NOT NULL AND ROWID NOT IN (SELECT ROWID FROM template WHERE [name] IS NOT NULL GROUP BY model_id, [name]))`).
		Or(`(front IS NULL AND model_id IN (SELECT id FROM model WHERE front IS NULL))`).
		Delete(&Template{}); r.Error != nil {
		return r.Error
	}

	return nil
}
