package shared

import (
	"io"
	"log"
	"path/filepath"

	"github.com/patarapolw/atexit"
	"github.com/rep2recall/duolog"
)

var LogWriter io.Writer
var Logger *log.Logger

func init() {
	d := duolog.Duolog{
		Filename: filepath.Join(UserDataDir, "r2r.log"),
	}
	d.New()

	LogWriter = d
	Logger = d.Logger()
}

func Fatalln(e ...interface{}) {
	Logger.Fatalln(e...)
	atexit.Exit(1)
}
