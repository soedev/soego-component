package emongo

import (
	"github.com/soedev/soego/core/elog"
)

const PackageName = "component.emongo"

// Component client (cmdable and config)
type Component struct {
	config *config
	client *Client
	logger *elog.Component
}

// Client returns emongo Client
func (c *Component) Client() *Client {
	return c.client
}
