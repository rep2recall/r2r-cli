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
	"github.com/google/uuid"
	"github.com/rep2recall/rep2recall/browser"
	"github.com/rep2recall/rep2recall/shared"
	"gopkg.in/yaml.v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var validate *validator.Validate

type LoadedStruct struct {
	Model []struct {
		ID        string `validate:"required,uuid"`
		Name      string
		Front     string
		Back      string
		Shared    string
		Generated map[string]interface{} `validate:"blank-is-string"`
	} `validate:"dive"`
	Template []struct {
		ID      string `validate:"required,uuid"`
		ModelID string `validate:"required,uuid" yaml:"modelId"`
		Name    string
		Front   string
		Back    string
		Shared  string
		If      string
	} `validate:"dive"`
	Note []struct {
		ID      string                 `validate:"required,uuid"`
		ModelID string                 `validate:"required,uuid" yaml:"modelId"`
		Data    map[string]interface{} `validate:"required"`
	} `validate:"dive"`
	Card []struct {
		ID         string `validate:"required,uuid"`
		TemplateID string `validate:"required,uuid" yaml:"templateId"`
		NoteID     string `validate:"required,uuid" yaml:"noteId"`
		Tag        []string
		Front      string
		Back       string
		Shared     string
		Mnemonic   string
	} `validate:"dive"`
}

func ValidateBlankIsString(fl validator.FieldLevel) bool {
	bl := fl.Field().MapIndex(reflect.ValueOf("_"))
	if !bl.IsNil() {
		if bl.Elem().Type().String() != "string" {
			return false
		}
	}

	return true
}

type LoadOptions struct {
	Debug bool
	Port  int
}

func init() {
	validate = validator.New()
	validate.RegisterValidation("blank-is-string", ValidateBlankIsString)
}

func LoadStruct(f string) (LoadedStruct, error) {
	var loadFile LoadedStruct

	b, e := ioutil.ReadFile(filepath.Join(shared.UserDataDir(), f))
	if e != nil {
		return loadFile, e
	}

	if e := yaml.Unmarshal(b, &loadFile); e != nil {
		return loadFile, e
	}

	if e := validate.Struct(&loadFile); e != nil {
		return loadFile, e
	}

	return loadFile, nil
}

