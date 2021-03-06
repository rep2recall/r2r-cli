package server

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/form3tech-oss/jwt-go"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/patarapolw/atexit"
	"github.com/rep2recall/r2r/shared"
	"gorm.io/gorm"

	jwtware "github.com/gofiber/jwt/v2"
)

type Server struct {
	DB         *gorm.DB
	Engine     *fiber.App
	Server     net.Listener
	port       int
	SubCommand []*exec.Cmd
}

type ServerOptions struct {
	Debug bool
	Proxy bool
	Port  int
}

func Serve(opts ServerOptions) Server {
	app := fiber.New()

	r := Server{
		Engine: app,
		port:   opts.Port,
	}

	app.Static("/", filepath.Join(shared.ExecDir, "public"))

	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))

	app.Use(logger.New(
		logger.Config{
			Output: shared.LogWriter,
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
		shared.Fatalln(e)
	}

	apiSrv.Post(
		"/login",
		limiter.New(limiter.Config{
			Max:        1,
			Expiration: 1 * time.Second,
		}),
		basicauth.New(basicauth.Config{
			Users: map[string]string{
				"DEFAULT": shared.Config.Secret,
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

	proxyRouter := app.Group(("/proxy"))

	for k, v := range shared.Config.Proxy {
		if len(v.Command) > 0 {
			dir := filepath.Join(shared.UserDataDir, "plugins", "app")
			cmd := exec.Command(filepath.Join(dir, v.Command[0]), v.Command[1:]...)
			cmd.Env = append(cmd.Env, fmt.Sprintf("PORT=%d", v.Port))
			cmd.Dir = dir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			r.SubCommand = append(r.SubCommand, cmd)
			e := cmd.Start()

			if e == nil {
				port := v.Port
				proxyRouter.Group(k).Use(func(c *fiber.Ctx) error {
					return proxy.Do(c, fmt.Sprintf("http://localhost:%d", port)+c.OriginalURL()[len(c.Route().Path):])
				})
			} else {
				shared.Logger.Println(e)
			}
		}
	}

	apiRouter := Router{
		Router: app.Group(
			"/api",
			limiter.New(limiter.Config{
				Max:        50,
				Expiration: 1 * time.Second,
			}),
			jwtware.New(jwtware.Config{
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

	shared.Logger.Printf("Server running at http://localhost:%d\n", opts.Port)

	listener, e := net.Listen("tcp", fmt.Sprintf(":%d", opts.Port))
	if e != nil {
		shared.Fatalln(e)
	}
	r.Server = listener

	go func() {
		if err := app.Listener(listener); err != nil && err != http.ErrServerClosed {
			shared.Fatalln(fmt.Sprintf("listen: %s\n", err))
		}
	}()

	atexit.Register(func() {
		r.Close()
	})

	return r
}

func (r Server) WaitUntilReady() {
	rootURL := fmt.Sprintf("http://localhost:%d", r.port)

	for {
		time.Sleep(1 * time.Second)
		_, err := http.Head(rootURL + "/server/ready")
		if err == nil {
			break
		}
	}
}

func (r Server) Close() {
	if r.DB != nil {
		r.DB.Commit()
		shared.Logger.Println("Committed the database")
	}

	for _, c := range r.SubCommand {
		c.Process.Kill()
		shared.Logger.Printf("Killed: %s\n", strings.Join(c.Args, " "))
	}

	r.Server.Close()
}
