package main

import "os"
import "encoding/json"

type Config struct {
	Logging      bool
	Port         int
	DatabaseFile string
}

func parseConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(file)

	cfg := &Config{}
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
