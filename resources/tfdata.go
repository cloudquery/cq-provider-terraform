package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/cloudquery/cq-provider-sdk/provider/schema"
	"github.com/cloudquery/cq-provider-terraform/client"
)

var providerNameRegex = regexp.MustCompile(`^.*\["(?P<Hostname>.*)/(?P<Namespace>.*)/(?P<Type>.*)"\].*?$`)

func TFData() *schema.Table {
	return &schema.Table{
		Name:         "tf_data",
		Resolver:     resolveTerraformMetaData,
		DeleteFilter: client.DeleteLineageFilter,
		Multiplex:    client.BackendMultiplex,
		Columns: []schema.Column{
			{
				Name:     "backend",
				Type:     schema.TypeString,
				Resolver: resolveBackend,
			},
			{
				Name:     "backend_name",
				Type:     schema.TypeString,
				Resolver: resolveBackendName,
			},
			{
				Name: "version",
				Type: schema.TypeBigInt,
			},
			{
				Name: "terraform_version",
				Type: schema.TypeString,
			},
			{
				Name: "serial",
				Type: schema.TypeBigInt,
			},
			{
				Name: "lineage",
				Type: schema.TypeString,
			},
		},
		Relations: []*schema.Table{
			{
				Name:     "tf_resources",
				Resolver: resolveTerraformResources,
				Columns: []schema.Column{
					{
						Name:     "running_id",
						Type:     schema.TypeUUID,
						Resolver: schema.ParentIdResolver,
					},
					{
						Name: "module",
						Type: schema.TypeString,
					},
					{
						Name: "mode",
						Type: schema.TypeString,
					},
					{
						Name: "type",
						Type: schema.TypeString,
					},
					{
						Name: "name",
						Type: schema.TypeString,
					},
					{
						Name:     "provider_path",
						Type:     schema.TypeString,
						Resolver: schema.PathResolver("ProviderConfig"),
					},
					{
						Name:     "provider",
						Type:     schema.TypeString,
						Resolver: resolveProviderName,
					},
				},
				Relations: []*schema.Table{
					{
						Name:     "tf_resource_instances",
						Resolver: resolveTerraformResourceInstances,
						Columns: []schema.Column{
							{
								Name:     "resource_id",
								Type:     schema.TypeUUID,
								Resolver: schema.ParentIdResolver,
							},
							{
								Name:     "internal_id",
								Type:     schema.TypeString,
								Resolver: resolveInstanceInternalId,
							},
							{
								Name: "schema_version",
								Type: schema.TypeBigInt,
							},
							{
								Name:     "attribute",
								Type:     schema.TypeJSON,
								Resolver: resolveInstanceAttributes,
							},
							{
								Name: "dependencies",
								Type: schema.TypeStringArray,
							},
							{
								Name: "create_before_destroy",
								Type: schema.TypeBool,
							},
						},
					},
				},
			},
		},
	}
}

// ====================================================================================================================
//                                               Table Resolver Functions
// ====================================================================================================================
func resolveTerraformMetaData(_ context.Context, meta schema.ClientMeta, _ *schema.Resource, res chan interface{}) error {
	c := meta.(*client.Client)
	backend := c.Backend()
	res <- backend.Data().State
	return nil
}

func resolveBackend(_ context.Context, meta schema.ClientMeta, resource *schema.Resource, _ schema.Column) error {
	c := meta.(*client.Client)
	backend := c.Backend()
	return resource.Set("backend", backend.Type())
}

func resolveBackendName(_ context.Context, meta schema.ClientMeta, resource *schema.Resource, _ schema.Column) error {
	c := meta.(*client.Client)
	backend := c.Backend()
	return resource.Set("backend_name", backend.Name())
}

func resolveTerraformResources(_ context.Context, _ schema.ClientMeta, parent *schema.Resource, res chan interface{}) error {
	state, ok := parent.Item.(client.State)
	if !ok {
		return fmt.Errorf("not terraform state")
	}
	for _, resource := range state.Resources {
		res <- resource
	}
	return nil
}

func resolveTerraformResourceInstances(_ context.Context, _ schema.ClientMeta, parent *schema.Resource, res chan interface{}) error {
	resource, ok := parent.Item.(client.Resource)
	if !ok {
		return fmt.Errorf("not terraform Resource")
	}
	for _, instance := range resource.Instances {
		res <- instance
	}
	return nil
}

func resolveProviderName(_ context.Context, _ schema.ClientMeta, resource *schema.Resource, c schema.Column) error {
	res, ok := resource.Item.(client.Resource)
	if !ok {
		return fmt.Errorf("not terraform Resource")
	}

	matches := providerNameRegex.FindStringSubmatch(res.ProviderConfig)
	typeIndex := providerNameRegex.SubexpIndex("Type")
	if len(matches) >= 3 {
		return resource.Set(c.Name, matches[typeIndex])
	}
	return nil
}

func resolveInstanceAttributes(_ context.Context, _ schema.ClientMeta, resource *schema.Resource, c schema.Column) error {
	instance, ok := resource.Item.(client.Instance)
	if !ok {
		return fmt.Errorf("not terraform Instance")
	}
	attrs, err := instance.AttributesRaw.MarshalJSON()
	if err != nil {
		return fmt.Errorf("not valid JSON attributes")
	}
	return resource.Set(c.Name, attrs)
}

func resolveInstanceInternalId(_ context.Context, _ schema.ClientMeta, resource *schema.Resource, c schema.Column) error {
	instance := resource.Item.(client.Instance)
	data := make(map[string]interface{})
	if err := json.Unmarshal(instance.AttributesRaw, &data); err != nil {
		return nil
	}
	if val, ok := data["id"]; ok {
		return resource.Set(c.Name, val)
	}
	return nil
}
