package argocd

import "context"

type Client struct {}

func New() *Client { return &Client{} }

func (c *Client) Ping(ctx context.Context) error {
	// TODO: implement argocd adapter connectivity check.
	return nil
}
