package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: go-envdir envdir command [args...]")
		os.Exit(1)
	}

	env, err := ReadDir(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	returnCode := RunCmd(os.Args[2:], env)

	os.Exit(returnCode)
}
