package yandex

import (
	"context"
	"fmt"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

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

	// GetResourceID Returns the unique resource identifier.
	GetResourceID() string

	// DescribeResource Textual description of this resource to be used in error message.
	// The description should include the unique resource identifier.
	DescribeResource() string
}

type newResourceIamUpdaterFunc func(d *schema.ResourceData, config *Config) (ResourceIamUpdater, error)
type iamPolicyModifyFunc func(p *Policy) error

type resourceIDParserFunc func(d *schema.ResourceData, config *Config) error

func iamPolicyReadModifySet(ctx context.Context, updater ResourceIamUpdater, modify iamPolicyModifyFunc) error {
	mutexKey := updater.GetMutexKey()
	mutexKV.Lock(mutexKey)
	defer mutexKV.Unlock(mutexKey)

	log.Printf("[DEBUG]: Retrieving access bindings for %s\n", updater.DescribeResource())

	p, err := updater.GetResourceIamPolicy(ctx)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG]: Retrieved access bindings for %s: %+v\n", updater.DescribeResource(), p)

	err = modify(p)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG]: Setting access bindings for %s to %+v\n", updater.DescribeResource(), p)

	err = updater.SetResourceIamPolicy(ctx, p)
	if err != nil {
		return fmt.Errorf("Error applying access bindings to %s: %w", updater.DescribeResource(), err)
	}

	log.Printf("[DEBUG]: Set policy for %s", updater.DescribeResource())

	return nil
}

func iamPolicyReadModifyUpdate(ctx context.Context, updater ResourceIamUpdater, policyDelta *PolicyDelta) error {
	mutexKey := updater.GetMutexKey()
	mutexKV.Lock(mutexKey)
	defer mutexKV.Unlock(mutexKey)

	log.Printf("[DEBUG]: Retrieving access bindings for %s\n", updater.DescribeResource())

	p, err := updater.GetResourceIamPolicy(ctx)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG]: Retrieved access bindings for %s: %+v\n", updater.DescribeResource(), p)

	log.Printf("[DEBUG]: Updating access bindings of %s with %+v\n", updater.DescribeResource(), policyDelta)

	err = updater.UpdateResourceIamPolicy(ctx, policyDelta)
	if err != nil {
		return fmt.Errorf("Error updating access bindings of %s: %w", updater.DescribeResource(), err)
	}

	log.Printf("[DEBUG]: Updated access bindings for %s", updater.DescribeResource())

	return nil
}
