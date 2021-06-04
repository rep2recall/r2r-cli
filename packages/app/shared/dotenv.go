package shared

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func init() {
	loadenv()
}

// loadenv loads from .env and .env.local to os.Getenv
func loadenv() {
	godotenv.Load(filepath.Join(ExecDir, ".env.local"))
	godotenv.Load(filepath.Join(ExecDir, ".env"))
}

// Setenv sets to .env.local and os.Getenv
func Setenv(key string, value string) {
	env, _ := godotenv.Read(filepath.Join(ExecDir, ".env.local"))
	env[key] = value
	godotenv.Write(env, filepath.Join(ExecDir, ".env.local"))

	os.Setenv(key, value)
}

// GetenvOrSetDefault writes to .env if env not exists
func GetenvOrSetDefault(key string, value string) string {
	v := os.Getenv(key)
	if v == "" {
		v = value
		Setenv(key, v)
	}

	return v
}

// GetenvOrSetDefaultFn writes to .env if env not exists, using function
func GetenvOrSetDefaultFn(key string, fn func() string) string {
	v := os.Getenv(key)
	if v == "" {
		v = fn()
		Setenv(key, v)
	}

	return v
}
