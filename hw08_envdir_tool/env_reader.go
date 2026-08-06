package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Environment map[string]EnvValue

type EnvValue struct {
	Value      string
	NeedRemove bool
}

func ReadDir(dir string) (Environment, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	env := make(Environment)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		if strings.Contains(name, "=") {
			return nil, fmt.Errorf("Неверное название переменной: %s", name)
		}

		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}

		if len(data) == 0 {
			env[name] = EnvValue{
				NeedRemove: true,
			}
			continue
		}

		if idx := bytes.IndexByte(data, '\n'); idx >= 0 {
			data = data[:idx]
		}

		data = bytes.ReplaceAll(data, []byte{0}, []byte{'\n'})

		value := strings.TrimRight(string(data), " \t")

		env[name] = EnvValue{
			Value: value,
		}
	}

	return env, nil
}