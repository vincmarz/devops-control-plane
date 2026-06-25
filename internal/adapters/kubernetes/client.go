package kubernetes

import "context"

type Client struct {}

func New() *Client { return &Client{} }

func (c *Client) Ping(ctx context.Context) error {
	// TODO: implement kubernetes adapter connectivity check.
	return nil
}
