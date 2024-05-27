package client

import (
	"bytes"
	"context"
	"net"

	"github.com/tidwall/resp"
)

type Client struct {
	addr string
	conn net.Conn
}

func New(addr string) (*Client, error) {

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &Client{
		addr: addr,
		conn: conn,
	}, nil
}

// Set does the set request
func (c *Client) Set(ctx context.Context, key string, val string) error {

	var buf bytes.Buffer
	wr := resp.NewWriter(&buf)
	wr.WriteArray([]resp.Value{
		resp.StringValue("SET"),
		resp.StringValue(key),
		resp.StringValue(val)})

	_, err := c.conn.Write(buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {

	var buf bytes.Buffer
	wr := resp.NewWriter(&buf)
	wr.WriteArray([]resp.Value{
		resp.StringValue("GET"),
		resp.StringValue(key),
	})

	_, err := c.conn.Write(buf.Bytes())
	if err != nil {
		return "", err
	}

	// Read the result from the server
	b := make([]byte, 1024)
	n, err := c.conn.Read(b)

	return string(b[:n]), err

}

func (c *Client) Close() error {
	return c.conn.Close()
}
