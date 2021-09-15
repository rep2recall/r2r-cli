package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/patarapolw/atexit"
	"github.com/yanyiwu/gojieba"
)

var jieba *gojieba.Jieba

func Tokenize(s string) []string {
	return jieba.CutForSearch(s, true)
}

func main() {
	jieba = gojieba.NewJieba()
	atexit.Register(jieba.Free)
	atexit.Listen()
	defer jieba.Free()

	if len(os.Args) > 1 {
		fmt.Println(strings.Join(Tokenize(os.Args[1]), " "))
	} else {
		app := fiber.New()
		app.Get("/proxy/zh/tokenize", func(c *fiber.Ctx) error {
			var query struct {
				Q string `query:"q" validate:"required"`
			}

			if e := c.QueryParser(&query); e != nil {
				return fiber.ErrBadRequest
			}

			return c.JSON(map[string]interface{}{
				"result": Tokenize(query.Q),
			})
		})
		port := os.Getenv("PORT")
		if port == "" {
			port = "27002"
		}
		app.Listen(":" + port)
	}
}
