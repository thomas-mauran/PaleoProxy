package main

import (
	"fmt"
	"os"

	"github.com/go-yaml/yaml"
)

type Conf struct {
	Services []Service `yaml:"services"`
}

type Service struct {
	Name string `yaml:"name"`
	Description string `yaml:"description"`
	Enabled bool `yaml:"enabled"`
	Subdomain string `yaml:"subdomain"`
	Endpoints []Endpoint `yaml:"endpoints"`
	Port int64 `yaml:"port"`
}

type Endpoint struct {
	Ip string `yaml:"ip"`
}


func ReadConfig(filename string) (*Conf, error) {
	buf, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	c := &Conf{}
	err = yaml.Unmarshal(buf, c)
	if err != nil {
		return nil, fmt.Errorf("in file %q: %w", filename, err)
	}
	return c, nil
}