package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
)

type Tokens []tokenizer.Token

func (t Tokens) SearchForm() []string {
	segments := []string{}

	for _, t := range t {
		segments = append(segments, t.Surface)

		base, is_base := t.BaseForm()
		if is_base && base != t.Surface {
			segments = append(segments, base)
		}
	}

	return segments
}

func (t Tokens) BaseForm() []string {
	segments := []string{}

	for _, t := range t {
		base, is_base := t.BaseForm()
		if is_base {
			segments = append(segments, base)
		} else {
			segments = append(segments, t.Surface)
		}
	}

	return segments
}

var t *tokenizer.Tokenizer

func Tokenize(s string) Tokens {
	return t.Tokenize(s)
}

func main() {
	t0, err := tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		panic(err)
	}
	t = t0

	if len(os.Args) > 1 {
		fmt.Println(strings.Join(Tokenize(os.Args[1]).SearchForm(), " "))
	} else {
		app := fiber.New()
		app.Get("/proxy/ja/tokenize", func(c *fiber.Ctx) error {
			var query struct {
				Q string `query:"q" validate:"required"`
			}

			if e := c.QueryParser(&query); e != nil {
				return fiber.ErrBadRequest
			}

			return c.JSON(map[string]interface{}{
				"result": Tokenize(query.Q).BaseForm(),
			})
		})
		port := os.Getenv("PORT")
		if port == "" {
			port = "24899"
		}
		app.Listen(":" + port)
	}
}
