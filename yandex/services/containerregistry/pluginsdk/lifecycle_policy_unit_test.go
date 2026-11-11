package pluginsdk_test

import (
	"reflect"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
	svc_container_registry "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/containerregistry/pluginsdk"
)

func TestFlattenContainerRepositoryLifecyclePolicyRule(t *testing.T) {
	t.Parallel()

	t.Run("FlattenContainerRepositoryLifecyclePolicyRule", func(t *testing.T) {
		test := struct {
			rule     *containerregistry.LifecycleRule
			expected interface{}
		}{
			rule: &containerregistry.LifecycleRule{
				Description:  "test description",
				ExpirePeriod: durationpb.New(24 * time.Hour),
				TagRegexp:    ".*",
				Untagged:     true,
				RetainedTop:  int64(5),
			},
			expected: map[string]interface{}{
				"description":   "test description",
				"expire_period": "24h0m0s",
				"tag_regexp":    ".*",
				"untagged":      true,
				"retained_top":  int64(5),
			},
		}

		got := svc_container_registry.FlattenContainerRepositoryLifecyclePolicyRule(test.rule)
		if !reflect.DeepEqual(test.expected, got) {
			t.Errorf("%#v is not equal to %#v", test.expected, got)
		}
	})
}
