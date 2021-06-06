package client

import "github.com/hashicorp/hcl/v2"

type Config struct {
	Backend BackendType `hcl:"backend"`
	Config  *hcl.Attribute `hcl:"config"`
}

func (c Config) Example() string {
	return `configuration {

	// local backend
    backend = "local"
	config = {
		path = "/path/to/tfstate/file"
	}
	// s3 backend
	backend = "s3"
	config = {
		bucket = "terraform-state-prod"
		key    = "network/terraform.tfstate"
		region = "us-east-1"
	}
}
`
}
