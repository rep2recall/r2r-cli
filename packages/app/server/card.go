package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rep2recall/r2r/db"
)

func (r *Router) cardRouter() {
	router := r.Router.Group("/card")

	router.Get("/", func(c *fiber.Ctx) error {
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
			Preload("Template").Preload("Template.Model").Preload("Note.Attrs").
			First(&card); rTx.Error != nil {
			return fiber.NewError(fiber.StatusInternalServerError, rTx.Error.Error())
		}

		type outStruct struct {
			Raw  string                 `json:"raw"`
			Data map[string]interface{} `json:"data"`
		}
		out := new(outStruct)
		out.Data = make(map[string]interface{})

		for _, a := range card.Note.Attrs {
			key := a.Key
			v, e := a.Value.Get()
			if e != nil {
				return fiber.NewError(fiber.StatusInternalServerError, e.Error())
			}

			out.Data[key] = v
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

	router.Get("/mnemonic", func(c *fiber.Ctx) error {
		type queryStruct struct {
			ID string `validate:"required,uuid"`
		}

		query := new(queryStruct)
		if e := c.QueryParser(query); e != nil {
			return fiber.NewError(fiber.StatusBadRequest, e.Error())
		}

		var card db.Card
		if rTx := r.DB.
			Where("id = ?", query.ID).
			First(&card); rTx.Error != nil {
			return fiber.NewError(fiber.StatusInternalServerError, rTx.Error.Error())
		}

		if card.Mnemonic == "" {
			return c.SendString("{}")
		}

		return c.SendString(card.Mnemonic)
	})

	router.Put("/mnemonic", func(c *fiber.Ctx) error {
		type queryStruct struct {
			ID string `validate:"required,uuid"`
		}

		query := new(queryStruct)
		if e := c.QueryParser(query); e != nil {
			return fiber.NewError(fiber.StatusBadRequest, e.Error())
		}

		rTx := r.DB.
			Model(&db.Card{}).
			Where("id = ?", query.ID).
			Update("mnemonic", string(c.Body()))

		if rTx.Error != nil {
			return fiber.NewError(fiber.StatusInternalServerError, rTx.Error.Error())
		}

		if rTx.RowsAffected == 0 {
			return fiber.ErrNotFound
		}

		return c.SendStatus(fiber.StatusCreated)
	})

	router.Patch("/dSrsLevel", func(c *fiber.Ctx) error {
		type queryStruct struct {
			ID        string `validate:"required,uuid"`
			DSRSLevel int    `query:"dSrsLevel"`
			Session   string `validate:"required,uuid"`
		}

		query := new(queryStruct)
		if e := c.QueryParser(query); e != nil {
			return fiber.NewError(fiber.StatusBadRequest, e.Error())
		}

		sess, err := r.Store.Get(c)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		quizSession := sess.Get(query.Session)
		if quizSession == nil {
			return fiber.ErrNotFound
		}

		card := db.Card{}
		cards := quizSession.([]db.Card)
		for _, c := range cards {
			if c.ID == query.ID {
				card = c
			}
		}

		if card.ID == "" {
			return fiber.ErrNotFound
		}

		if err := card.UpdateSRSLevel(r.DB, query.DSRSLevel); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		return c.Status(fiber.StatusCreated).JSON(map[string]interface{}{
			"updated": true,
		})
	})

	router.Patch("/toggleMarked", func(c *fiber.Ctx) error {
		type queryStruct struct {
			ID string `validate:"required,uuid"`
		}

		query := new(queryStruct)
		if e := c.QueryParser(query); e != nil {
			return fiber.NewError(fiber.StatusBadRequest, e.Error())
		}

		var card db.Card
		if rTx := r.DB.
			Where("id = ?", query.ID).
			First(&card); rTx.Error != nil {
			return fiber.NewError(fiber.StatusInternalServerError, rTx.Error.Error())
		}

		tag, err := card.Tag.Get()
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		tag["marked"] = !tag["marked"]
		if err := card.Tag.Set(tag); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		if rTx := r.DB.
			Updates(&card); rTx.Error != nil {
			return fiber.NewError(fiber.StatusInternalServerError, rTx.Error.Error())
		}

		return c.Status(fiber.StatusCreated).JSON(map[string]bool{
			"isMarked": tag["marked"],
		})
	})
}
