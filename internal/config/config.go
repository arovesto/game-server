package config

import (
	"fmt"
	"io/ioutil"

	"github.com/pelletier/go-toml"
)

const (
	LogDebug = "debug"
	LogProd  = "prod"

	MaxClientsCap = 10000
)

type Config struct {
	General GeneralConfig
	Logging LoggingConfig
}

func (c *Config) Validate() error {
	if err := c.General.Validate(); err != nil {
		return err
	}
	if err := c.Logging.Validate(); err != nil {
		return err
	}
	return nil
}

type GeneralConfig struct {
	Address    string
	ClientsCap int
}

func (c *GeneralConfig) Validate() error {
	if c.ClientsCap > 0 && c.ClientsCap <= MaxClientsCap {
		return nil
	}
	return fmt.Errorf("%d not in (0, %d]", c.ClientsCap, MaxClientsCap)
}

type LoggingConfig struct {
	Type string
}

func (c *LoggingConfig) Validate() error {
	if c.Type == LogDebug || c.Type == LogProd {
		return nil
	}
	return fmt.Errorf("bad logger format: %s", c.Type)
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
	if err = toml.Unmarshal(content, cfg); err != nil {
		return nil, err
	}
	return cfg, cfg.Validate()
}
