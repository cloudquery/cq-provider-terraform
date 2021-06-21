package client

import (
	"fmt"

	"errors"

	"github.com/cloudquery/cq-provider-sdk/provider/schema"
	"github.com/hashicorp/go-hclog"
)

type Client struct {
	Backends map[string]*TerraformBackend
	logger   hclog.Logger

	CurrentBackend string
}

func NewTerraformClient(logger hclog.Logger, backends map[string]*TerraformBackend) Client {
	return Client{
		Backends: backends,
		logger:   logger,
	}
}

func (c *Client) Logger() hclog.Logger {
	return c.logger
}

func Configure(logger hclog.Logger, providerConfig interface{}) (schema.ClientMeta, error) {
	terraformConfig := providerConfig.(*Config)

	if terraformConfig.Config == nil || len(terraformConfig.Config) == 0 {
		return nil, errors.New("no config were provided")
	}

	var backends = make(map[string]*TerraformBackend)
	for _, config := range terraformConfig.Config {
		if b, err := NewBackend(&config); err == nil {
			backends[b.BackendName] = b
		} else {
			return nil, fmt.Errorf("cannot load backend, %s", err)
		}
	}

	client := NewTerraformClient(logger, backends)

	return &client, nil
}

func (c *Client) Backend() *TerraformBackend {
	if c.CurrentBackend != "" {
		backend := c.Backends[c.CurrentBackend]
		return backend
	}
	for _, backend := range c.Backends {
		return backend
	}
	return nil
}

func (c *Client) withSpecificBackend(backendName string) *Client {
	return &Client{
		Backends:       c.Backends,
		logger:         c.logger,
		CurrentBackend: backendName,
	}
}
