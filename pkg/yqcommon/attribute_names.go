package yqcommon

const (
	AttributeID               = "id"
	AttributeBucket           = "bucket"
	AttributeCloudID          = "cloud_id"
	AttributeFolderID         = "folder_id"
	AttributeDatabaseID       = "database_id"
	AttributeDescription      = "description"
	AttributeName             = "name"
	AttributeServiceAccountID = "service_account_id"
	AttributeSharedReading    = "shared_reading"
)

// bindings
const (
	AttributeCompression   = "compression"
	AttributeConnectionID  = "connection_id"
	AttributeFormat        = "format"
	AttributeFormatSetting = "format_setting"
	AttributePartitionedBy = "partitioned_by"
	AttributePathPattern   = "path_pattern"
	AttributeProjection    = "projection"
	AttributeStream        = "stream"

	// the same names as for ydb table
	AttributeColumn        = "column"
	AttributeColumnName    = "name"
	AttributeColumnNotNull = "not_null"
	AttributeColumnType    = "type"
)
