package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
	"github.com/rep2recall/duolog"
)

type Tokens []tokenizer.Token

func (t Tokens) SearchForm() []string {
	regex := regexp.MustCompile(`[\p{Han}\p{Katakana}\p{Hiragana}]`)
	m := make(map[string]bool)
	doAppend := func(t string) {
		if regex.MatchString(t) {
			m[t] = true
		}
	}

	for _, t := range t {
		doAppend(t.Surface)

		base, is_base := t.BaseForm()
		if is_base && base != t.Surface {
			doAppend(base)
		}
	}

	tokens := []string{}
	for k, v := range m {
		if v {
			tokens = append(tokens, k)
		}
	}

	return tokens
}

func (t Tokens) BaseForm() []string {
	regex := regexp.MustCompile(`[\p{Han}\p{Katakana}\p{Hiragana}]`)
	m := make(map[string]bool)
	doAppend := func(t string) {
		if regex.MatchString(t) {
			m[t] = true
		}
	}

	for _, t := range t {
		base, is_base := t.BaseForm()
		if is_base {
			doAppend(base)
		} else {
			doAppend(t.Surface)
		}
	}

	tokens := []string{}
	for k, v := range m {
		if v {
			tokens = append(tokens, k)
		}
	}

	return tokens
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

		d := duolog.Duolog{
			NoColor: true,
		}
		d.New()
		app.Use(logger.New(logger.Config{
			Output: d,
			Format: "[${time}] :${port} ${status} - ${latency} ${method} ${path} ${queryParams}\n",
		}))

		app.Get("/tokenize", func(c *fiber.Ctx) error {
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
