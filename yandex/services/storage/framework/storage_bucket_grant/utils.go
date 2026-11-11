package storage_bucket_grant

import (
	"slices"

	"github.com/aws/aws-sdk-go/service/s3"
	storage "github.com/yandex-cloud/terraform-provider-yandex/pkg/storage/s3"
)

// DetectACLFromGrants analyzes S3 grants and tries to detect if they match a predefined ACL pattern
func DetectACLFromGrants(s3Grants []*s3.Grant) string {
	if len(s3Grants) == 0 {
		return storage.BucketACLPrivate
	}

	grantPatterns := make(map[string][]string) // grantee -> permissions
	for _, grant := range s3Grants {
		if grant.Grantee == nil || grant.Permission == nil {
			continue
		}

		var granteeKey string
		if grant.Grantee.Type != nil {
			granteeType := *grant.Grantee.Type
			if granteeType == storage.TypeGroup && grant.Grantee.URI != nil {
				granteeKey = *grant.Grantee.URI
			}
		}
		if granteeKey != "" {
			grantPatterns[granteeKey] = append(grantPatterns[granteeKey], *grant.Permission)
		}
	}

	// Check for public-read pattern: AllUsers with READ permission
	if permissions, exists := grantPatterns["http://acs.amazonaws.com/groups/global/AllUsers"]; exists {
		if len(grantPatterns) == 1 {
			if slices.Contains(permissions, storage.PermissionRead) && len(permissions) == 1 {
				return storage.BucketCannedACLPublicRead
			}
		}
	}

	// Check for public-read-write pattern: AllUsers with READ and WRITE permissions
	if permissions, exists := grantPatterns["http://acs.amazonaws.com/groups/global/AllUsers"]; exists {
		if len(grantPatterns) == 1 {
			if slices.Contains(permissions, storage.PermissionRead) &&
				slices.Contains(permissions, storage.PermissionWrite) &&
				len(permissions) == 2 {
				return storage.BucketCannedACLPublicReadWrite
			}
		}
	}

	// Check for authenticated-read pattern: AuthenticatedUsers with READ permission
	if permissions, exists := grantPatterns["http://acs.amazonaws.com/groups/global/AuthenticatedUsers"]; exists {
		if len(grantPatterns) == 1 {
			if slices.Contains(permissions, storage.PermissionRead) && len(permissions) == 1 {
				return storage.BucketCannedACLAuthenticatedRead
			}
		}
	}

	return ""
}
