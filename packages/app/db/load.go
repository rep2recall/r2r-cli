package db

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/go-playground/validator"
	"github.com/rep2recall/rep2recall/browser"
	"github.com/rep2recall/rep2recall/shared"
	"gopkg.in/yaml.v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var validate *validator.Validate

type LoadFile struct {
	Model []struct {
		ID        string `validate:"required,uuid"`
		Name      string
		Front     string
		Back      string
		Shared    string
		Generated map[string]interface{} `validate:"blank-is-string"`
	}
	Template []struct {
		ID      string `validate:"required,uuid"`
		ModelID string `validate:"required,uuid"`
		Name    string
		Front   string
		Back    string
		Shared  string
	}
	Note []struct {
		ID      string                 `validate:"required,uuid"`
		ModelID string                 `validate:"required,uuid"`
		Data    map[string]interface{} `validate:"required"`
	}
	Card []struct {
		ID         string `validate:"required,uuid"`
		TemplateID string `validate:"required,uuid"`
		NoteID     string `validate:"required,uuid"`
		Tag        string
		Front      string
		Back       string
		Shared     string
		Mnemonic   string
	}
}

func ValidateBlankIsString(fl validator.FieldLevel) bool {
	bl := fl.Field().MapIndex(reflect.ValueOf("_"))
	if !bl.IsNil() {
		return bl.Type().Name() == "string"
	}

	return true
}

func Load(tx *gorm.DB, f string) error {
	b, e := ioutil.ReadFile(filepath.Join(shared.UserDataDir(), f))
	if e != nil {
		return e
	}

	var loadFile LoadFile
	if e := yaml.Unmarshal(b, &loadFile); e != nil {
		return e
	}

	validate = validator.New()
	validate.RegisterValidation("blank-is-string", ValidateBlankIsString)

	if e := validate.Struct(loadFile); e != nil {
		return e
	}

	modelGenMap := make(map[string]map[string]interface{})

	for _, m := range loadFile.Model {
		if m.Generated != nil {
			modelGenMap[m.ID] = m.Generated
		}

		if r := tx.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(Model{
			ID:        m.ID,
			Name:      m.Name,
			Front:     m.Front,
			Back:      m.Back,
			Shared:    m.Shared,
			Generated: m.Generated,
		}); r.Error != nil {
			return r.Error
		}
	}

	for _, t := range loadFile.Template {
		if r := tx.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(Template{
			ID:      t.ID,
			ModelID: t.ModelID,
			Name:    t.Name,
			Front:   t.Front,
			Back:    t.Back,
			Shared:  t.Shared,
		}); r.Error != nil {
			return r.Error
		}
	}

	var toGenerate []*browser.EvalContext
	for _, n := range loadFile.Note {
		if n.ModelID != "" && modelGenMap[n.ModelID] == nil {
			var m Model
			if r := tx.Where("id = ?", n.ModelID).First(&m); r.Error != nil {
				return r.Error
			}
			if m.Generated != nil {
				modelGenMap[n.ModelID] = m.Generated
			}

			if modelGenMap[n.ModelID] != nil && modelGenMap[n.ModelID]["_"] != nil {
				jsb, e := json.Marshal(modelGenMap[n.ModelID]["_"].(string))
				if e != nil {
					return e
				}
				datab, e := json.Marshal(n.Data)
				if e != nil {
					return e
				}
				idb, e := json.Marshal(n.ID)
				if e != nil {
					return e
				}

				toGenerate = append(toGenerate, &browser.EvalContext{
					JS: fmt.Sprintf(
						`(async function() {
							const data = %s;
							await eta.renderAsync(%s, data);
							return {
								id: %s,
								data
							};
						})();`,
						string(datab),
						string(jsb),
						string(idb),
					),
				})
			}
		}
	}

	noteGenResultMap := make(map[string]map[string]interface{})

	if len(toGenerate) > 0 {
		plugins := []string{}

		e = filepath.Walk(filepath.Join(shared.UserDataDir(), "plugins"), func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if strings.HasSuffix(path, ".js") {
				plugins = append(plugins, filepath.ToSlash(path))
			}

			return nil
		})
		if e != nil {
			return e
		}

		b := browser.Browser{}
		b.Eval([]string{
			"https://cdn.jsdelivr.net/npm/eta/dist/browser/eta.min.js",
		}, toGenerate...)
		for _, g := range toGenerate {
			out := g.Output.(map[string]interface{})
			noteGenResultMap[out["id"].(string)] = out["data"].(map[string]interface{})
		}
	}

	for _, n := range loadFile.Note {
		if noteGenResultMap[n.ID] != nil {
			for key, v := range noteGenResultMap[n.ID] {
				if n.Data[key] == nil {
					n.Data[key] = v
				}
			}
		}

		for key, v := range n.Data {
			data := NoteData{}
			if err := data.Set(v); err != nil {
				return err
			}

			if r := tx.Where(Note{
				ID:      n.ID,
				ModelID: n.ModelID,
				Key:     key,
			}).UpdateColumn("id", "_"+n.ID); r.Error != nil {
				return r.Error
			}

			if r := tx.Create(Note{
				ID:      n.ID,
				ModelID: n.ModelID,
				Key:     key,
				Data:    data,
			}); r.Error != nil {
				return r.Error
			}
		}
	}

	for _, c := range loadFile.Card {
		if r := tx.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(Card{
			ID:         c.ID,
			TemplateID: c.TemplateID,
			NoteID:     c.NoteID,
			Tag:        c.Tag,
			Front:      c.Front,
			Back:       c.Back,
			Shared:     c.Shared,
			Mnemonic:   c.Mnemonic,
		}); r.Error != nil {
			return r.Error
		}
	}

	return nil
}
