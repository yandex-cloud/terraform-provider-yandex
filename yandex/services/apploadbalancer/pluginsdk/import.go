package pluginsdk

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type crudFunc = func(d *schema.ResourceData, meta interface{}) error

func WithALBVirtualHostID(r *schema.Resource) *schema.Resource {
	r.Read = WrapParseVirtualHostID(r.Read)
	r.Update = WrapParseVirtualHostID(r.Update)
	r.Delete = WrapParseVirtualHostID(r.Delete)
	return r
}

func WrapParseVirtualHostID(f crudFunc) crudFunc {
	return func(d *schema.ResourceData, meta interface{}) error {
		attrs := strings.Split(d.Id(), "/")
		if len(attrs) < 2 {
			return fmt.Errorf("error reading virtual_host, wrong id: %q", d.Id())
		}
		if err := d.Set("http_router_id", attrs[0]); err != nil {
			return err
		}
		if err := d.Set("name", attrs[1]); err != nil {
			return err
		}
		return f(d, meta)
	}
}
