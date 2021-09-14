package db

import (
	"regexp"
	"strings"
	"testing"
)

func TestQSearch(t *testing.T) {
	input := []string{
		`type:emoji`,
		`type:emoji ?function`,
		`type:emoji ?function -funeral`,
		`"a b":c ?c:"a b" -"a b"`,
	}

	for _, inp := range input {
		out, err := qSearch(inp)
		if err != nil {
			t.Fatal(err)
		}

		for _, current := range out {
			if current.Key == "" {
				t.Fatalf("no key: %s", inp)
			}

			if len(current.Sign) > 1 || (len(current.Sign) > 0 && strings.Contains(string(current.Sign[0]), "?-")) {
				t.Fatalf("bad sign: %s", current.Sign)
			}

			if current.Op != "" && !regexp.MustCompile(`^([><]=?|[=:])$`).Match([]byte(current.Op)) {
				t.Fatalf("bad op: %s", current.Op)
			}
		}

		t.Log(out)
	}
}
