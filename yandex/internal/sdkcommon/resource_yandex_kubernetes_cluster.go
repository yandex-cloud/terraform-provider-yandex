package sdkcommon

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ShouldSuppressDiffForTimeDuration(k, old, new string, d *schema.ResourceData) bool {
	d1, err := ParseDuration(old)
	if err != nil {
		return false
	}

	d2, err := ParseDuration(new)
	if err != nil {
		return false
	}

	if d1 == nil && d2 == nil {
		return true
	}

	if d1 != nil && d2 != nil {
		return d1.Seconds == d2.Seconds && d1.Nanos == d2.Nanos
	}

	return false
}
