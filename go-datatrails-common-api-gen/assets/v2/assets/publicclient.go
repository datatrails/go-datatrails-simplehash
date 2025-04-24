package assets

import (
	"github.com/datatrails/go-datatrails-common/grpcclient"
)

type PublicClient struct {
	PublicAssetsClient
	g *grpcclient.Client
}

func NewPublicClient(log Logger, address string, opts ...ClientOption) *PublicClient {
	return &PublicClient{
		g: grpcclient.New(log, "publicassetsv2", address, opts...),
	}
}

func (c *PublicClient) Open() error {
	if c.PublicAssetsClient != nil {
		return nil
	}
	err := c.g.Open()
	if err != nil {
		return err
	}
	c.PublicAssetsClient = NewPublicAssetsClient(c.g.Connector())
	return nil
}

func (c *PublicClient) Close() {
	c.g.Close()
}
