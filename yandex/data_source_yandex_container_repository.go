package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/grpc/codes"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexContainerRepository() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexContainerRepositoryRead,
		Schema: map[string]*schema.Schema{
			"repository_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceYandexContainerRepositoryRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()

	err := checkOneOf(d, "repository_id", "name")
	if err != nil {
		return err
	}

	repositoryID := d.Get("repository_id").(string)
	_, repositoryNameOk := d.GetOk("name")

	if repositoryNameOk {
		repositoryID, err = resolveObjectID(ctx, config, d, sdkresolvers.RepositoryResolver)
		if err != nil {
			return fmt.Errorf("failed to resolve data source Сontainer Repository by name: %v", err)
		}
	}

	repository, err := config.sdk.ContainerRegistry().Repository().Get(ctx,
		&containerregistry.GetRepositoryRequest{
			RepositoryId: repositoryID,
		})

	if err != nil {
		if isStatusWithCode(err, codes.NotFound) {
			return fmt.Errorf("Container Repository not found: %s", repositoryID)
		}
		return err
	}

	d.Set("repository_id", repository.Id)
	d.Set("name", repository.Name)

	d.SetId(repository.Id)

	return nil
}
