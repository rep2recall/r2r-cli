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

	r.quizRouter()
	r.cardRouter()
}
