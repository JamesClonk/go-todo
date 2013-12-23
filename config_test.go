package main

import "testing"

func Test_config_parseConfig(t *testing.T) {
	cfg, err := parseConfig("go-todo.json")
	if err != nil {
		t.Error(err)
		return
	}

	if cfg.Port != 8008 || cfg.Logging != true || cfg.DatabaseFile != "data/tasks.db" {
		t.Errorf("Configfile was not as expected: [%v]", cfg)
	}
}
