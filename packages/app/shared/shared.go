package shared

import (
	"os"
	"path/filepath"
)

// ExecDir is dirname of executable
var ExecDir string

func init() {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	ExecDir = dir
}

// UserDataDir is used to store all writable data
func UserDataDir() string {
	dir := os.Getenv("USER_DATA_DIR")
	if dir == "" {
		return ExecDir
	}

	return dir
}
