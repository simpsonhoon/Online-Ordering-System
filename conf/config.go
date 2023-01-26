package conf

import (
	"os"

	"github.com/naoina/toml"
)

type Log struct {
	Level   string
	Fpath   string
	Msize   int
	Mage    int
	Mbackup int
}

type Config struct {
	Server struct {
		Mode string
		Port string
	}
	Log
	DB map[string]map[string]interface{}
}

func GetConfig(fpath string) *Config {
	c := new(Config)

	if file, err := os.Open(fpath); err != nil {
		panic(err)
	} else {
		defer file.Close()
		//toml 파일 디코딩
		if err := toml.NewDecoder(file).Decode(c); err != nil {
			panic(err)
		}
	}
	return c
}
