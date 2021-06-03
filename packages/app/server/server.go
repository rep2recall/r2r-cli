package server

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rep2recall/rep2recall/db"
	"github.com/rep2recall/rep2recall/shared"
	"gorm.io/gorm"
)

type Server struct {
	DB     *gorm.DB
	Engine *gin.Engine
	Server *http.Server
	port   int
}

type ServerOptions struct {
	Debug bool
	Proxy bool
	Port  int
}

func Serve(opts ServerOptions) Server {
	if !opts.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	f, _ := os.Create(filepath.Join(shared.ExecDir, "gin.log"))
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)

	app := gin.New()

	r := Server{
		DB:     db.Connect(),
		Engine: app,
		port:   opts.Port,
	}

	app.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		ps := strings.SplitN(param.Path, "?", 2)
		path := ps[0]
		if len(ps) > 1 {
			q, e := url.QueryUnescape(ps[1])
			if e != nil {
				path += "?" + ps[1]
			} else {
				path += "?" + q
			}
		}

		out := []string{"[" + param.TimeStamp.Format(time.RFC3339) + "]"}
		out = append(out, param.Method)
		out = append(out, strconv.Itoa(param.StatusCode))
		out = append(out, param.Latency.String())
		out = append(out, path)

		if param.ErrorMessage != "" {
			out = append(out, param.ErrorMessage)
		}

		out = append(out, "\n")

		return strings.Join(out, " ")
	}))
	app.Use(gin.Recovery())

	app.Use(func(c *gin.Context) {
		b, _ := ioutil.ReadAll(c.Request.Body)

		if len(b) > 0 {
			c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(b))

			gin.DefaultWriter.Write([]byte(c.Request.Method + " " + c.Request.URL.Path + " body: "))
			gin.DefaultWriter.Write(b)
			gin.DefaultWriter.Write([]byte("\n"))
		}
		c.Next()
	})

	csrfToken := ""
	if !opts.Proxy {
		csrfToken = uuid.NewString()
	}

	app.Use(func(c *gin.Context) {
		if c.Request.Method == "GET" {
			if csrfToken != "" {
				c.SetCookie("csrf_token", csrfToken, 2592000, "/", "localhost", false, true)
			}

			static.Serve("/", static.LocalFile(filepath.Join(shared.ExecDir, "public"), true))(c)
			return
		}
		c.Next()
	})

	app.GET("/server/config", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"ready": true,
		})
	})

	// apiRouter := app.Group("/api")

	fmt.Printf("Server running at http://localhost:%d\n", opts.Port)

	r.Server = &http.Server{
		Addr:    fmt.Sprintf(":%d", opts.Port),
		Handler: app,
	}

	go func() {
		if err := r.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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
	s.DB.Commit()
	s.Server.Close()
}
