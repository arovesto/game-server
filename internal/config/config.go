package config

import (
	"github.com/pelletier/go-toml"
	"io/ioutil"
)

type Config struct {
	General GeneralConfig
}

type GeneralConfig struct {
	Address string
}

var cfg *Config

func GetConfig() *Config {
	if cfg == nil {
		panic("Config is not set")
	}
	return cfg
}

func LoadConfig(file string) (*Config, error) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	cfg = &Config{}
	return cfg, toml.Unmarshal(content, cfg)
}