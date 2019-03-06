package main

import (
	"github.com/go-yaml/yaml"
	"github.com/kovetskiy/ko"
)

type Config struct {
	Picker       []string `yaml:"picker" required:"true"`
	Threads      int      `yaml:"threads" required:"true"`
	VotesPath    string   `yaml:"votes_path" required:"true"`
	IgnoreGlobal []string `yaml:"ignore_global"`
	Trees        Trees    `yaml:"trees" required:"true"`
}

func LoadConfig(path string) (*Config, error) {
	var config Config
	err := ko.Load(path, &config, yaml.Unmarshal)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
