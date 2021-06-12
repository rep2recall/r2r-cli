package db

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

type qStruct struct {
	Sign  string
	Key   string
	Op    string
	Value string
}

func qSearch(q string) ([]qStruct, error) {
	out := make([]qStruct, 0)

	if strings.TrimSpace(q) == "" {
		return out, nil
	}

	quoted := ""
	pos := "sign"
	current := qStruct{}

	for ci, c := range q {
		err := func() error {
			if c == '"' && (ci == 0 || q[ci-1] != '\\') {
				if len(quoted) > 0 && quoted[0] == '"' {
					if strings.ContainsRune(quoted[1:], '"') {
						return fmt.Errorf("overquoted: %s", q[:ci+1])
					}

					quoted += string(c)
					switch pos {
					case "sign", "key":
						current.Key = quoted
						quoted = ""
						pos = "op"
					default:
						current.Value = quoted
						quoted = ""
						out = append(out, current)
						current = qStruct{}
						pos = "sign"
					}

					return nil
				}
				quoted += string(c)
				return nil
			} else if quoted != "" {
				quoted += string(c)
				return nil
			}

			if c == ' ' {
				if quoted == "" {
					if current.Key != "" {
						out = append(out, current)
					}
					current = qStruct{}
					pos = "sign"
				}
			} else {
				if pos == "sign" {
					if c == '?' || c == '-' {
						pos = "key"
						current.Sign = string(c)
						return nil
					}

					pos = "key"
				}

				if pos == "key" {
					if c == '>' || c == '<' || c == ':' || c == '=' {
						pos = "op"
						current.Op = string(c)
						return nil
					}
					current.Key += string(c)
				}

				if pos == "op" {
					if len(current.Op) == 1 && c == '=' {
						current.Op += string(c)
						pos = "value"
						return nil
					}
					pos = "value"
				}

				if pos == "value" {
					current.Value += string(c)
				}
			}

			return nil
		}()

		if err != nil {
			return nil, err
		}
	}

	if current.Key != "" {
		out = append(out, current)
	}

	return out, nil
}

func dequote(q string) string {
	if len(q) >= 2 && q[0] == '"' && q[len(q)-1] == '"' {
		return q[1 : len(q)-2]
	}
	return q
}

