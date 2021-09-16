package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"

	"github.com/gofiber/fiber/v2"
	"github.com/patarapolw/atexit"
	"github.com/rep2recall/r2r/browser"
	"github.com/rep2recall/r2r/db"
	"github.com/rep2recall/r2r/server"
	"github.com/rep2recall/r2r/shared"
	"github.com/thatisuday/commando"
	"gorm.io/gorm"
)

func main() {
	defer atexit.ListenPanic()

	version := "0.4.1"

	commando.
		SetExecutableName("r2r").
		SetVersion(version).
		SetDescription("Repeat Until Recall - a simple, yet powerful, flashcard app")

	commando.
		Register(nil).
		AddFlag("db,o", "database to use", commando.String, shared.Config.DB).
		AddFlag("port,p", "port to run the server", commando.Int, shared.Config.Port).
		AddFlag("browser,b", "browser to open (default: Chrome with Edge fallback)", commando.String, ".").
		AddFlag("mode,m", "mode to run in (app / server / proxy / quiz)", commando.String, "app").
		AddFlag("file,f", "files to use (must be loaded first)", commando.String, ".").
		AddFlag("filter", "keyword to filter", commando.String, ".").
		AddFlag("debug", "whether to run in debug mode", commando.Bool, false).
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			debug := false
			browserOfChoice := ""
			mode := ""
			files := make([]string, 0)
			filter := ""

			for k, v := range flags {
				switch k {
				case "db", "o":
					shared.Config.DB = v.Value.(string)
				case "port", "p":
					shared.Config.Port = v.Value.(int)
				case "browser", "b":
					browserOfChoice = v.Value.(string)
				case "mode", "m":
					mode = v.Value.(string)
				case "file", "f":
					f := v.Value.(string)
					if f != "." {
						files = append(files, f)
					}
				case "filter":
					value := v.Value.(string)
					if value != "." {
						filter = value
					}
				}
			}

			atexit.Listen()

			switch mode {
			case "server", "proxy":
				s := server.Serve(server.ServerOptions{
					Proxy: mode == "proxy",
					Debug: debug,
					Port:  shared.Config.Port,
				})

				forever := make(chan bool)

				log.Printf("[*] To exit press CTRL+C")
				<-forever

				s.Close()
			case "quiz":
				fileString := ""
				if len(files) > 0 {
					b, e := json.Marshal(&files)
					if e == nil {
						fileString = string(b)
					}
				}

				if browserOfChoice == "." {
					browserOfChoice = ""
				}
				s := server.Serve(server.ServerOptions{
					Proxy: false,
					Debug: debug,
					Port:  shared.Config.Port,
				})

				s.WaitUntilReady()

				rootURL := fmt.Sprintf("http://localhost:%d", shared.Config.Port)

				var authOutput struct {
					Token string `json:"token"`
				}
				code, _, e := fiber.Post(rootURL+"/server/login").BasicAuth("DEFAULT", shared.Config.Secret).Struct(&authOutput)
				if e != nil {
					shared.Fatalln(e)
				}
				if code != 200 {
					shared.Fatalln(fiber.ErrUnauthorized)
				}

				b := browser.Browser{
					ExecPath: browserOfChoice,
				}
				b.AppMode(
					rootURL+fmt.Sprintf(
						"/quiz?q=%s&files=%s&token=%s",
						url.QueryEscape(filter),
						url.QueryEscape(fileString),
						authOutput.Token,
					),
					browser.WindowSize(600, 800),
				)

				s.Close()
			default:
				if browserOfChoice == "." {
					browserOfChoice = ""
				}
				s := server.Serve(server.ServerOptions{
					Proxy: false,
					Debug: debug,
					Port:  shared.Config.Port,
				})

				s.WaitUntilReady()

				rootURL := fmt.Sprintf("http://localhost:%d", shared.Config.Port)

				var authOutput struct {
					Token string `json:"token"`
				}
				code, _, e := fiber.Post(rootURL+"/server/login").BasicAuth("DEFAULT", shared.Config.Secret).Struct(&authOutput)
				if e != nil {
					shared.Fatalln(e)
				}
				if code != 200 {
					shared.Fatalln(fiber.ErrUnauthorized)
				}

				b := browser.Browser{
					ExecPath: browserOfChoice,
				}
				b.AppMode(rootURL+fmt.Sprintf("/app?token=%s", authOutput.Token), browser.IsMaximized())

				s.Close()
			}
		})

	commando.
		Register("load").
		SetShortDescription("load the YAML into the database and exit").
		AddArgument("files...", "directory or YAML to scan for IDs", ""). // required
		AddFlag("db,o", "database to use", commando.String, shared.Config.DB).
		AddFlag("port,p", "port to run the server", commando.Int, shared.Config.Port).
		AddFlag("debug", "debug mode (Chrome headful mode)", commando.Bool, false).
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			debug := false

			for k, v := range flags {
				switch k {
				case "db", "o":
					shared.Config.DB = v.Value.(string)
				case "port", "p":
					shared.Config.Port = v.Value.(int)
				case "debug":
					debug = v.Value.(bool)
				}
			}

			atexit.Listen()

			s := server.Serve(server.ServerOptions{
				Proxy: false,
				Debug: debug,
				Port:  shared.Config.Port,
			})

			s.WaitUntilReady()

			if e := s.DB.Transaction(func(tx *gorm.DB) error {
				for k, v := range args {
					if k == "files" {
						if e := db.Load(tx, v.Value, db.LoadOptions{
							Debug: debug,
							Port:  shared.Config.Port,
						}); e != nil {
							return e
						}
					}
				}

				return nil
			}); e != nil {
				panic(e)
			}

			s.Close()
		})

	// parse command-line arguments
	commando.Parse(nil)
}
