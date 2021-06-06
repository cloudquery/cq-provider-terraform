package client

import (
	"github.com/cloudquery/cq-provider-sdk/provider/schema"
	"github.com/hashicorp/go-hclog"
)

func Configure(logger hclog.Logger, providerConfig interface{}) (schema.ClientMeta, error) {
	//ctx := context.Background()
	terraformConfig := providerConfig.(*Config)

	client, err := NewBackend(terraformConfig)
	if err != nil {
		return nil, err
	}

	client.Init()

	return nil, nil
}
