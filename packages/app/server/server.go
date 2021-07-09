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

	"github.com/form3tech-oss/jwt-go"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/rep2recall/rep2recall/server/api"
	"github.com/rep2recall/rep2recall/shared"
	"gorm.io/gorm"

	jwtware "github.com/gofiber/jwt/v2"
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

	app.Use(logger.New(
		logger.Config{
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
		},
	))
	app.Use(logger.New(
		logger.Config{
			Output: filterUTF8{
				Target: f,
			},
			Format: "[${time}] ${status} - ${latency} ${method} ${path} ${queryParams}\t${body}\t${resBody}\n",
		},
	))

	apiSrv := app.Group("/server")

	apiSrv.Get(
		"/ready",
		limiter.New(limiter.Config{
			Max:        2,
			Expiration: 1 * time.Second,
		}),
		func(ctx *fiber.Ctx) error {
			return ctx.JSON(map[string]interface{}{
				"ready": true,
			})
		},
	)

	bootRand, e := shared.GenerateRandomBytes(64)
	if e != nil {
		log.Fatalln(e)
	}

	apiSrv.Post(
		"/login",
		limiter.New(limiter.Config{
			Max:        1,
			Expiration: 1 * time.Second,
		}),
		basicauth.New(basicauth.Config{
			Users: map[string]string{
				"DEFAULT": shared.ServerSecret(),
			},
		}),
		func(c *fiber.Ctx) error {
			token := jwt.New(jwt.SigningMethodHS256)

			// Set claims
			claims := token.Claims.(jwt.MapClaims)
			claims["name"] = c.Locals("username")
			claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

			// Generate encoded token and send it as response.
			t, err := token.SignedString(bootRand)
			if err != nil {
				return c.SendStatus(fiber.StatusInternalServerError)
			}

			return c.JSON(map[string]interface{}{
				"token": t,
			})
		},
	)

	apiRouter := api.Router{
		Router: app.Group(
			"/api",
			limiter.New(limiter.Config{
				Max:        50,
				Expiration: 1 * time.Second,
			}),
			jwtware.New(jwtware.Config{
				Filter: func(c *fiber.Ctx) bool {
					path := c.Path()
					return path == "/api/extra/gtts"
				},
				SigningKey: bootRand,
			}),
		),
	}
	apiRouter.Router.Post("/ok", func(c *fiber.Ctx) error {
		return c.JSON(map[string]interface{}{
			"ok": true,
		})
	})

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
	rootURL := fmt.Sprintf("http://localhost:%d", s.port)

	for {
		time.Sleep(1 * time.Second)
		_, err := http.Head(rootURL + "/server/ready")
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

type filterUTF8 struct {
	Target *os.File
}

func (f filterUTF8) Write(p []byte) (n int, err error) {
	segs := strings.Split(strings.TrimRight(string(p), "\n"), "\t")

	s := ""
	if len(segs) == 3 {
		if !isObject(segs[1]) {
			segs[1] = ""
		}
		if !isObject(segs[2]) {
			segs[2] = ""
		}
		s = strings.Join(segs, "\t")
	} else {
		s = segs[0]
	}
	if s[len(s)-1] != '\n' {
		s += "\n"
	}

	return f.Target.Write([]byte(s))
}

func isObject(s string) bool {
	return len(s) >= 2 && s[0] == '{' && s[len(s)-1] == '}'
}
