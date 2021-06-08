package api

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rep2recall/rep2recall/db"
)

func (r *Router) quizRouter() {
	router := r.Router.Group("/quiz")

	router.Get("/stat", func(c *fiber.Ctx) error {
		type queryStruct struct {
			Q     string
			State string
		}

		query := new(queryStruct)
		if e := c.QueryParser(query); e != nil {
			return fiber.NewError(fiber.StatusBadRequest, e.Error())
		}

		rTx := db.Search(r.DB, query.Q)

		rState := r.DB.Where("FALSE")
		if len(query.State) > 0 {
			for _, s := range strings.Split(query.State, ",") {
				switch s {
				case "new":
					rState = rState.Or("card.next_review IS NULL")
				case "learning":
					rState = rState.Or("card.srs_level <= 3")
				case "graduated":
					rState = rState.Or("card.srs_level > 3")
				case "leech":
					rState = rState.Or("card.wrong_streak > 1")
				}
			}

			for _, s := range strings.Split(query.State, ",") {
				switch s {
				case "due":
					rState = rState.Where("strftime('%s', card.next_review) < strftime('%s', 'now')")
				}
			}
		}

		var cards []db.Card
		if rTx := rTx.Where(rState).
			Select("card.status").
			Find(&cards); rTx.Error != nil {
			return fiber.NewError(fiber.StatusInternalServerError, rTx.Error.Error())
		}

		type outStruct struct {
			New   int `json:"new"`
			Due   int `json:"due"`
			Leech int `json:"leech"`
		}
		out := outStruct{}

		for _, c := range cards {
			if strings.Contains(c.Status, "new") {
				out.New += 1
			}

			if strings.Contains(c.Status, "due") {
				out.Due += 1
			}

			if strings.Contains(c.Status, "leech") {
				out.Leech += 1
			}
		}

		return c.JSON(out)
	})

	router.Get("/leech", func(c *fiber.Ctx) error {
		type queryStruct struct {
			Page  int `validate:"required,min=1"`
			Limit int `validate:"required,min=3"`
			Q     string
		}

		query := new(queryStruct)
		if e := c.QueryParser(query); e != nil {
			return fiber.NewError(fiber.StatusBadRequest, e.Error())
		}

		var cards []db.Card
		if rTx := db.Search(r.DB, query.Q).
			// Where("card.wrong_streak > 1").
			Limit(query.Limit).
			Offset((query.Page - 1) * query.Limit).
			Select("card.id").
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
}
