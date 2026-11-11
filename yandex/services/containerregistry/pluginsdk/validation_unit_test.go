package pluginsdk_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
	svc_container_registry "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/containerregistry/pluginsdk"
)

func TestEveryOf(t *testing.T) {
	t.Parallel()

	t.Run("checkEveryOf", func(t *testing.T) {
		d := (&schema.Resource{
			//lintignore:XS001
			Schema: map[string]*schema.Schema{
				"field_one": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"field_two": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
			},
		}).TestResourceData()

		d.Set("field_one", "value_one")
		d.Set("field_two", "value_two")

		require.NoError(t, svc_container_registry.CheckEveryOf(d, "field_one", "field_two"), "success case")
		require.Error(t, svc_container_registry.CheckEveryOf(d, "field_one", "field_two", "field_three"), "non existent keys")
		require.Error(t, svc_container_registry.CheckEveryOf(d, ""), "empty keys not allowed")
		require.Error(t, svc_container_registry.CheckEveryOf(d, "field_one", ""), "empty keys not allowed")
	})
}
