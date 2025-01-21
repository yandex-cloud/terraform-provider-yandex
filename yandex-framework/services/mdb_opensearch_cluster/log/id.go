package log

import "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_opensearch_cluster/model"

// IdFromStr returns ID as tflog additional parameters
func IdFromStr(cid string) map[string]interface{} {
	return map[string]interface{}{
		"id": cid,
	}
}

// IdFromModel returns ID as tflog additional parameters
func IdFromModel(m *model.OpenSearch) map[string]interface{} {
	return map[string]interface{}{
		"id": m.ID.ValueString(),
	}
}
