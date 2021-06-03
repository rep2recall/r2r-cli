package main

import (
	"fmt"

	"github.com/rep2recall/rep2recall/browser"
)

func main() {
	b := browser.Browser{}

	ev := browser.EvalContext{
		JS: "1 + 1",
	}
	b.Eval([]string{}, &ev)
	fmt.Println(ev.Output)

	b.AppMode("https://www.duckduckgo.com", browser.IsMaximized())
}
