package api

import (
	"math/rand"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rep2recall/rep2recall/db"
	"gorm.io/gorm"
)

func (r *Router) quizRouter() {
	router := r.Router.Group("/quiz")

	router.Get("/session", func(c *fiber.Ctx) error {
		type queryStruct struct {
			Session string `validate:"required,uuid"`
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

		type cardStruct struct {
			ID       string `json:"id"`
			IsMarked bool   `json:"isMarked"`
		}

		type outStruct struct {
			Result []cardStruct `json:"result"`
		}
		out := outStruct{
			Result: make([]cardStruct, 0),
		}

		cards := quizSession.([]db.Card)
		for _, c := range cards {
			tag, err := c.Tag.Get()
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, err.Error())
			}

			out.Result = append(out.Result, cardStruct{
				ID:       c.ID,
				IsMarked: tag["marked"],
			})
		}

		return c.JSON(out)
	})

	router.Post("/init", func(c *fiber.Ctx) error {
		query := getCardStruct{}
		if e := c.QueryParser(&query); e != nil {
			return fiber.NewError(fiber.StatusBadRequest, e.Error())
		}

		cards, err := getCard(r.DB, query)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(cards), func(i, j int) {
			cards[i], cards[j] = cards[j], cards[i]
		})

		sess, err := r.Store.Get(c)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		sessionID := uuid.NewString()
		sess.Set(sessionID, cards)
		if err := sess.Save(); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		type outStruct struct {
			ID string `json:"id"`
		}

		return c.JSON(outStruct{
			ID: sessionID,
		})
	})

	router.Get("/stat", func(c *fiber.Ctx) error {
		query := getCardStruct{}
		if e := c.QueryParser(&query); e != nil {
			return fiber.NewError(fiber.StatusBadRequest, e.Error())
		}

		cards, err := getCard(r.DB, query)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		type outStruct struct {
			New   int `json:"new"`
			Due   int `json:"due"`
			Leech int `json:"leech"`
		}
		out := outStruct{}

		for _, c := range cards {
			status, err := c.Status.Get()
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, err.Error())
			}

			if status["new"] {
				out.New += 1
			}

			if status["due"] {
				out.Due += 1
			}

			if status["leech"] {
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

type getCardStruct struct {
	Q     string
	State string
}

func getCard(tx *gorm.DB, query getCardStruct) ([]db.Card, error) {
	rTx := db.Search(tx, query.Q)

	rState := tx.Where("FALSE")
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
		Find(&cards); rTx.Error != nil {
		return nil, rTx.Error
	}

	return cards, nil
}
