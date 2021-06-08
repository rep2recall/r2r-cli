package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rep2recall/rep2recall/db"
	"gorm.io/gorm"
)

type Router struct {
	DB     *gorm.DB
	Router fiber.Router
}

func (r *Router) Init() {
	r.DB = db.Connect()

	r.Router.Get("/leech", func(c *fiber.Ctx) error {
		type queryStruct struct {
			Page   int `validate:"required,min=1"`
			Limit  int `validate:"required,min=3"`
			Filter string
		}

		query := new(queryStruct)
		if e := c.QueryParser(query); e != nil {
			return fiber.NewError(fiber.StatusBadRequest, e.Error())
		}

		var cards []db.Card
		if rTx := r.DB.
			Limit(query.Limit).
			Offset((query.Page - 1) * query.Limit).
			Select("ID").
			// Preload("Template").Preload("Template.Model").Preload("Notes").
			Find(&cards); rTx.Error != nil {
			return fiber.NewError(fiber.StatusInternalServerError, rTx.Error.Error())
		}

		type outStruct struct {
			Result []string `json:"result"`
		}
		out := outStruct{
			Result: make([]string, 0),
		}
		for _, c := range cards {
			out.Result = append(out.Result, c.ID)
		}

		return c.JSON(out)
	})

	r.Router.Get("/item", func(c *fiber.Ctx) error {
		type queryStruct struct {
			ID   string `validate:"required,uuid"`
			Side string `validate:"required,oneof=front back mnemonic"`
		}

		query := new(queryStruct)
		if e := c.QueryParser(query); e != nil {
			return fiber.NewError(fiber.StatusBadRequest, e.Error())
		}

		var card db.Card
		if rTx := r.DB.
			Where("id = ?", query.ID).
			Preload("Template").Preload("Template.Model").
			First(&card); rTx.Error != nil {
			return fiber.NewError(fiber.StatusInternalServerError, rTx.Error.Error())
		}

		type outStruct struct {
			Raw  string                 `json:"raw"`
			Data map[string]interface{} `json:"data"`
		}
		out := new(outStruct)
		out.Data = make(map[string]interface{})

		notes, err := card.Notes(r.DB)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		for _, n := range notes {
			v, e := n.Data.Get()
			if e != nil {
				return fiber.NewError(fiber.StatusInternalServerError, e.Error())
			}

			out.Data[n.Key] = v
		}

		if query.Side != "mnemonic" {
			out.Raw = func() string {
				if card.Shared != "" {
					return card.Shared
				}

				if card.TemplateID != "" {
					if card.Template.Shared != "" {
						return card.Template.Shared
					}

					if card.Template.ModelID != "" {
						return card.Template.Model.Shared
					}
				}

				return ""
			}()
		}

		switch query.Side {
		case "front":
			out.Raw += "\n" + func() string {
				if card.Front != "" {
					return card.Front
				}

				if card.TemplateID != "" {
					if card.Template.Front != "" {
						return card.Template.Front
					}

					if card.Template.ModelID != "" {
						return card.Template.Model.Front
					}
				}

				return ""
			}()
		case "back":
			out.Raw += "\n" + func() string {
				if card.Back != "" {
					return card.Back
				}

				if card.TemplateID != "" {
					if card.Template.Back != "" {
						return card.Template.Back
					}

					if card.Template.ModelID != "" {
						return card.Template.Model.Back
					}
				}

				return ""
			}()
		case "mnemonic":
			out.Raw += "\n" + card.Mnemonic
		}

		return c.JSON(out)
	})
}
