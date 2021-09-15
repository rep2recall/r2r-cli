package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
)

func main() {
	t, err := tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		panic(err)
	}

	segments := []string{}

	for _, t := range t.Tokenize(os.Args[len(os.Args)-1]) {
		segments = append(segments, t.Surface)

		base, is_base := t.BaseForm()
		if is_base && base != t.Surface {
			segments = append(segments, base)
		}
	}

	fmt.Println(strings.Join(segments, " "))
}
