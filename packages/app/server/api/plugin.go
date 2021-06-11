package api

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"plugin"
	"runtime"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rep2recall/rep2recall/shared"
)

func (r *Router) pluginRouter() {
	router := r.Router.Group("/plugin")

	files := []string{}
	ext := ".so"
	switch runtime.GOOS {
	case "darwin":
		ext = ".dylib"
	case "windows":
		ext = ".dll"
	}

	e := filepath.Walk(filepath.Join(shared.UserDataDir(), "plugins"), func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, ext) {
			files = append(files, path)
		}

		return nil
	})
	if e != nil {
		log.Fatalln(e)
	}

	for _, f := range files {
		p, err := plugin.Open(f)
		if err != nil {
			log.New(os.Stderr, "", log.LstdFlags).Println(err)
		}
		if p == nil {
			log.Printf("did not load plugin: %s", f)
			continue
		}

		f, err := p.Lookup("Router")
		if err != nil {
			log.New(os.Stderr, "", log.LstdFlags).Println(err)
			continue
		}
		if f == nil {
			log.Printf("did not load plugin.Router: %s", f)
			continue
		}

		f.(func(router *fiber.Router))(&router) // prints "Hello, number 7"
	}
}
