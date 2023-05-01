package informant

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type conf struct {
	Bind       string `yaml:"bind"`
	Homeserver string `yaml:"homeserver"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
	Display    string `yaml:"display"`
	Avatar     string `yaml:"avatar"`
	Database   struct {
		Path string `yaml:"path"`
		Key  string `yaml:"key"`
	} `yaml:"database"`
	Debug bool   `yaml:"debug"`
	PSK   string `yaml:"psk"`
}

func ReadConfig(path string) (*conf, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error: failed to read yaml config: %w", err)
	}

	c := conf{}
	err = yaml.Unmarshal(buf, &c)
	if err != nil {
		return nil, fmt.Errorf("error: failed to unmarshal yaml config: %w", err)
	}

	return &c, nil
}
