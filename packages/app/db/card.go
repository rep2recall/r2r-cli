package db

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
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

	Front       string
	Back        string
	Shared      string
	Mnemonic    string
	SRSLevel    int            `gorm:"index"`
	NextReview  *time.Time     `gorm:"index"`
	LastRight   *time.Time     `gorm:"index"`
	LastWrong   *time.Time     `gorm:"index"`
	MaxRight    int            `gorm:"index"`
	MaxWrong    int            `gorm:"index"`
	RightStreak int            `gorm:"index"`
	WrongStreak int            `gorm:"index"`
	Tag         SpaceSeparated `gorm:"index"`
	Status      SpaceSeparated `gorm:"index;default:' new '"`
}

func (c *Card) BeforeSave(tx *gorm.DB) error {
	status := map[string]bool{}
	now := time.Now()

	if c.NextReview == nil {
		status["new"] = true
	} else if c.NextReview.Before(now) {
		status["due"] = true
	}

	if c.WrongStreak > 1 {
		status["leech"] = true
	}

	if c.SRSLevel > 3 {
		status["graduated"] = true
	} else if c.NextReview != nil {
		status["learning"] = true
	}

	return c.Status.Set(status)
}

type SpaceSeparated struct {
	Raw string
}

func (j *SpaceSeparated) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	s, ok := value.(string)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal SpaceSeparated value:", value))
	}

	j.Raw = s
	return nil
}

func (j SpaceSeparated) Value() (driver.Value, error) {
	return j.Raw, nil
}

func (j SpaceSeparated) Get() (map[string]bool, error) {
	if j.Raw == "" {
		return map[string]bool{}, nil
	}

	if len(j.Raw) < 2 || (j.Raw[0] != ' ' && j.Raw[len(j.Raw)-1] != ' ') {
		return nil, errors.New(fmt.Sprint("Failed to unmarshal SpaceSeparated value:", j.Raw))
	}

	out := map[string]bool{}
	for _, s := range strings.Split(j.Raw[1:len(j.Raw)-1], " ") {
		out[s] = true
	}

	return out, nil
}

func (j *SpaceSeparated) Set(v map[string]bool) error {
	if len(v) == 0 {
		j.Raw = ""
		return nil
	}

	out := " "
	for k, v := range v {
		if v {
			out += k + " "
		}
	}
	j.Raw = out
	return nil
}

func (c Card) Data(tx *gorm.DB) (map[string]interface{}, error) {
	var notes []Note
	r := tx.Where("id = ?", c.NoteID).Find(&notes)
	if r.Error != nil {
		return nil, r.Error
	}

	out := map[string]interface{}{}

	for _, n := range notes {
		v, e := n.Data.Get()
		if e != nil {
			return nil, e
		}

		out[n.Key] = v
	}

	return out, nil
}

func (Card) Tidy(tx *gorm.DB) error {
	if r := tx.
		Where("template_id IS NOT NULL AND note_id IS NULL").
		Or("template_id IS NULL AND note_id IS NOT NULL").
		Or("template_id IS NOT NULL AND note_id IS NOT NULL AND ROWID NOT IN (SELECT ROWID FROM [card] GROUP BY template_id, note_id)").
		Or("template_id IS NULL AND note_id IS NULL AND front IS NULL").
		Delete(&Card{}); r.Error != nil {
		return r.Error
	}

	return nil
}

var srsMap []time.Duration = []time.Duration{
	4 * time.Hour,
	8 * time.Hour,
	24 * time.Hour,
	3 * 24 * time.Hour,
	7 * 24 * time.Hour,
	2 * 7 * 24 * time.Hour,
	4 * 7 * 24 * time.Hour,
	16 * 7 * 24 * time.Hour,
}

func getNextReview(srsLevel int) time.Time {
	if srsLevel >= 0 && srsLevel < len(srsMap) {
		return time.Now().Add(srsMap[srsLevel])
	}

	return time.Now().Add(1 * time.Hour)
}

// UpdateSRSLevel updates SRSLevel and also updates stats
func (c Card) UpdateSRSLevel(tx *gorm.DB, dSRSLevel int) error {
	now := time.Now()
	q := Card{
		ID:          c.ID,
		LastRight:   c.LastRight,
		LastWrong:   c.LastWrong,
		RightStreak: c.RightStreak,
		WrongStreak: c.WrongStreak,
		MaxRight:    c.MaxRight,
		MaxWrong:    c.MaxWrong,
		SRSLevel:    c.SRSLevel,
	}

	if dSRSLevel > 0 {
		q.LastRight = &now
		q.RightStreak++

		if q.MaxRight < q.RightStreak {
			q.MaxRight = q.RightStreak
		}
	} else if dSRSLevel < 0 {
		q.LastWrong = &now
		q.WrongStreak++

		if q.MaxWrong < q.WrongStreak {
			q.MaxWrong = q.WrongStreak
		}
	}

	q.SRSLevel += dSRSLevel

	if q.SRSLevel >= len(srsMap) {
		q.SRSLevel = len(srsMap) - 1
	}

	if q.SRSLevel < 0 {
		q.SRSLevel = 0
		nextReview := getNextReview(-1)
		q.NextReview = &nextReview
	} else {
		nextReview := getNextReview(q.SRSLevel)
		q.NextReview = &nextReview
	}

	r := tx.Updates(&q)
	return r.Error
}
