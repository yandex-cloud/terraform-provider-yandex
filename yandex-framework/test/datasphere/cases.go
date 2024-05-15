package dataspheretest

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/datasphere/v2"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/test"
)

const (
	CommunityResourceName = "yandex_datasphere_community.test-community"
	ProjectResourceName   = "yandex_datasphere_project.test-project"
)

func ProjectExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		id := rs.Primary.ID

		found, err := config.SDK.Datasphere().Project().Get(context.Background(), &datasphere.GetProjectRequest{
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

func AccCheckProjectDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_project_project" {
			continue
		}
		id := rs.Primary.ID

		_, err := config.SDK.Datasphere().Project().Get(context.Background(), &datasphere.GetProjectRequest{
			ProjectId: id,
		})
		if err == nil {
			return fmt.Errorf("project still exists")
		}
	}

	return nil
}

func CommunityExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()
		a := s.RootModule().Resources
		fmt.Printf("%s", a)
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		id := rs.Primary.ID

		found, err := config.SDK.Datasphere().Community().Get(context.Background(), &datasphere.GetCommunityRequest{
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

func AccCheckCommunityDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_datasphere_community" {
			continue
		}
		id := rs.Primary.ID

		_, err := config.SDK.Datasphere().Community().Get(context.Background(), &datasphere.GetCommunityRequest{
			CommunityId: id,
		})
		if err == nil {
			return fmt.Errorf("community still exists")
		}
	}

	return nil
}