func Search(tx *gorm.DB, q string) *gorm.DB {
	if strings.TrimSpace(q) == "" {
		return tx.Where("TRUE")
	}

	rootTx := tx
	includes := struct {
		Model    bool
		Template bool
	}{}

	makeNumber := func(tx *gorm.DB, str qStruct) *gorm.DB {
		switch str.Key {
		case "srsLevel":
			str.Key = "card.srs_level"
		case "maxRight":
			str.Key = "card.max_right"
		case "maxWrong":
			str.Key = "card.max_wrong"
		case "rightStreak":
			str.Key = "card.right_streak"
		case "wrongStreak":
			str.Key = "card.wrong_streak"
		}

		if str.Value == "NULL" {
			return tx.Where(fmt.Sprintf("%s IS NULL", str.Key))
		}

		if str.Op == ":" {
			str.Op = "="
		}

		v, e := strconv.Atoi(str.Value)
		if e != nil {
			return tx.Where("FALSE")
		}

		return tx.Where(fmt.Sprintf("%s %s ?", str.Key, str.Op), v)
	}

	makeDate := func(tx *gorm.DB, str qStruct) *gorm.DB {
		switch str.Key {
		case "nextReview":
			str.Key = "card.next_review"
		case "lastRight":
			str.Key = "card.last_right"
		case "lastWrong":
			str.Key = "card.last_wrong"
		case "createdAt":
			str.Key = "card.created_at"
		case "updatedAt":
			str.Key = "card.updated_at"
		}

		if str.Value == "NULL" {
			return tx.Where(fmt.Sprintf("%s IS NULL", str.Key))
		}

		m := regexp.MustCompile(`^([+-]?)(\d+)(min|h|d|w)$`).FindStringSubmatch(str.Value)
		if len(m) == 4 {
			var n time.Duration = -1
			if m[1] == "+" {
				n = 1
			}
			n = n * time.Duration(func() int {
				p, _ := strconv.Atoi(m[2])
				return p
			}())

			unit := time.Hour
			switch m[3] {
			case "min":
				unit = time.Minute
			case "d":
				unit = time.Hour * 24
			case "w":
				unit = time.Hour * 24 * 7
			}

			time0 := time.Now().Add(n * unit)
			time1 := time0

			switch str.Op {
			case ":", "=":
				time0 = time0.Add(-time.Duration(unit.Nanoseconds() / 2))
				time1 = time1.Add(time.Duration(unit.Nanoseconds() / 2))

				return tx.
					Where(fmt.Sprintf("strftime('%%s',%s) > strftime('%%s',?)", str.Key), time0.String()).
					Where(fmt.Sprintf("strftime('%%s',%s) < strftime('%%s',?)", str.Key), time1.String())
			}

			return tx.Where(fmt.Sprintf("strftime('%%s',%s) %s strftime('%%s',?)", str.Key, str.Op), time0.String())
		}

		return tx.Where("FALSE")
	}

	makeClause := func(tx *gorm.DB, str qStruct) *gorm.DB {
		if str.Value == "" {
			str.Value = str.Key
			str.Key = ""
		}

		switch str.Key {
		case "srsLevel", "maxRight", "maxWrong", "rightStreak", "wrongStreak":
			return makeNumber(tx, str)
		case "nextReview", "lastRight", "lastWrong", "createdAt", "updatedAt":
			return makeDate(tx, str)
		}

		value := dequote(str.Value)

		switch str.Key {
		case "tag":
			return tx.Where("card.tag LIKE '% '||?||' %'", value)
		case "status":
			return tx.Where("card.status LIKE '% '||?||' %'", value)
		case "id":
			return tx.Where("card.id = ?", value)
		case "noteId":
			return tx.Where("card.note_id = ?", value)
		case "templateId":
			return tx.Where("card.template_id = ?", value)
		case "modelId":
			includes.Template = true
			return tx.Where("template.model_id = ?", value)
		case "template":
			includes.Template = true
			if str.Op == "=" {
				return tx.Where("template.name = ?", value)
			} else {
				return tx.Where("template.name LIKE '%'||?||'%'", value)
			}
		case "model":
			includes.Template = true
			includes.Model = true
			if str.Op == "=" {
				return tx.Where("model.name = ?", value)
			} else {
				return tx.Where("model.name LIKE '%'||?||'%'", value)
			}
		}

		key := dequote(str.Key)

		if key != "" {
			return tx.Where(`card.note_id IN (
				SELECT note_id FROM note_fts WHERE "key" = ? AND note_fts MATCH 'data:"'||?||'"'
			)`, key, strings.ReplaceAll(value, `"`, `""`))
		}

		return tx.Where(`card.note_id IN (
			SELECT note_id FROM note_fts WHERE note_fts MATCH 'data:"'||?||'"'
		)`, strings.ReplaceAll(value, `"`, `""`))
	}

	arr, err := qSearch(q)
	if err != nil {
		return tx.Where("FALSE")
	}

	var andCond *gorm.DB
	var orCond *gorm.DB
	var notCond *gorm.DB

	for _, r := range arr {
		switch r.Sign {
		case "?":
			if orCond == nil {
				orCond = makeClause(tx, r)
			} else {
				orCond = orCond.Or(makeClause(tx, r))
			}
		case "-":
			if notCond == nil {
				notCond = makeClause(tx, r)
			} else {
				notCond = notCond.Where(makeClause(tx, r))
			}
		default:
			if andCond == nil {
				andCond = makeClause(tx, r)
			} else {
				andCond = andCond.Where(makeClause(tx, r))
			}
		}
	}

	if andCond == nil {
		andCond = tx.Where("TRUE")
	}

	if notCond != nil {
		andCond = andCond.Not(notCond)
	}

	if includes.Template {
		tx = tx.Joins("Template")
	}

	if includes.Model {
		tx = tx.Joins("Model")
	}

	if orCond != nil {
		tx = tx.Where(orCond.Or(andCond))
	} else {
		tx = tx.Where(andCond)
	}

	return rootTx.Where(tx)
}
