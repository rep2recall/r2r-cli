package shared

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

var ExecDir string

// Compile with
// go build -ldflags "-X shared.UserDataDir=$USER_DATA_DIR"
var UserDataDir string

type SegmenterStruct struct {
	Command []string
}

type ProxyStruct struct {
	Port    int
	Command []string
}

type ConfigStruct struct {
	DB        string
	Port      int
	Secret    string
	Segmenter map[string]SegmenterStruct // map[Lang]SegmenterStruct
	Proxy     map[string]ProxyStruct     // map[Path]ProxyStruct
}

var Config ConfigStruct

func init() {
	ExecDir, _ = filepath.Abs(filepath.Dir(os.Args[0]))

	if UserDataDir == "" {
		UserDataDir = ExecDir
	}

	if _, e := os.Stat(filepath.Join(UserDataDir, "config.yaml")); e == nil {
		b, e := ioutil.ReadFile(filepath.Join(UserDataDir, "config.yaml"))
		if e != nil {
			Fatalln(e)
		}

		if e := yaml.Unmarshal(b, &Config); e != nil {
			Fatalln(e)
		}
	}

	if Config.DB == "" {
		Config.DB = "data.db"
	}

	if Config.Port == 0 {
		Config.Port = 25459
	}

	if Config.Secret == "" {
		s, e := GenerateRandomString(32)
		if e != nil {
			Fatalln(e)
		}
		Config.Secret = s
	}

	b, e := yaml.Marshal(&Config)
	if e != nil {
		Fatalln(e)
	}
	e = ioutil.WriteFile(filepath.Join(UserDataDir, "config.yaml"), b, 0644)
	if e != nil {
		Fatalln(e)
	}
}
