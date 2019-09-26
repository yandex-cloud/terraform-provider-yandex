package yandex

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type ResourceIamUpdater interface {
	// Fetch the existing IAM policy attached to a resource.
	GetResourceIamPolicy() (*Policy, error)

	// Replaces the existing IAM Policy attached to a resource.
	SetResourceIamPolicy(policy *Policy) error

	// A mutex guards against concurrent call to the SetResourceIamPolicy method.
	// The mutex key should be made of the resource type and resource id.
	// For example: `iam-folder-{id}`.
	GetMutexKey() string

	// Returns the unique resource identifier.
	GetResourceID() string

	// Textual description of this resource to be used in error message.
	// The description should include the unique resource identifier.
	DescribeResource() string
}

type newResourceIamUpdaterFunc func(d *schema.ResourceData, config *Config) (ResourceIamUpdater, error)
type iamPolicyModifyFunc func(p *Policy) error

type resourceIDParserFunc func(d *schema.ResourceData, config *Config) error

func iamPolicyReadModifyWrite(updater ResourceIamUpdater, modify iamPolicyModifyFunc) error {
	mutexKey := updater.GetMutexKey()
	mutexKV.Lock(mutexKey)
	defer mutexKV.Unlock(mutexKey)

	log.Printf("[DEBUG]: Retrieving policy for %s\n", updater.DescribeResource())

	p, err := updater.GetResourceIamPolicy()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG]: Retrieved policy for %s: %+v\n", updater.DescribeResource(), p)

	err = modify(p)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG]: Setting policy for %s to %+v\n", updater.DescribeResource(), p)

	err = updater.SetResourceIamPolicy(p)
	if err != nil {
		return fmt.Errorf("Error applying IAM policy for %s: %s", updater.DescribeResource(), err)
	}

	log.Printf("[DEBUG]: Set policy for %s", updater.DescribeResource())

	return nil
}
