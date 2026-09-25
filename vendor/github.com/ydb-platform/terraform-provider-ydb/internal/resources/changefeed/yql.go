package changefeed

import "github.com/ydb-platform/terraform-provider-ydb/internal/helpers"

func PrepareCreateRequest(cdc *changeDataCaptureSettings) string {
	buf := make([]byte, 0, 256)
	buf = append(buf, "ALTER TABLE `"...)
	buf = helpers.AppendWithEscape(buf, helpers.TrimPath(cdc.getTablePath()))
	buf = append(buf, '`', ' ')
	buf = append(buf, "ADD CHANGEFEED `"...)
	buf = helpers.AppendWithEscape(buf, cdc.Name)
	buf = append(buf, '`', ' ')
	buf = append(buf, "WITH ("...)
	buf = append(buf, '\n')
	buf = append(buf, "MODE = \""...)
	buf = append(buf, cdc.Mode...)
	buf = append(buf, '"')
	if cdc.Format != nil && *cdc.Format != "" {
		buf = append(buf, ',', '\n')
		buf = append(buf, "FORMAT = \""...)
		buf = append(buf, *cdc.Format...)
		buf = append(buf, '"')
	}
	if cdc.VirtualTimestamps != nil {
		buf = append(buf, ',', '\n')
		buf = append(buf, "VIRTUAL_TIMESTAMPS = "...)
		if *cdc.VirtualTimestamps {
			buf = append(buf, "true"...)
		} else {
			buf = append(buf, "false"...)
		}
	}
	if cdc.RetentionPeriod != nil && *cdc.RetentionPeriod != "" {
		buf = append(buf, ',', '\n')
		buf = append(buf, "RETENTION_PERIOD = Interval(\""...)
		buf = helpers.AppendWithEscape(buf, *cdc.RetentionPeriod)
		buf = append(buf, '"', ')')
	}
	buf = append(buf, '\n', ')')
	return string(buf)
}

func PrepareDropRequest(tablePath, cdcName string) string {
	buf := make([]byte, 0, 64)
	buf = append(buf, "ALTER TABLE `"...)
	buf = helpers.AppendWithEscape(buf, helpers.TrimPath(tablePath))
	buf = append(buf, '`', ' ')
	buf = append(buf, "DROP CHANGEFEED `"...)
	buf = helpers.AppendWithEscape(buf, cdcName)
	buf = append(buf, '`')

	return string(buf)
}
