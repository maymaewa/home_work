package main

import "testing"

func TestNewConfig(t *testing.T) {
	config, err := NewConfig("../../configs/config.toml")
	if err != nil {
		t.Fatalf("NewConfig() error = %v", err)
	}

	if config.Logger.Level != "debug" {
		t.Errorf("Logger.Level = %q, want %q", config.Logger.Level, "debug")
	}

	if config.Storage.Type != "memory" {
		t.Errorf("Storage.Type = %q, want %q", config.Storage.Type, "memory")
	}

	if config.HTTP.Host != "127.0.0.1" {
		t.Errorf("HTTP.Host = %q, want %q", config.HTTP.Host, "127.0.0.1")
	}

	if config.HTTP.Port != 8080 {
		t.Errorf("HTTP.Port = %d, want %d", config.HTTP.Port, 8080)
	}

	if config.Storage.SQL.Host != "localhost" {
		t.Errorf("SQL.Host = %q, want %q", config.Storage.SQL.Host, "localhost")
	}

	if config.Storage.SQL.Port != 5432 {
		t.Errorf("SQL.Port = %d, want %d", config.Storage.SQL.Port, 5432)
	}

	if config.Storage.SQL.Database != "calendar" {
		t.Errorf("SQL.Database = %q, want %q", config.Storage.SQL.Database, "calendar")
	}
}
