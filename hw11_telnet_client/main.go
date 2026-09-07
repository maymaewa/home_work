package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	timeout := flag.Duration("timeout", 10*time.Second, "connection timeout")
	flag.Parse()

	if flag.NArg() != 2 {
		fmt.Fprintln(os.Stderr, "usage: go-telnet [--timeout=10s] host port")
		os.Exit(1)
	}

	address := net.JoinHostPort(flag.Arg(0), flag.Arg(1))

	client := NewTelnetClient(address, *timeout, os.Stdin, os.Stdout)

	if err := client.Connect(); err != nil {
		fmt.Fprintln(os.Stderr, "connection error:", err)
		os.Exit(1)
	}
	defer client.Close()

	fmt.Fprintf(os.Stderr, "Connected to %s\n", address)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	errCh := make(chan error, 2)

	go func() {
		for {
			if err := client.Send(); err != nil {
				errCh <- err
				return
			}
		}
	}()

	go func() {
		errCh <- client.Receive()
	}()

	select {
	case <-ctx.Done():
		_ = client.Close()
	case err := <-errCh:
		if errors.Is(err, io.EOF) {
			fmt.Fprintln(os.Stderr, "EOF")
		} else {
			fmt.Fprintln(os.Stderr, "Connection was closed by peer:", err)
		}
	}
}
