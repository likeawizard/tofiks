package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

const CONFIG_PATH = "config.yml"

type Config struct {
	Init   Init   `yaml:"init"`
	Engine Engine `yaml:"engine"`
	Book   Book   `yaml:"book"`
	Render Render `yaml:"render"`
}

type Init struct {
	StartingFen string `yaml:"startingFen"`
}

type Engine struct {
	Algorithm     string `yaml:"algorithm"`
	MaxDepth      int    `yaml:"maxDepth"`
	EnableTT      bool   `yaml:"enableTT"`
	TTSize        int    `yaml:"ttSize"`
	MaxGoRoutines int    `yaml:"maxGoRoutines"`
	Debug         bool   `yaml:"debug"`
}

type Book struct {
	Enable bool `yaml:"enable"`
	Method string
	Path   string `yaml:"path"`
}

type Render struct {
	Mode string `yaml:"mode"`
}

func LoadConfig() (*Config, error) {
	var cfg Config
	cfgFile, err := os.Open(CONFIG_PATH)
	if err != nil {
		return nil, err
	}
	defer cfgFile.Close()

	d := yaml.NewDecoder(cfgFile)
	err = d.Decode(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
