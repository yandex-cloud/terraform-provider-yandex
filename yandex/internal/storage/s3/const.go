package s3

import "github.com/aws/aws-sdk-go/service/s3"

const (
	StorageClassStandard = s3.StorageClassStandardIa
	StorageClassCold     = "COLD"
	StorageClassIce      = "ICE"
)

const (
	TypeCanonicalUser = s3.TypeCanonicalUser
	TypeGroup         = s3.TypeGroup
)

const (
	PermissionFullControl = s3.PermissionFullControl
	PermissionRead        = s3.PermissionRead
	PermissionWrite       = s3.PermissionWrite
)

const (
	ObjectLockEnabled = s3.ObjectLockEnabledEnabled
)

const (
	ServerSideEncryptionAwsKms = s3.ServerSideEncryptionAwsKms
)

var (
	ObjectLockEnabledValues         = s3.ObjectLockEnabled_Values()
	ObjectLockRetentionModeValues   = s3.ObjectLockRetentionMode_Values()
	ObjectLockLegalHoldStatusValues = s3.ObjectLockLegalHoldStatus_Values()
)
