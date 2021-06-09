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

	Front       string
	Back        string
	Shared      string
	Mnemonic    string
	SRSLevel    int        `gorm:"index"`
	NextReview  *time.Time `gorm:"index"`
	LastRight   *time.Time `gorm:"index"`
	LastWrong   *time.Time `gorm:"index"`
	MaxRight    int        `gorm:"index"`
	MaxWrong    int        `gorm:"index"`
	RightStreak int        `gorm:"index"`
	WrongStreak int        `gorm:"index"`
	Tag         string     `gorm:"index"`
	Status      string     `gorm:"index;->;type:TEXT AS (' '||IIF(next_review IS NULL, 'new', IIF(strftime('%s', next_review) < strftime('%s', 'now'), 'due', ''))||' '||IIF(wrong_streak > 1, 'leech', '')||' '||IIF(srs_level > 3, 'graduated', 'learning')||' ')"`
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

	r := tx.Updates(&c)
	return r.Error
}
