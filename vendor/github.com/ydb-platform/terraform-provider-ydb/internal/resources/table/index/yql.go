package index

import "github.com/ydb-platform/terraform-provider-ydb/internal/helpers"

func prepareCreateIndexRequest(r *resource) string {
	req := []byte("ALTER TABLE `")
	req = helpers.AppendWithEscape(req, helpers.TrimPath(r.getTablePath()))
	req = append(req, '`', ' ')
	req = append(req, "ADD INDEX `"...)
	req = helpers.AppendWithEscape(req, r.Name)
	req = append(req, '`', ' ')
	// TODO(shmel1k@): add ToYQL for index
	if r.Type == "global_async" { // TODO(shmel1k@): move to consts
		req = append(req, "GLOBAL ASYNC ON ("...)
	} else {
		req = append(req, "GLOBAL SYNC ON ("...)
	}
	for i := 0; i < len(r.Columns); i++ {
		req = append(req, '`')
		req = helpers.AppendWithEscape(req, r.Columns[i])
		req = append(req, '`')
		if i != len(r.Columns)-1 {
			req = append(req, ',', ' ')
		}
	}
	req = append(req, ')')
	if len(r.Cover) > 0 {
		req = append(req, " COVER ("...)
		for i := 0; i < len(r.Cover); i++ {
			req = append(req, '`')
			req = helpers.AppendWithEscape(req, r.Cover[i])
			req = append(req, '`')
			if i != len(r.Cover)-1 {
				req = append(req, ',', ' ')
			}
		}
		req = append(req, ')')
	}

	return string(req)
}

func prepareDropRequest(tablePath, indexName string) string {
	req := []byte("ALTER TABLE `")
	req = helpers.AppendWithEscape(req, helpers.TrimPath(tablePath))
	req = append(req, '`', ' ')
	req = append(req, "DROP INDEX `"...)
	req = helpers.AppendWithEscape(req, indexName)
	req = append(req, '`')
	return string(req)
}
