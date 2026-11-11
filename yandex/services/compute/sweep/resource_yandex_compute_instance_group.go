package sweep

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1/instancegroup"
	instancegroupsdk "github.com/yandex-cloud/go-sdk/services/compute/v1/instancegroup"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	iamsweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/iam/sweep"
	svc_vpc "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/vpc/pluginsdk"
	vpcsweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/vpc/sweep"
)

func registerResourceYandexComputeInstanceGroup(add func(string, *resource.Sweeper)) {
	add("yandex_compute_instance_group", &resource.Sweeper{
		Name: "yandex_compute_instance_group",
		F:    sweepAllComputeInstanceGroups,
		Dependencies: []string{
			"yandex_kubernetes_node_group",
		},
	})
}

func sweepAllComputeInstanceGroups(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	var serviceAccountID, networkID, subnetID string
	var depsCreated bool

	req := &instancegroup.ListInstanceGroupsRequest{FolderId: conf.FolderID}
	it := instancegroupsdk.NewInstanceGroupClient(conf.SDK).Iterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		instanceGroup := it.Value()
		if !depsCreated {
			depsCreated = true
			serviceAccountID, err = iamsweep.CreateIAMServiceAccountForSweeper(conf)
			if err != nil {
				result = multierror.Append(result, err)
				break
			}
			networkID, err = vpcsweep.CreateVPCNetworkForSweeper(conf)
			if err != nil {
				result = multierror.Append(result, err)
				break
			}
			subnetID, err = vpcsweep.CreateVPCSubnetForSweeper(conf, networkID)
			if err != nil {
				result = multierror.Append(result, err)
				break
			}
		}

		id := instanceGroup.GetId()
		status := instanceGroup.GetStatus()
		if !updateComputeInstanceGroupWithSweeperDeps(conf, status, id, serviceAccountID, networkID, subnetID) {
			result = multierror.Append(result,
				fmt.Errorf("failed to sweep (update with dependencies) compute instance group %q", id))
			continue
		}

		if !sweepComputeInstanceGroup(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep compute instance group %q", id))
		}
	}

	if serviceAccountID != "" {
		if !iamsweep.SweepIAMServiceAccount(conf, serviceAccountID) {
			result = multierror.Append(result,
				fmt.Errorf("failed to sweep IAM service account %q", serviceAccountID))
		}
	}
	if subnetID != "" {
		if !vpcsweep.SweepVPCSubnet(conf, subnetID) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep VPC subnet %q", subnetID))
		}
	}
	if networkID != "" {
		if !vpcsweep.SweepVPCNetwork(conf, networkID) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep VPC network %q", networkID))
		}
	}

	return result.ErrorOrNil()
}

func updateComputeInstanceGroupWithSweeperDeps(conf *providerconfig.Config, status instancegroup.InstanceGroup_Status, instanceGroupID, serviceAccountID, networkID, subnetID string) bool {
	sweep.DebugLog("started updating instance group %q", instanceGroupID)
	updateMaskPath := []string{
		"deletion_protection",
		"allocation_policy",
		"service_account_id",
		"instance_template.network_interface_specs",
	}
	if status == instancegroup.InstanceGroup_DELETING {
		updateMaskPath = []string{"service_account_id"}
	}

	client := instancegroupsdk.NewInstanceGroupClient(conf.
		SDK)

	for i := 1; i <= conf.MaxRetries; i++ {
		req := &instancegroup.UpdateInstanceGroupRequest{
			InstanceGroupId:    instanceGroupID,
			DeletionProtection: false,
			ServiceAccountId:   serviceAccountID,
			AllocationPolicy: &instancegroup.AllocationPolicy{
				Zones: []*instancegroup.AllocationPolicy_Zone{
					{ZoneId: conf.Zone},
				},
			},
			InstanceTemplate: &instancegroup.InstanceTemplate{
				NetworkInterfaceSpecs: []*instancegroup.NetworkInterfaceSpec{
					{
						NetworkId:            networkID,
						SubnetIds:            []string{subnetID},
						PrimaryV4AddressSpec: &instancegroup.PrimaryAddressSpec{},
					},
				},
			},
			UpdateMask: &field_mask.FieldMask{
				Paths: updateMaskPath,
			},
		}

		_, err := client.Update(conf.Context(), req)
		if err != nil {
			sweep.DebugLog("[instance group %q] update try #%d: %v", instanceGroupID, i, err)
		} else {
			sweep.DebugLog("[instance group %q] update try #%d: request was successfully sent", instanceGroupID, i)
			return true
		}
	}

	sweep.DebugLog("instance group %q update failed", instanceGroupID)
	return false
}

func sweepComputeInstanceGroup(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepComputeInstanceGroupOnce, conf, "Compute instance group", id)
}

func sweepComputeInstanceGroupOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(svc_vpc.YandexVPCNetworkDefaultTimeout)
	defer cancel()

	op, err := instancegroupsdk.NewInstanceGroupClient(conf.
		SDK).
		Delete(ctx, &instancegroup.DeleteInstanceGroupRequest{
			InstanceGroupId: id,
		})
	return sweep.HandleSweepOperationV2(ctx, op, err)
}
