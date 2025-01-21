package billing_cloud_binding

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/billing/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

type InstanceID struct {
	BillingAccountId    string
	ServiceInstanceType string
	ServiceInstanceId   string
}

func (id InstanceID) compute() string {
	return id.BillingAccountId + "/" + id.ServiceInstanceType + "/" + id.ServiceInstanceId
}

func ParseBindServiceInstanceId(s string) (*InstanceID, error) {
	splitId := strings.Split(s, "/")

	if len(splitId) != 3 {
		return nil, fmt.Errorf("unexcepted Id format occured while parsing InstanceID")
	}

	id := InstanceID{
		BillingAccountId:    splitId[0],
		ServiceInstanceType: splitId[1],
		ServiceInstanceId:   splitId[2],
	}

	return &id, nil
}

type yandexBillingBindingState struct {
	billingAccountID  types.String
	id                types.String
	serviceInstanceID types.String
}

func isObjectExist(ctx context.Context, sdk *ycsdk.SDK, resourceType,
	billingAccountId, serviceInstanceId string) bool {
	bindingsRequest := billing.ListBillableObjectBindingsRequest{
		BillingAccountId: billingAccountId,
	}

	for it := sdk.Billing().BillingAccount().BillingAccountBillableObjectBindingsIterator(
		ctx, &bindingsRequest); it.Next(); {
		billableObject := it.Value().BillableObject
		if billableObject.Type == resourceType && billableObject.Id == serviceInstanceId {
			return true
		}
	}

	return false
}

type extractable interface {
	GetAttribute(ctx context.Context, path path.Path, target interface{}) diag.Diagnostics
}

func getAllRequestAttributes(ctx context.Context, state *yandexBillingBindingState, serviceInstanceIdFieldName string,
	request extractable, diagnostics *diag.Diagnostics) {
	diagnostics.Append(request.GetAttribute(ctx, path.Root(idFieldName), &state.id)...)
	diagnostics.Append(request.GetAttribute(ctx, path.Root(accountIDFieldName), &state.billingAccountID)...)
	diagnostics.Append(request.GetAttribute(ctx, path.Root(serviceInstanceIdFieldName), &state.serviceInstanceID)...)
}

func setAllResponseAttributes(ctx context.Context, state yandexBillingBindingState, serviceInstanceIdFieldName string,
	responseState *tfsdk.State, diagnostics *diag.Diagnostics) {
	defer func() {
		if diagnostics.HasError() {
			diagnostics.AddError("Failed to update state", "")
		}
	}()

	diagnostics.Append(responseState.SetAttribute(ctx, path.Root(idFieldName), state.id.ValueString())...)
	if diagnostics.HasError() {
		return
	}
	diagnostics.Append(responseState.SetAttribute(ctx, path.Root(accountIDFieldName), state.billingAccountID.ValueString())...)
	if diagnostics.HasError() {
		return
	}
	diagnostics.Append(responseState.SetAttribute(ctx, path.Root(serviceInstanceIdFieldName), state.serviceInstanceID.ValueString())...)
}
