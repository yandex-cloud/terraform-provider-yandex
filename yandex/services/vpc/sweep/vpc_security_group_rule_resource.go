package sweep

import (
	"context"
	"fmt"
	"time"

	vpcsdk "github.com/yandex-cloud/go-sdk/services/vpc/v1"

	"github.com/hashicorp/go-multierror"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/internal/provider/framework/config"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
)

const YandexVPCNetworkDefaultTimeout = 1 * time.Minute

func registerVpcSecurityGroupRuleResource(add func(string, *resource.Sweeper)) {
	add("yandex_vpc_security_group_rule", &resource.Sweeper{
		Name: "yandex_vpc_security_group_rule",
		F:    sweepAllVPCSecurityGroupsV2,
	})
}

func sweepAllVPCSecurityGroupsV2(_ string) error {
	conf, err := test.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &vpc.ListSecurityGroupsRequest{FolderId: conf.ProviderState.FolderID.ValueString()}
	it := vpcsdk.NewSecurityGroupClient(conf.SDKv2).Iterator(context.Background(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepVPCSecurityGroupV2(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep VPC security group %q", it.Value().GetId()))
		}
	}

	return result.ErrorOrNil()
}

func sweepVPCSecurityGroupV2(conf *provider_config.Config, id string) bool {
	return test.SweepWithRetry(sweepVPCSecurityGroupOnceV2, conf, "VPC Security Group", id)
}

func sweepVPCSecurityGroupOnceV2(conf *provider_config.Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), YandexVPCNetworkDefaultTimeout)
	defer cancel()

	sg, err := vpcsdk.NewSecurityGroupClient(conf.SDKv2).Get(ctx, &vpc.GetSecurityGroupRequest{
		SecurityGroupId: id,
	})
	if err != nil {
		return err
	}

	if sg.DefaultForNetwork {
		return nil
	}

	op, err := vpcsdk.NewSecurityGroupClient(conf.SDKv2).Delete(ctx, &vpc.DeleteSecurityGroupRequest{
		SecurityGroupId: id,
	})
	return test.HandleSweepOperation(ctx, op, err)
}
