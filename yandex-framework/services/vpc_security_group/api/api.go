package api

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/retry"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	"google.golang.org/grpc/codes"
)

func ReadSecurityGroup(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, sgID string) *vpc.SecurityGroup {
	tflog.Debug(ctx, "Reading VPC SecurityGroup", map[string]interface{}{"id": sgID})
	sg, err := sdk.VPC().SecurityGroup().Get(ctx, &vpc.GetSecurityGroupRequest{
		SecurityGroupId: sgID,
	})

	if err != nil {
		if validate.IsStatusWithCode(err, codes.NotFound) {
			return nil
		}

		diag.AddError(
			"Failed to Read resource",
			"Error while requesting API to get SecurityGroup:"+err.Error(),
		)
		return nil
	}
	return sg
}

func DeleteSecurityGroup(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, sgID string) {
	tflog.Debug(ctx, "Deleting VPC SecurityGroup", map[string]interface{}{"id": sgID})
	op, err := retry.ConflictingOperation(ctx, sdk, func() (*operation.Operation, error) {
		return sdk.VPC().SecurityGroup().Delete(ctx, &vpc.DeleteSecurityGroupRequest{
			SecurityGroupId: sgID,
		})
	})

	if err != nil {
		diag.AddError(
			"Failed to Delete resource",
			"Error while requesting API to delete Security Group: "+err.Error(),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Delete resource",
			"Error while waiting for operation to delete Security Group: "+err.Error(),
		)
	}
}

func FindSecurityGroupRule(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, sgID string, ruleID string) *vpc.SecurityGroupRule {
	sg := ReadSecurityGroup(ctx, sdk, diag, sgID)
	if diag.HasError() {
		return nil
	}

	if sg == nil {
		diag.AddError(
			"Failed to get SecurityGroup data",
			fmt.Sprintf("SecurityGroup with id %s not found", sgID))
		return nil
	}

	for _, rule := range sg.Rules {
		if rule.Id == ruleID {
			return rule
		}
	}

	return nil
}

func UpdateSecurityGroupRuleMetadata(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, req *vpc.UpdateSecurityGroupRuleRequest) {
	tflog.Debug(ctx, "Updating VPC SecurityGroupRule Metadata", map[string]interface{}{"security_group_binding": req.SecurityGroupId})
	op, err := retry.ConflictingOperation(ctx, sdk, func() (*operation.Operation, error) {
		return sdk.VPC().SecurityGroup().UpdateRule(ctx, req)
	})

	if err != nil {
		diag.AddError(
			"Failed to Update SecurityGroupRule Metadata",
			"Error while requesting API to update SecurityGroup rule metadata: "+err.Error(),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Update SecurityGroupRule Metadata",
			"Error while waiting for operation to update SecurityGroup rule metadata: "+err.Error(),
		)
		return
	}
}

func UpdateSecurityGroupRules(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, sgID string, addRule *vpc.SecurityGroupRuleSpec, deleteRuleID string) *vpc.UpdateSecurityGroupMetadata {
	if addRule != nil && deleteRuleID != "" {
		tflog.Debug(ctx, "Replacing VPC SecurityGroupRule", map[string]interface{}{"security_group_binding": sgID, "id": deleteRuleID})
	} else if addRule != nil {
		tflog.Debug(ctx, "Adding VPC SecurityGroupRule", map[string]interface{}{"security_group_binding": sgID})
	} else if deleteRuleID != "" {
		tflog.Debug(ctx, "Deleting VPC SecurityGroupRule", map[string]interface{}{"security_group_binding": sgID, "id": deleteRuleID})
	}
	req := vpc.UpdateSecurityGroupRulesRequest{
		SecurityGroupId: sgID,
	}
	if addRule != nil {
		req.AdditionRuleSpecs = []*vpc.SecurityGroupRuleSpec{addRule}
	}
	if deleteRuleID != "" {
		req.DeletionRuleIds = []string{deleteRuleID}
	}
	op, err := retry.ConflictingOperation(ctx, sdk, func() (*operation.Operation, error) {
		return sdk.VPC().SecurityGroup().UpdateRules(ctx, &req)
	})

	if err != nil {
		diag.AddError(
			"Failed to Update SecurityGroup Rules",
			"Error while requesting API to update SecurityGroup rules: "+err.Error(),
		)
		return nil
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Update SecurityGroup Rules",
			"Error while waiting for operation to update SecurityGroup rules: "+err.Error(),
		)
		return nil
	}

	metadata, err := op.Metadata()
	if err != nil {
		diag.AddError(
			"Failed to get UpdateSecurityGroupRules operation metadata",
			"Error while getting UpdateSecurityGroupRules operation metadata: "+err.Error(),
		)
		return nil
	}

	meta, ok := metadata.(*vpc.UpdateSecurityGroupMetadata)
	if !ok {
		diag.AddError(
			"Failed to get UpdateSecurityGroupRules operation metadata",
			"Error while getting UpdateSecurityGroupRules operation metadata: "+err.Error(),
		)
		return nil
	}

	return meta
}
