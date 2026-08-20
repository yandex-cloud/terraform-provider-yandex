package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
	containerregistrysdk "github.com/yandex-cloud/go-sdk/services/containerregistry/v1"
)

type containerRepositoryByNameClient interface {
	GetByName(context.Context, *containerregistry.GetRepositoryByNameRequest, ...grpc.CallOption) (*containerregistry.Repository, error)
}

func getContainerRepositoryIDByName(ctx context.Context, client containerRepositoryByNameClient, name string) (string, error) {
	repository, err := client.GetByName(ctx, &containerregistry.GetRepositoryByNameRequest{
		RepositoryName: name,
	})
	if err != nil {
		return "", err
	}
	if repository == nil || repository.GetId() == "" {
		return "", fmt.Errorf("repository %q returned without an ID", name)
	}
	return repository.GetId(), nil
}

func dataSourceYandexContainerRepository() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a Yandex Container Repository. For more information, see [the official documentation](https://yandex.cloud/docs/container-registry/concepts/repository).",

		Read: dataSourceYandexContainerRepositoryRead,
		Schema: map[string]*schema.Schema{
			"repository_id": {
				Type:        schema.TypeString,
				Description: "The ID of a specific repository.",
				Optional:    true,
				Computed:    true,
			},

			"name": {
				Type:        schema.TypeString,
				Description: "A name of the repository. The name of the repository should start with id of a container registry and match the name of the images that will be pushed in the repository.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func dataSourceYandexContainerRepositoryRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()
	client := containerregistrysdk.NewRepositoryClient(config.SDK)

	err := checkOneOf(d, "repository_id", "name")
	if err != nil {
		return err
	}

	repositoryID := d.Get("repository_id").(string)
	_, repositoryNameOk := d.GetOk("name")

	if repositoryNameOk {
		repositoryID, err = getContainerRepositoryIDByName(ctx, client, d.Get("name").(string))
		if err != nil {
			return fmt.Errorf("failed to resolve data source Сontainer Repository by name: %v", err)
		}
	}

	repository, err := client.Get(ctx,
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
