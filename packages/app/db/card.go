package db

import (
	"time"

	"gorm.io/gorm"
)

type Card struct {
	ID        string `gorm:"primarykey;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time      `gorm:"index"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	TemplateID string   `gorm:"index"`
	Template   Template `gorm:"constraint:OnDelete:CASCADE"`

	NoteID string `gorm:"index"`
	Note   Note   `gorm:"constraint:OnDelete:CASCADE"`

	Tag         string `gorm:"index"`
	Front       string
	Back        string
	Shared      string
	SRSLevel    int       `gorm:"index"`
	NextReview  time.Time `gorm:"index"`
	LastRight   time.Time `gorm:"index"`
	LastWrong   time.Time `gorm:"index"`
	MaxRight    int       `gorm:"index"`
	MaxWrong    int       `gorm:"index"`
	RightStreak int       `gorm:"index"`
	WrongStreak int       `gorm:"index"`
}

func (Card) Tidy(tx *gorm.DB) error {
	if r := tx.
		Where("template_id IS NOT NULL AND note_id IS NULL").
		Or("template_id IS NULL AND note_id IS NOT NULL").
		Or("template_id IS NOT NULL AND note_id IS NOT NULL AND ROWID NOT IN (SELECT ROWID FROM [card] GROUP BY template_id, note_id)").
		Or("template_id IS NULL AND note_id IS NULL AND front IS NULL").
		Delete(Card{}); r.Error != nil {
		return r.Error
	}

	return nil
}
