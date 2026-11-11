package datasphere_project_test

import (
	dataspheresdk "github.com/yandex-cloud/go-sdk/services/datasphere/v2"
	framework "github.com/yandex-cloud/terraform-provider-yandex/internal/provider/framework"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	fixtures "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/datasphere/testfixtures"
)

func projectClient() fixtures.ProjectClient {
	return dataspheresdk.NewProjectClient(test.AccProvider.(*framework.Provider).GetConfig().SDKv2)
}
