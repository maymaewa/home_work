package main

import (
	"bytes"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTelnetClient(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:")
		require.NoError(t, err)
		defer func() { require.NoError(t, l.Close()) }()

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()

			in := &bytes.Buffer{}
			out := &bytes.Buffer{}

			timeout, err := time.ParseDuration("10s")
			require.NoError(t, err)

			client := NewTelnetClient(l.Addr().String(), timeout, io.NopCloser(in), out)
			require.NoError(t, client.Connect())
			defer func() { require.NoError(t, client.Close()) }()

			in.WriteString("hello\n")
			err = client.Send()
			require.NoError(t, err)

			err = client.Receive()
			require.NoError(t, err)
			require.Equal(t, "world\n", out.String())
		}()

		go func() {
			defer wg.Done()

			conn, err := l.Accept()
			require.NoError(t, err)
			require.NotNil(t, conn)
			defer func() { require.NoError(t, conn.Close()) }()

			request := make([]byte, 1024)
			n, err := conn.Read(request)
			require.NoError(t, err)
			require.Equal(t, "hello\n", string(request)[:n])

			n, err = conn.Write([]byte("world\n"))
			require.NoError(t, err)
			require.NotEqual(t, 0, n)
		}()

		wg.Wait()
	})

	t.Run("send returns EOF", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:")
		require.NoError(t, err)
		defer func() { require.NoError(t, l.Close()) }()

		connCh := make(chan net.Conn)

		go func() {
			conn, err := l.Accept()
			require.NoError(t, err)
			connCh <- conn
		}()

		in := bytes.NewBufferString("hello")
		out := &bytes.Buffer{}

		client := NewTelnetClient(
			l.Addr().String(),
			time.Second,
			io.NopCloser(in),
			out,
		)

		require.NoError(t, client.Connect())
		defer func() { require.NoError(t, client.Close()) }()

		// Первый вызов отправляет данные.
		err = client.Send()
		require.NoError(t, err)

		// Второй вызов получает EOF от входного потока.
		err = client.Send()
		require.ErrorIs(t, err, io.EOF)

		conn := <-connCh
		defer func() { require.NoError(t, conn.Close()) }()

		buf := make([]byte, 100)
		n, err := conn.Read(buf)
		require.NoError(t, err)
		require.Equal(t, "hello", string(buf[:n]))
	})

	t.Run("connection refused", func(t *testing.T) {
		client := NewTelnetClient(
			"127.0.0.1:1",
			100*time.Millisecond,
			io.NopCloser(bytes.NewBuffer(nil)),
			&bytes.Buffer{},
		)

		err := client.Connect()
		require.Error(t, err)
	})
}
