package testhelpers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"google.golang.org/protobuf/types/known/durationpb"
	"strconv"
	"text/template"
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
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		if rs.Primary.Attributes[idFieldName] != rs.Primary.ID {
			return fmt.Errorf("resource: %s id field: %s, doesn't match resource ID", resourceName, idFieldName)
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

func TemplateConfig(tmpl string, ctx ...map[string]interface{}) string {
	p := make(map[string]interface{})
	for _, c := range ctx {
		for k, v := range c {
			p[k] = v
		}
	}
	b := &bytes.Buffer{}
	err := template.Must(template.New("").Parse(tmpl)).Execute(b, p)
	if err != nil {
		panic(fmt.Errorf("failed to execute config template: %v", err))
	}
	return b.String()
}

func CheckImportFolderID(folderID string) resource.ImportStateCheckFunc {
	return func(s []*terraform.InstanceState) error {
		if len(s) == 0 {
			return errors.New("No InstanceState found")
		}

		if len(s) != 1 {
			return fmt.Errorf("Expected one InstanceState, found: %d", len(s))
		}

		fID := s[0].Attributes["folder_id"]
		if fID != folderID {
			return fmt.Errorf("Expected folder_id %q, got %q", folderID, fID)
		}

		return nil
	}
}

func AccCheckBoolValue(resourceName, attributePath string, expectedValue bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("can't find %s in state", resourceName)
		}

		actualValue, ok := rs.Primary.Attributes[attributePath]
		if !ok {
			return fmt.Errorf("can't find '%s' attr for %s resource", attributePath, resourceName)
		}

		parseBool, err := strconv.ParseBool(actualValue)
		if err != nil {
			return err
		}
		if !parseBool == expectedValue {
			return fmt.Errorf("stored value: '%t' doesn't match expected value: '%t'", parseBool, expectedValue)
		}

		return nil
	}
}

func AccCheckDuration(resourceName string, attributePath string, expectedValue string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("can't find %s in state", resourceName)
		}

		startTime, ok := rs.Primary.Attributes[attributePath]
		if !ok {
			return fmt.Errorf("can't find '%s' attr for %s resource", attributePath, resourceName)
		}

		if !shouldSuppressDiffForTimeDuration(expectedValue, startTime) {
			return fmt.Errorf("stored value: '%s' doesn't match expected value: '%s'", startTime, expectedValue)
		}

		return nil
	}
}

func shouldSuppressDiffForTimeDuration(old, new string) bool {
	oldD, err := time.ParseDuration(old)
	if err != nil {
		return false
	}
	d1 := durationpb.New(oldD)

	newD, err := time.ParseDuration(new)
	if err != nil {
		return false
	}
	d2 := durationpb.New(newD)

	if d1 == nil && d2 == nil {
		return true
	}

	if d1 != nil && d2 != nil {
		return d1.Seconds == d2.Seconds && d1.Nanos == d2.Nanos
	}

	return false
}
