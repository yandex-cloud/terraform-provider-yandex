package yandex

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/containers/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexServerlessContainer() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexServerlessContainerRead,

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"container_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"folder_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"labels": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"memory": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Container memory in megabytes, should be aligned to 128",
			},

			"cores": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"core_fraction": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"execution_timeout": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"concurrency": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"service_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"image": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"work_dir": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"digest": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"command": {
							Type:     schema.TypeList,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Computed: true,
						},
						"args": {
							Type:     schema.TypeList,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Computed: true,
						},
						"environment": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
					},
				},
			},

			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"revision_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceYandexServerlessContainerRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	err := checkOneOf(d, "container_id", "name")
	if err != nil {
		return err
	}
	containerID := d.Get("container_id").(string)

	if _, ok := d.GetOk("name"); ok {
		containerID, err = resolveObjectID(ctx, config, d, sdkresolvers.ContainerResolver)
		if err != nil {
			return fmt.Errorf("failed to resolve data source Yandex Cloud Serverless Container by name: %v", err)
		}
	}

	req := containers.GetContainerRequest{
		ContainerId: containerID,
	}

	container, err := config.sdk.Serverless().Containers().Container().Get(ctx, &req)
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Yandex Cloud Container %q", d.Id()))
	}

	revision, err := resolveContainerLastRevision(ctx, config, containerID)
	if err != nil {
		return fmt.Errorf("Failed to resolve last revision of data source Yandex Cloud Container: %s", err)
	}

	d.SetId(container.Id)
	d.Set("container_id", container.Id)
	return flattenYandexServerlessContainer(d, container, revision)
}
