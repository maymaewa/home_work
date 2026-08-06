package main

import (
	"os"
	"os/exec"
	"strings"
)

func RunCmd(cmd []string, env Environment) (returnCode int) {
	if len(cmd) == 0 {
		return 1
	}

	current := make(map[string]string)

	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", 2)

		if len(parts) == 2 {
			current[parts[0]] = parts[1]
		}
	}

	for key, value := range env {
		if value.NeedRemove {
			delete(current, key)
			continue
		}

		current[key] = value.Value
	}

	envList := make([]string, 0, len(current))

	for k, v := range current {
		envList = append(envList, k+"="+v)
	}

	command := exec.Command(cmd[0], cmd[1:]...)
	command.Env = envList

	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	err := command.Run()
	if err == nil {
		return 0
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.ExitCode()
	}

	return 1
}