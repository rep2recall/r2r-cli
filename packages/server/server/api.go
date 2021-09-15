package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/rep2recall/r2r/db"
	"gorm.io/gorm"
)

type Router struct {
	DB     *gorm.DB
	Router fiber.Router
	Store  *session.Store
}

func (r *Router) Init() {
	r.DB = db.Connect()
	r.Store = session.New()
	r.Store.RegisterType([]db.Card{})

	r.quizRouter()
	r.cardRouter()
}
