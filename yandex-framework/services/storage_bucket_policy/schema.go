package storage_bucket_policy

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	awspolicy "github.com/jen20/awspolicyequivalence"
)

var (
	_ = basetypes.StringValuable(&PolicyValue{})
	_ = basetypes.StringValuableWithSemanticEquals(&PolicyValue{})
	_ = fmt.Stringer(&PolicyValue{})
)

// PolicyType is a custom type for policy that implements semantic equality
type PolicyType struct {
	basetypes.StringType
}

type PolicyValue basetypes.StringValue

func (t PolicyType) Equal(o attr.Type) bool {
	_, ok := o.(PolicyType)

	return ok
}

func (t PolicyType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	v, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	str, ok := v.(basetypes.StringValue)
	if !ok {
		return nil, err
	}

	return PolicyValue(str), nil
}

func (t PolicyType) ValueFromString(ctx context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return PolicyValue(in), nil
}

func (v PolicyValue) Type(ctx context.Context) attr.Type {
	return PolicyType{}
}

func (v PolicyValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	return basetypes.StringValue(v).ToTerraformValue(ctx)
}

func (v PolicyValue) Equal(o attr.Value) bool {
	other, ok := o.(PolicyValue)
	if !ok {
		return false
	}
	return basetypes.StringValue(v).Equal(basetypes.StringValue(other))
}

func (v PolicyValue) IsNull() bool {
	return basetypes.StringValue(v).IsNull()
}

func (v PolicyValue) IsUnknown() bool {
	return basetypes.StringValue(v).IsUnknown()
}

func (v PolicyValue) String() string {
	return basetypes.StringValue(v).String()
}

func (v PolicyValue) ValueString() string {
	return basetypes.StringValue(v).ValueString()
}

func (v PolicyValue) ToStringValue(ctx context.Context) (basetypes.StringValue, diag.Diagnostics) {
	return basetypes.StringValue(v), nil
}

func (v PolicyValue) StringSemanticEquals(ctx context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	tflog.Debug(ctx, "PolicyType.SemanticEquals called")

	newVal, ok := newValuable.(PolicyValue)
	if !ok {
		tflog.Debug(ctx, "newAttrV is not PolicyValue", map[string]interface{}{
			"type": fmt.Sprintf("%T", newValuable),
		})
		return false, diag.Diagnostics{}
	}

	if newVal.IsNull() || newVal.IsUnknown() || v.IsNull() || v.IsUnknown() {
		result := newVal.Equal(v)
		tflog.Debug(ctx, "One of the values is null or unknown", map[string]interface{}{
			"result": result,
		})
		return result, nil
	}

	newPolicy := newVal.ValueString()
	oldPolicy := v.ValueString()

	tflog.Debug(ctx, "Comparing policies", map[string]interface{}{
		"newPolicy": newPolicy,
		"oldPolicy": oldPolicy,
	})

	equivalent, err := awspolicy.PoliciesAreEquivalent(oldPolicy, newPolicy)
	if err != nil {
		tflog.Debug(ctx, "Error comparing policies", map[string]interface{}{
			"error": err.Error(),
		})
		// If we can't determine equivalence, return false
		return false, nil
	}

	tflog.Debug(ctx, "Policy comparison result", map[string]interface{}{
		"equivalent": equivalent,
	})

	return equivalent, nil
}

func ResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Allows management of policy of an existing [Yandex Cloud Storage Bucket](https://yandex.cloud/docs/storage/concepts/bucket).\n\n~> By default, for authentication, you need to use [IAM token](https://yandex.cloud/docs/iam/concepts/authorization/iam-token) with the necessary permissions.\n\n~> Alternatively, you can provide [static access keys](https://yandex.cloud/docs/iam/concepts/authorization/access-key) (Access and Secret). To generate these keys, you will need a Service Account with the appropriate permissions.\n\n~> \"Version\" element is required and must be set to `2012-10-17`.",
		Attributes: map[string]schema.Attribute{
			"bucket": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the bucket.",
			},
			"policy": schema.StringAttribute{
				CustomType:          PolicyType{},
				Required:            true,
				MarkdownDescription: "The text of the policy.",
			},
			"access_key": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "The access key to use when applying changes. This value can also be provided as `storage_access_key` specified in provider config (explicitly or within `shared_credentials_file`) is used.",
			},
			"secret_key": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "The secret key to use when applying changes. This value can also be provided as `storage_secret_key` specified in provider config (explicitly or within `shared_credentials_file`) is used.",
			},
		},
	}
}
