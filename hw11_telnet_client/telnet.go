package main

import (
	"io"
	"net"
	"time"
)

type TelnetClient interface {
	Connect() error
	io.Closer
	Send() error
	Receive() error
}

type telnetClient struct {
	address string
	timeout time.Duration
	in      io.Reader
	out     io.Writer
	conn    net.Conn
}

func NewTelnetClient(address string, timeout time.Duration, in io.ReadCloser, out io.Writer) TelnetClient {
	return &telnetClient{
		address: address,
		timeout: timeout,
		in:      in,
		out:     out,
	}
}

func (c *telnetClient) Connect() error {
	conn, err := net.DialTimeout("tcp", c.address, c.timeout)
	if err != nil {
		return err
	}

	c.conn = conn
	return nil
}

func (c *telnetClient) Close() error {
	if c.conn == nil {
		return nil
	}

	return c.conn.Close()
}

func (c *telnetClient) Send() error {
	buf := make([]byte, 4096)

	n, err := c.in.Read(buf)

	if n > 0 {
		if _, writeErr := c.conn.Write(buf[:n]); writeErr != nil {
			return writeErr
		}
	}

	return err
}

func (c *telnetClient) Receive() error {
	_, err := io.Copy(c.out, c.conn)
	return err
}
