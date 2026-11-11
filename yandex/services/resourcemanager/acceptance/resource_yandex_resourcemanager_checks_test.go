package acceptance

import (
	"testing"

	svc_resource_manager "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/resourcemanager/pluginsdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

func memberType(ab *access.AccessBinding) string {
	return ab.Subject.Type
}

func memberID(ab *access.AccessBinding) string {
	return ab.Subject.Id
}

func userAccountIDs(p *sdkcommon.Policy) []string {
	usersMap := map[string]bool{}
	for _, b := range p.Bindings {
		if memberType(b) == "userAccount" {
			usersMap[memberID(b)] = true
		}
	}
	userIDs := []string{}
	for userID := range usersMap {
		userIDs = append(userIDs, userID)
	}
	return userIDs
}

func getFolderIamPolicyByFolderID(folderID string, config *providerconfig.Config) (*sdkcommon.Policy, error) {
	f, err := svc_resource_manager.NewFolderIamUpdaterFromFolderID(folderID, config)
	if err != nil {
		return nil, err
	}
	return f.GetResourceIamPolicy(config.Context())
}

func CustomProvidersTest(t *testing.T, providers []map[string]*schema.Provider, testCase resource.TestCase) {
	for _, provider := range providers {
		customTest := testCase
		customTest.Providers = provider
		resource.Test(t, customTest)
	}
}
