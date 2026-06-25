package tekton

import "context"

type Client struct {}

func New() *Client { return &Client{} }

func (c *Client) Ping(ctx context.Context) error {
	// TODO: implement tekton adapter connectivity check.
	return nil
}
