package main

import (
	"fmt"

	"github.com/rep2recall/rep2recall/db"
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
		AddFlag("browser", "browser to open (default: Chrome with Edge fallback)", commando.String, "."). // not required
		AddFlag("server", "run in server mode (don't open the browser)", commando.Bool, false).
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			fmt.Printf("Printing options of the `root` command...\n\n")

			// print arguments
			for k, v := range args {
				fmt.Printf("arg -> %v: %v(%T)\n", k, v.Value, v.Value)
			}

			// print flags
			for k, v := range flags {
				fmt.Printf("flag -> %v: %v(%T)\n", k, v.Value, v.Value)
			}
		})

	commando.
		Register("load").
		SetShortDescription("load the directory / YAML into the database and exit").
		AddArgument("dir...", "directory or YAML to scan for IDs", ""). // required
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			fmt.Printf("Printing options of the `load` command...\n\n")

			// print arguments
			for k, v := range args {
				fmt.Printf("arg -> %v: %v(%T)\n", k, v.Value, v.Value)
			}

			// print flags
			for k, v := range flags {
				fmt.Printf("flag -> %v: %v(%T)\n", k, v.Value, v.Value)
			}
		})

	commando.
		Register("clean").
		SetShortDescription("clean the to-be-delete part of the database and exit").
		AddArgument("dir...", "directory or YAML to scan for IDs, or none to use the whole database", "."). // not required
		AddFlag("filter,f", "keyword to filter", commando.String, ".").                                     // not required
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			database := db.Connect()

			fmt.Printf("Printing options of the `clean` command...\n\n")

			// print arguments
			for k, v := range args {
				fmt.Printf("arg -> %v: %v(%T)\n", k, v.Value, v.Value)
			}

			// print flags
			for k, v := range flags {
				fmt.Printf("flag -> %v: %v(%T)\n", k, v.Value, v.Value)
			}

			database.Transaction(func(tx *gorm.DB) error {
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

				if r := tx.Unscoped().Delete(db.Card{}); r.Error != nil {
					return r.Error
				}

				if r := tx.Unscoped().Delete(db.Note{}); r.Error != nil {
					return r.Error
				}

				if r := tx.Unscoped().Delete(db.Template{}); r.Error != nil {
					return r.Error
				}

				if r := tx.Unscoped().Delete(db.Model{}); r.Error != nil {
					return r.Error
				}

				return nil
			})
		})

	commando.
		Register("quiz").
		SetShortDescription("open the quiz window only").
		AddArgument("dir...", "directory or YAML to scan for IDs, or none to use the whole database", "."). // not required
		AddFlag("filter,f", "keyword to filter", commando.String, ".").                                     // not required
		AddFlag("port,p", "port to run the server", commando.Int, defaultPort).                             // not required
		AddFlag("browser", "browser to open (default: Chrome with Edge fallback)", commando.String, ".").   // not required
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			fmt.Printf("Printing options of the `quiz` command...\n\n")

			// print arguments
			for k, v := range args {
				fmt.Printf("arg -> %v: %v(%T)\n", k, v.Value, v.Value)
			}

			// print flags
			for k, v := range flags {
				fmt.Printf("flag -> %v: %v(%T)\n", k, v.Value, v.Value)
			}
		})

	// parse command-line arguments
	commando.Parse(nil)
}
