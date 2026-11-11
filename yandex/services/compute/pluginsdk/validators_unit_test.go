package pluginsdk_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	svc_compute "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/compute/pluginsdk"
)

func TestFloatAtLeast(t *testing.T) {
	testCases := []struct {
		val         interface{}
		f           schema.SchemaValidateFunc
		expectedErr *regexp.Regexp
	}{
		{
			val: float64(0.1),
			f:   svc_compute.FloatAtLeast(0),
		},
		{
			val: float64(0),
			f:   svc_compute.FloatAtLeast(0),
		},
		{
			val:         int(1),
			f:           svc_compute.FloatAtLeast(0),
			expectedErr: regexp.MustCompile("expected type of test_property to be float64"),
		},
		{
			val:         float64(-1),
			f:           svc_compute.FloatAtLeast(0),
			expectedErr: regexp.MustCompile(`expected test_property to be at least \(0.000000\), got -1.000000`),
		},
	}

	for i, tc := range testCases {
		_, errs := tc.f(tc.val, "test_property")

		if len(errs) == 0 && tc.expectedErr == nil {
			continue
		}

		if len(errs) != 0 && tc.expectedErr == nil {
			t.Fatalf("expected test case %d to produce no errors, got %v", i, errs)
		}

		if !matchErr(errs, tc.expectedErr) {
			t.Fatalf("expected test case %d to produce error matching \"%s\", got %v", i, tc.expectedErr, errs)
		}
	}
}

func matchErr(errs []error, r *regexp.Regexp) bool {
	// err must match one provided
	for _, err := range errs {
		if r.MatchString(err.Error()) {
			return true
		}
	}

	return false
}
