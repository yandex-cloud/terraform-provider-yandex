package s3

import (
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	BucketACLPrivate = "private"
)

const (
	BucketOwnerFullControl           = "bucket-owner-full-control"
	BucketCannedACLPublicRead        = s3.BucketCannedACLPublicRead
	BucketCannedACLPublicReadWrite   = s3.BucketCannedACLPublicReadWrite
	BucketCannedACLAuthenticatedRead = s3.BucketCannedACLAuthenticatedRead
	BucketCannedACLPrivate           = s3.BucketCannedACLPrivate
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
