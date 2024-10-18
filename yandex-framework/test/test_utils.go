package test

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"google.golang.org/grpc/status"
)

func AccCheckCreatedAtAttr(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		const createdAtAttrName = "created_at"
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("can't find %s in state", resourceName)
		}

		createdAt, ok := rs.Primary.Attributes[createdAtAttrName]
		if !ok {
			return fmt.Errorf("can't find '%s' attr for %s resource", createdAtAttrName, resourceName)
		}

		if _, err := time.Parse(time.RFC3339, createdAt); err != nil {
			return fmt.Errorf("can't parse timestamp in attr '%s': %s", createdAtAttrName, createdAt)
		}
		return nil
	}
}

func ImportIamBindingIdFunc(resourceName, role string) func(*terraform.State) (string, error) {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("can't find %s in state", resourceName)
		}
		tflog.Error(context.Background(), rs.Primary.ID)
		return fmt.Sprintf("%s,%s", rs.Primary.ID, role), nil
	}
}

func AccCheckResourceIDField(resourceName string, idFieldName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		if rs.Primary.Attributes[idFieldName] != rs.Primary.ID {
			return fmt.Errorf("Resource: %s id field: %s, doesn't match resource ID", resourceName, idFieldName)
		}

		return nil
	}
}

func AccCheckResourceAttrWithValueFactory(name, key string, valueFactory func() string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return resource.TestCheckResourceAttr(name, key, valueFactory())(s)
	}
}

func ErrorMessage(err error) string {
	grpcStatus, _ := status.FromError(err)
	return grpcStatus.Message()
}
