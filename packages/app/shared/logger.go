package shared

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/patarapolw/atexit"
)

var LogWriter io.Writer
var Logger *log.Logger

func init() {
	f, _ := os.Create(filepath.Join(UserDataDir, "r2r.log"))

	LogWriter = filterUTF8{
		Target: f,
	}

	Logger = log.New(io.MultiWriter(LogWriter, os.Stderr), "error", log.LstdFlags)
}

func Fatalln(e ...interface{}) {
	Logger.Fatalln(e...)
	atexit.Exit(1)
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
