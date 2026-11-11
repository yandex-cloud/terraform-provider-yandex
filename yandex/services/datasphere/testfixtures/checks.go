package testfixtures

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/datasphere/v2"
	"google.golang.org/grpc"
)

const (
	CommunityResourceName = "yandex_datasphere_community.test-community"
	ProjectResourceName   = "yandex_datasphere_project.test-project"
)

func ProjectExists(client func() ProjectClient, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		id := rs.Primary.ID

		found, err := client().Get(context.Background(), &datasphere.GetProjectRequest{
			ProjectId: id,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("project not found")
		}

		return nil
	}
}

func AccCheckProjectDestroy(client func() ProjectClient) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "yandex_datasphere_project" {
				continue
			}
			id := rs.Primary.ID

			_, err := client().Get(context.Background(), &datasphere.GetProjectRequest{
				ProjectId: id,
			})
			if err == nil {
				return fmt.Errorf("project still exists")
			}
		}

		return nil
	}
}

func CommunityExists(client func() CommunityClient, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		id := rs.Primary.ID

		found, err := client().Get(context.Background(), &datasphere.GetCommunityRequest{
			CommunityId: id,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("community not found")
		}

		return nil
	}
}

func AccCheckCommunityDestroy(client func() CommunityClient) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "yandex_datasphere_community" {
				continue
			}
			id := rs.Primary.ID

			_, err := client().Get(context.Background(), &datasphere.GetCommunityRequest{
				CommunityId: id,
			})
			if err == nil {
				return fmt.Errorf("community still exists")
			}
		}

		return nil
	}
}

type ProjectClient interface {
	Get(context.Context, *datasphere.GetProjectRequest, ...grpc.CallOption) (*datasphere.Project, error)
}
type CommunityClient interface {
	Get(context.Context, *datasphere.GetCommunityRequest, ...grpc.CallOption) (*datasphere.Community, error)
}
