package pluginsdk

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func stringSliceToLower(s []string) []string {
	var ret []string
	for _, v := range s {
		ret = append(ret, strings.ToLower(v))
	}
	return ret
}

type schemaGetHelper struct {
	pathPrefix string
	d          *schema.ResourceData
}

func schemaHelper(d *schema.ResourceData, path string) *schemaGetHelper {
	return &schemaGetHelper{
		pathPrefix: path,
		d:          d,
	}
}

func (h *schemaGetHelper) AppendPath(path string) *schemaGetHelper {
	return &schemaGetHelper{
		pathPrefix: h.pathPrefix + path,
		d:          h.d,
	}
}

func (h *schemaGetHelper) Get(key string) interface{} {
	return h.d.Get(h.pathPrefix + key)
}

func (h *schemaGetHelper) GetOk(key string) (interface{}, bool) {
	return h.d.GetOk(h.pathPrefix + key)
}

func (h *schemaGetHelper) GetString(key string) string {
	return h.d.Get(h.pathPrefix + key).(string)
}

func (h *schemaGetHelper) GetInt(key string) int {
	return h.d.Get(h.pathPrefix + key).(int)
}
