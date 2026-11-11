package mdb_greenplum_resource_group

import (
	"context"
	"fmt"

	"github.com/yandex-cloud/go-sdk/services/mdb/greenplum/v1"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/greenplum/v1"
	operationpb "github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	ycsdk "github.com/yandex-cloud/go-sdk/v2"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/operationcompat"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/retry"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func requestCreateResourceGroup(ctx context.Context, sdk *ycsdk.SDK, req *greenplum.CreateResourceGroupRequest) (*operationpb.Operation, error) {
	conn, err := sdk.GetConnection(ctx, greenplumsdk.ResourceGroupCreate)
	if err != nil {
		return nil, err
	}
	return greenplum.NewResourceGroupServiceClient(conn).Create(ctx, req)
}

func requestUpdateResourceGroup(ctx context.Context, sdk *ycsdk.SDK, req *greenplum.UpdateResourceGroupRequest) (*operationpb.Operation, error) {
	conn, err := sdk.GetConnection(ctx, greenplumsdk.ResourceGroupUpdate)
	if err != nil {
		return nil, err
	}
	return greenplum.NewResourceGroupServiceClient(conn).Update(ctx, req)
}

func readResourceGroup(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, resourceGroupName string) *greenplum.ResourceGroup {
	rgs, err := greenplumsdk.NewResourceGroupClient(sdk).List(ctx, &greenplum.ListResourceGroupsRequest{
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
	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*operationpb.Operation, error) {
		return requestCreateResourceGroup(ctx, sdk, &greenplum.CreateResourceGroupRequest{
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

	if err = operationcompat.Wait(ctx, sdk, op.GetId()); err != nil {
		diag.AddError(
			"Failed to Create resource",
			"Error while waiting for operation to create Greenplum resource group: "+err.Error(),
		)
	}
}

func updateResourceGroup(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string, resourceGroup *greenplum.ResourceGroup, updatePaths []string) {
	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*operationpb.Operation, error) {
		return requestUpdateResourceGroup(ctx, sdk, &greenplum.UpdateResourceGroupRequest{
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

	if err = operationcompat.Wait(ctx, sdk, op.GetId()); err != nil {
		diag.AddError(
			"Failed to Update resource",
			"Error while waiting for operation to update Greenplum resource group: "+err.Error(),
		)
	}
}

func deleteResourceGroup(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid, resourceGroupName string) {
	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*greenplumsdk.ResourceGroupDeleteOperation, error) {
		return greenplumsdk.NewResourceGroupClient(sdk).Delete(ctx, &greenplum.DeleteResourceGroupRequest{
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

	if _, err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Delete resource",
			"Error while waiting for operation to delete Greenplum resource group: "+err.Error(),
		)
	}
}
