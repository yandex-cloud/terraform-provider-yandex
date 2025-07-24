package mdb_greenplum_resource_group

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/greenplum/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/retry"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func readResourceGroup(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, resourceGroupName string) *greenplum.ResourceGroup {
	rgs, err := sdk.MDB().Greenplum().ResourceGroup().List(ctx, &greenplum.ListResourceGroupsRequest{
		ClusterId: cid,
	})
	if err != nil {
		diag.AddError(
			"Failed to Read resource",
			"Error while requesting API to get Greenplum resource group: "+err.Error(),
		)
		return nil
	}

	for _, rg := range rgs.GetResourceGroups() {
		if rg.GetName() == resourceGroupName {
			return rg
		}
	}
	diag.AddError(
		"Failed to Read resource",
		fmt.Sprintf("Greenplum resource group %q not found", resourceGroupName),
	)
	return nil
}

func createResourceGroup(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, resourceGroup *greenplum.ResourceGroup) {
	op, err := retry.ConflictingOperation(ctx, sdk, func() (*operation.Operation, error) {
		return sdk.MDB().Greenplum().ResourceGroup().Create(ctx, &greenplum.CreateResourceGroupRequest{
			ClusterId:     cid,
			ResourceGroup: resourceGroup,
		})
	})

	if err != nil {
		diag.AddError(
			"Failed to Create resource",
			"Error while requesting API to create Greenplum resource group: "+err.Error(),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Create resource",
			"Error while waiting for operation to create Greenplum resource group: "+err.Error(),
		)
	}
}

func updateResourceGroup(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, resourceGroup *greenplum.ResourceGroup, updatePaths []string) {
	op, err := retry.ConflictingOperation(ctx, sdk, func() (*operation.Operation, error) {
		return sdk.MDB().Greenplum().ResourceGroup().Update(ctx, &greenplum.UpdateResourceGroupRequest{
			ClusterId:     cid,
			ResourceGroup: resourceGroup,
			UpdateMask:    &fieldmaskpb.FieldMask{Paths: updatePaths},
		})
	})

	if err != nil {
		diag.AddError(
			"Failed to Update resource",
			"Error while requesting API to update Greenplum resource group: "+err.Error(),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Update resource",
			"Error while waiting for operation to update Greenplum resource group: "+err.Error(),
		)
	}
}

func deleteResourceGroup(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid, resourceGroupName string) {
	op, err := retry.ConflictingOperation(ctx, sdk, func() (*operation.Operation, error) {
		return sdk.MDB().Greenplum().ResourceGroup().Delete(ctx, &greenplum.DeleteResourceGroupRequest{
			ClusterId:         cid,
			ResourceGroupName: resourceGroupName,
		})
	})

	if err != nil {
		diag.AddError(
			"Failed to Delete resource",
			"Error while requesting API to delete Greenplum resource group: "+err.Error(),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Delete resource",
			"Error while waiting for operation to delete Greenplum resource group: "+err.Error(),
		)
	}
}
