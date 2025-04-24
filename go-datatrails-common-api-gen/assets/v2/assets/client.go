package assets

import (
	"github.com/datatrails/go-datatrails-common/grpcclient"
)

type AssetsClientCloser interface {
	AssetsClient
	Open() error
	Close()
}

type ClientOption = grpcclient.ClientOption

type Client struct {
	AssetsClient
	g *grpcclient.Client
}

func NewClient(log Logger, address string, opts ...ClientOption) *Client {
	return &Client{
		g: grpcclient.New(log, "assetsv2", address, opts...),
	}
}

func (c *Client) Open() error {
	if c.AssetsClient != nil {
		return nil
	}
	err := c.g.Open()
	if err != nil {
		return err
	}
	c.AssetsClient = NewAssetsClient(c.g.Connector())
	return nil
}

func (c *Client) Close() {
	c.g.Close()
}