func Load(tx *gorm.DB, f string, opts LoadOptions) error {
	b, e := ioutil.ReadFile(filepath.Join(shared.UserDataDir(), f))
	if e != nil {
		return e
	}

	var loadFile LoadedStruct
	if e := yaml.Unmarshal(b, &loadFile); e != nil {
		return e
	}

	if e := validate.Struct(&loadFile); e != nil {
		return e
	}

	modelGenMap := make(map[string]map[string]interface{})

	for _, m := range loadFile.Model {
		if m.Generated != nil {
			modelGenMap[m.ID] = m.Generated
		}

		if r := tx.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&Model{
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
		}).Create(&Template{
			ID:      t.ID,
			ModelID: t.ModelID,
			Name:    t.Name,
			Front:   t.Front,
			Back:    t.Back,
			Shared:  t.Shared,
			If:      t.If,
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
						await Eta.renderAsync(%s, data);
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

	noteGenResultMap := make(map[string]map[string]interface{})
	plugins := []string{}

	e = filepath.Walk(filepath.Join(shared.UserDataDir(), "plugins"), func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, ".js") {
			b, e := ioutil.ReadFile(path)
			if e != nil {
				return e
			}

			plugins = append(plugins, string(b))
		}

		return nil
	})
	if e != nil {
		return e
	}

	if len(toGenerate) > 0 {
		b := browser.Browser{}
		b.Eval(toGenerate, browser.EvalOptions{
			Plugins: plugins,
			Visible: opts.Debug,
			Port:    opts.Port,
		})
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

		if r := tx.FirstOrCreate(&Note{
			ID:      n.ID,
			ModelID: n.ModelID,
		}); r.Error != nil {
			return r.Error
		}

		for key, v := range n.Data {
			data := NoteData{}
			if err := data.Set(v); err != nil {
				return err
			}

			if r := tx.Clauses(clause.OnConflict{
				DoUpdates: clause.AssignmentColumns([]string{"data"}),
			}).Create(&NoteAttr{
				NoteID: n.ID,
				Key:    key,
				Data:   data,
			}); r.Error != nil {
				return r.Error
			}
		}
	}

	templateToCreate := make(map[string]Template)

	for mid := range modelGenMap {
		tids := make([]string, 0)
		for _, t := range loadFile.Template {
			tids = append(tids, t.ID)
		}

		var templates []Template
		if r := tx.Where("model_id = ?", mid).Or("id IN ?", tids).Find(&templates); r.Error != nil {
			return r.Error
		}
		for _, t := range templates {
			templateToCreate[t.ID] = t
		}
	}

	type cardPre struct {
		If       string
		Model    Model
		Template Template
		NoteID   string
		Note     map[string]interface{}
	}
	cardToCompile := make(map[string]cardPre)
	modelMap := make(map[string]Model)
	templateMap := make(map[string]Template)

	for tid, t := range templateToCreate {
		if t.ModelID != "" {
			var notes []Note
			if r := tx.
				Where("model_id = ?", t.ModelID).
				Preload("Attrs").
				Find(&notes); r.Error != nil {
				return r.Error
			}

			model := modelMap[t.ModelID]
			if model.ID == "" {
				if r := tx.Where("id = ?", t.ModelID).First(&model); r.Error != nil {
					return r.Error
				}
				modelMap[t.ModelID] = model
			}

			template := templateMap[tid]
			if template.ID == "" {
				t0 := templateToCreate[tid]
				if t0.ID != "" {
					templateMap[tid] = t0
				}
			}
			if template.ID == "" {
				if r := tx.Where("id = ?", tid).First(&template); r.Error != nil {
					return r.Error
				}
				templateMap[tid] = template
			}

			noteMap := map[string]map[string]interface{}{}

			for _, n := range notes {
				if noteMap[n.ID] == nil {
					noteMap[n.ID] = map[string]interface{}{}
				}

				for _, a := range n.Attrs {
					k := a.Key
					v, e := a.Data.Get()
					if e != nil {
						return e
					}
					noteMap[n.ID][k] = v
				}
			}

			for nid, n := range noteMap {
				cardToCompile[uuid.New().String()] = cardPre{
					If:       t.If,
					NoteID:   nid,
					Note:     n,
					Model:    model,
					Template: template,
				}
			}
		}
	}

	toGenerate = []*browser.EvalContext{}
	for id, ca := range cardToCompile {
		if ca.If != "" {
			jsb, e := json.Marshal(ca.If)
			if e != nil {
				return e
			}
			datab, e := json.Marshal(ca.Note)
			if e != nil {
				return e
			}
			idb, e := json.Marshal(id)
			if e != nil {
				return e
			}

			toGenerate = append(toGenerate, &browser.EvalContext{
				JS: fmt.Sprintf(
					`(async function() {
					const data = %s;
					const base = %s;
					const raw = await Eta.renderAsync(base, data);
					return {
						id: %s,
						rendered: (() => {
							try {
								return !!JSON.parse(raw);
							} catch (e) {}
							return false;
						})(),
					};
				})();`,
					string(datab),
					string(jsb),
					string(idb),
				),
			})
		}
	}

	if len(toGenerate) > 0 {
		b := browser.Browser{}
		b.Eval(toGenerate, browser.EvalOptions{
			Plugins: plugins,
			Visible: opts.Debug,
			Port:    opts.Port,
		})
		for _, g := range toGenerate {
			out := g.Output.(map[string]interface{})
			c := cardToCompile[out["id"].(string)]

			if out["rendered"].(bool) {
				c.If = "true"
			} else {
				c.If = "false"
			}

			cardToCompile[out["id"].(string)] = c
		}
	}

	for id, ca := range cardToCompile {
		if ca.If != "false" {
			if r := tx.
				Clauses(clause.OnConflict{
					DoNothing: true,
				}).
				Create(&Card{
					ID:         id,
					TemplateID: ca.Template.ID,
					NoteID:     ca.NoteID,
				}); r.Error != nil {
				return r.Error
			}
		} else {
			if r := tx.
				Where("template_id = ?", ca.Template.ID).
				Where("note_id = ?", ca.NoteID).
				Delete(&Card{}); r.Error != nil {
				return r.Error
			}
		}
	}

	for _, c := range loadFile.Card {
		c0 := Card{}
		if c.TemplateID != "" && c.NoteID != "" {
			if r := tx.
				Where("template_id = ?", c.TemplateID).
				Where("note_id = ?", c.NoteID).
				FirstOrInit(&c0); r.Error != nil {
				return r.Error
			}
		}

		tag, e := c0.Tag.Get()
		if e != nil {
			return e
		}

		for _, t := range c.Tag {
			tag[t] = true
		}
		if e := c0.Tag.Set(tag); e != nil {
			return e
		}

		if r := tx.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&Card{
			ID:         c.ID,
			TemplateID: c.TemplateID,
			NoteID:     c.NoteID,
			Tag:        c0.Tag,
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
