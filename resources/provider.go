package resources

import (
	"github.com/cloudquery/cq-provider-sdk/provider"
	"github.com/cloudquery/cq-provider-sdk/provider/schema"
	"github.com/cloudquery/cq-provider-terraform/client"
)

func Provider() *provider.Provider {
	return &provider.Provider{
		Name:      "aws",
		Configure: client.Configure,
		ResourceMap: map[string]*schema.Table{
		},
		Config: func() provider.Config {
			return &client.Config{}
		},
	}
}
