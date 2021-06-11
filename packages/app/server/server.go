package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/rep2recall/rep2recall/server/api"
	"github.com/rep2recall/rep2recall/shared"
	"gorm.io/gorm"
)

type Server struct {
	DB     *gorm.DB
	Engine *fiber.App
	Server net.Listener
	port   int
}

type ServerOptions struct {
	Debug bool
	Proxy bool
	Port  int
}

func Serve(opts ServerOptions) Server {
	f, _ := os.Create(filepath.Join(shared.ExecDir, "server.log"))

	app := fiber.New()
	shared.ServerSecret()

	r := Server{
		Engine: app,
		port:   opts.Port,
	}

	app.Static("/", filepath.Join(shared.ExecDir, "public"))

	app.Use(recover.New())

	app.Use(logger.New(logger.Config{
		Next: func(c *fiber.Ctx) bool {
			body := c.Body()
			prettyBody := ""
			if len(body) > 0 {
				prettyBody = func() string {
					var str map[string]interface{}
					if e := json.Unmarshal(body, &str); e != nil {
						log.New(os.Stderr, "", log.LstdFlags).Println(e)
						return ""
					}

					b, e := json.MarshalIndent(str, "", "  ")
					if e != nil {
						log.New(os.Stderr, "", log.LstdFlags).Println(e)
						return ""
					}

					return string(b)
				}()
			}

			if prettyBody != "" {
				log.Printf("body: %s", prettyBody)
			}

			return false
		},
	}))
	app.Use(logger.New(logger.Config{
		Output: f,
		Format: "[${time}] ${status} - ${latency} ${method} ${path} ${queryParams} ${body} ${resBody}\n",
	}))

	if !opts.Proxy {
		app.Use(csrf.New(csrf.Config{
			Next: func(c *fiber.Ctx) bool {
				return strings.HasPrefix(c.Path(), "/server/")
			},
			KeyLookup:  "cookie:_csrf",
			CookieName: "_csrf",
		}))
	}

	apiSrv := app.Group("/server")

	apiSrv.Get("/config", func(ctx *fiber.Ctx) error {
		return ctx.JSON(map[string]interface{}{
			"ready": true,
		})
	})

	checkSecret := func(ctx *fiber.Ctx) bool {
		xSecret := ctx.Get("X-Secret")
		if xSecret == "" {
			xSecret = ctx.Query("secret")
		}

		return xSecret == shared.ServerSecret()
	}

	apiSrv.Post("/login", func(ctx *fiber.Ctx) error {
		if !checkSecret(ctx) {
			return fiber.ErrUnauthorized
		}

		return ctx.JSON(map[string]interface{}{
			"ok": true,
		})
	})

	apiRouter := api.Router{
		Router: app.Group("/api", func(ctx *fiber.Ctx) error {
			if !checkSecret(ctx) {
				return fiber.ErrUnauthorized
			}

			return ctx.Next()
		}),
	}
	apiRouter.Init()

	r.DB = apiRouter.DB

	log.Printf("Server running at http://localhost:%d\n", opts.Port)

	listener, e := net.Listen("tcp", fmt.Sprintf(":%d", opts.Port))
	if e != nil {
		log.Fatalln(e)
	}
	r.Server = listener

	go func() {
		if err := app.Listener(listener); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	return r
}

func (s Server) WaitUntilReady() {
	url := fmt.Sprintf("http://localhost:%d", s.port)

	for {
		time.Sleep(1 * time.Second)
		_, err := http.Head(url + "/server/config")
		if err == nil {
			break
		}
	}
}

func (s Server) Close() {
	if s.DB != nil {
		s.DB.Commit()
	}

	s.Server.Close()
}
