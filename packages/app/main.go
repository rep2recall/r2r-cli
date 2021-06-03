package main

import (
	"fmt"
	"log"

	"github.com/rep2recall/rep2recall/browser"
	"github.com/rep2recall/rep2recall/db"
	"github.com/rep2recall/rep2recall/server"
	"github.com/thatisuday/commando"
	"gorm.io/gorm"
)

func main() {
	version := "0.1.0"
	defaultPort := 25459

	commando.
		SetExecutableName("rep2recall").
		SetVersion(version).
		SetDescription("Repeat Until Recall - a simple, yet powerful, flashcard app")

	commando.
		Register(nil).
		SetShortDescription("open in GUI mode, for full interaction").
		AddFlag("port,p", "port to run the server", commando.Int, defaultPort).
		AddFlag("debug", "whether to run in debug mode", commando.Bool, false).
		AddFlag("browser", "browser to open (default: Chrome with Edge fallback)", commando.String, "."). // not required
		AddFlag("server", "run in server mode (don't open the browser)", commando.Bool, false).
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			port := defaultPort
			debug := false
			browserOfChoice := "."
			isServer := false

			for k, v := range flags {
				switch k {
				case "port":
					port = v.Value.(int)
				case "debug":
					debug = v.Value.(bool)
				case "browser":
					browserOfChoice = v.Value.(string)
				case "server":
					isServer = v.Value.(bool)
				}
			}

			if isServer {
				s := server.Serve(server.ServerOptions{
					Proxy: false,
					Debug: debug,
					Port:  port,
				})

				forever := make(chan bool)

				log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
				<-forever

				s.Close()
			} else {
				if browserOfChoice == "." {
					browserOfChoice = ""
				}
				s := server.Serve(server.ServerOptions{
					Proxy: false,
					Debug: debug,
					Port:  port,
				})

				s.WaitUntilReady()

				b := browser.Browser{
					ExecPath: browserOfChoice,
				}
				b.AppMode(fmt.Sprintf("http://localhost:%d", port), browser.IsMaximized())

				s.Close()
			}
		})

	commando.
		Register("proxy").
		SetShortDescription("start as proxy server, for development").
		AddFlag("port,p", "port to run the server", commando.Int, defaultPort).
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			port := defaultPort

			for k, v := range flags {
				if k == "port" {
					port = v.Value.(int)
				}
			}

			s := server.Serve(server.ServerOptions{
				Proxy: true,
				Debug: true,
				Port:  port,
			})

			forever := make(chan bool)

			log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
			<-forever

			s.Close()
		})

	commando.
		Register("load").
		SetShortDescription("load the YAML into the database and exit").
		AddArgument("files...", "directory or YAML to scan for IDs", ""). // required
		AddFlag("debug", "debug mode (Chrome headful mode)", commando.Bool, false).
		AddFlag("port,p", "port to run the server", commando.Int, defaultPort).
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			debug := false
			port := defaultPort

			for k, v := range flags {
				if k == "debug" {
					debug = v.Value.(bool)
				}
			}

			for k, v := range flags {
				switch k {
				case "port":
					port = v.Value.(int)
				case "debug":
					debug = v.Value.(bool)
				}
			}

			s := server.Serve(server.ServerOptions{
				Proxy: false,
				Debug: debug,
				Port:  port,
			})

			s.WaitUntilReady()

			if e := s.DB.Transaction(func(tx *gorm.DB) error {
				for k, v := range args {
					if k == "files" {
						if e := db.Load(tx, v.Value, db.LoadOptions{
							Debug: debug,
							Port:  port,
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

	commando.
		Register("clean").
		SetShortDescription("clean the to-be-delete part of the database and exit").
		AddArgument("files...", "directory or YAML to scan for IDs, or none to use the whole database", "."). // not required
		AddFlag("filter,f", "keyword to filter", commando.String, ".").                                       // not required
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			database := db.Connect()

			if e := database.Transaction(func(tx *gorm.DB) error {
				if e := (db.Card{}).Tidy(tx); e != nil {
					return e
				}

				if e := (db.Note{}).Tidy(tx); e != nil {
					return e
				}

				if e := (db.Template{}).Tidy(tx); e != nil {
					return e
				}

				if e := (db.Model{}).Tidy(tx); e != nil {
					return e
				}

				if r := tx.Unscoped().Where("TRUE").Delete(&db.Card{}); r.Error != nil {
					return r.Error
				}

				if r := tx.Unscoped().Where("TRUE").Delete(&db.Note{}); r.Error != nil {
					return r.Error
				}

				if r := tx.Unscoped().Where("TRUE").Delete(&db.Template{}); r.Error != nil {
					return r.Error
				}

				if r := tx.Unscoped().Where("TRUE").Delete(&db.Model{}); r.Error != nil {
					return r.Error
				}

				return nil
			}); e != nil {
				panic(e)
			}
		})

	commando.
		Register("quiz").
		SetShortDescription("open the quiz window only").
		AddArgument("files...", "directory or YAML to scan for IDs, or none to use the whole database", "."). // not required
		AddFlag("filter,f", "keyword to filter", commando.String, ".").                                       // not required
		AddFlag("port,p", "port to run the server", commando.Int, defaultPort).                               // not required
		AddFlag("browser", "browser to open (default: Chrome with Edge fallback)", commando.String, ".").     // not required
		AddFlag("debug", "debug mode (Chrome headful mode)", commando.Bool, false).
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			port := defaultPort
			debug := false
			browserOfChoice := "."

			for k, v := range flags {
				switch k {
				case "port":
					port = v.Value.(int)
				case "debug":
					debug = v.Value.(bool)
				case "browser":
					browserOfChoice = v.Value.(string)
				}
			}

			if browserOfChoice == "." {
				browserOfChoice = ""
			}
			s := server.Serve(server.ServerOptions{
				Proxy: false,
				Debug: debug,
				Port:  port,
			})

			s.WaitUntilReady()

			b := browser.Browser{
				ExecPath: browserOfChoice,
			}
			b.AppMode(fmt.Sprintf("http://localhost:%d/quiz.html", port), browser.IsMaximized())

			s.Close()
		})

	// parse command-line arguments
	commando.Parse(nil)
}
