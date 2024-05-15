package iam

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	"github.com/yandex-cloud/terraform-provider-yandex/common/mutexkv"
)

var mutexKV = mutexkv.NewMutexKV()

type Policy struct {
	Bindings []*access.AccessBinding
}

type PolicyDelta struct {
	Deltas []*access.AccessBindingDelta
}

type ResourceIamUpdater interface {
	// GetResourceIamPolicy Fetch the existing IAM policy attached to a resource.
	GetResourceIamPolicy(ctx context.Context) (*Policy, error)

	// SetResourceIamPolicy Replaces the existing IAM Policy attached to a resource.
	// Useful for `iam_binding` and `iam_policy` resources
	SetResourceIamPolicy(ctx context.Context, policy *Policy) error

	// UpdateResourceIamPolicy Updates the existing IAM Policy attached to a resource.
	// Useful for `iam_member` resources
	UpdateResourceIamPolicy(ctx context.Context, policy *PolicyDelta) error

	// GetMutexKey A mutex guards against concurrent call to the SetResourceIamPolicy method.
	// The mutex key should be made of the resource type and resource id.
	// For example: `iam-folder-{id}`.
	GetMutexKey() string

	// Initialize Initializing resource from given state
	Initialize(ctx context.Context, state Extractable, diag *diag.Diagnostics)

	// DescribeResource Textual description of this resource to be used in error message.
	// The description should include the unique resource identifier.
	DescribeResource() string

	// Configure Configurate resource with provider data etc.
	Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse)

	// GetSchemaAttributes Gets resource iam schema. Usually single field with resource id name alias.
	GetSchemaAttributes() map[string]schema.Attribute

	// GetNameSuffix Gets resource terraform name suffix without leading underscore.
	GetNameSuffix() string

	// GetIdAlias Gets resource id alias that used in resource schema for resource configuration.
	GetIdAlias() string

	// GetId Gets resource id.
	GetId() string
}

type Extractable interface {
	GetAttribute(ctx context.Context, path path.Path, target interface{}) diag.Diagnostics
}

type Settable interface {
	SetAttribute(ctx context.Context, path path.Path, val interface{}) diag.Diagnostics
}

type iamPolicyModifyFunc func(p *Policy) error

func iamPolicyReadModifySet(ctx context.Context, updater ResourceIamUpdater, modify iamPolicyModifyFunc) error {
	mutexKey := updater.GetMutexKey()
	mutexKV.Lock(mutexKey)
	defer mutexKV.Unlock(mutexKey)

	tflog.Debug(ctx, fmt.Sprintf("Retrieving access bindings for %s", updater.DescribeResource()))

	p, err := updater.GetResourceIamPolicy(ctx)
	if err != nil {
		return err
	}

	tflog.Debug(ctx, fmt.Sprintf("Retrieved access bindings for %s: %+v\n", updater.DescribeResource(), p))

	err = modify(p)
	if err != nil {
		return err
	}

	tflog.Debug(ctx, fmt.Sprintf("Setting access bindings for %s to %+v", updater.DescribeResource(), p))

	err = updater.SetResourceIamPolicy(ctx, p)
	if err != nil {
		return fmt.Errorf("Error applying access bindings to %s: %w", updater.DescribeResource(), err)
	}

	tflog.Debug(ctx, fmt.Sprintf("Set policy for %s", updater.DescribeResource()))

	return nil
}
