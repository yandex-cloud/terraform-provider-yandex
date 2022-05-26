//revive:disable:var-naming
package yandex

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1/awscompatibility"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/kms/v1"
)

func init() {
	resource.AddTestSweepers("yandex_storage_bucket", &resource.Sweeper{
		Name: "yandex_storage_bucket",
		F:    testSweepStorageBucket,
		Dependencies: []string{
			"yandex_storage_object",
		},
	})
}

func testSweepStorageBucket(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	result := &multierror.Error{}
	serviceAccountID, err := createIAMServiceAccountForSweeper(conf)
	if serviceAccountID != "" {
		defer func() {
			if !sweepIAMServiceAccount(conf, serviceAccountID) {
				result = multierror.Append(result,
					fmt.Errorf("failed to sweep IAM service account %q", serviceAccountID))
			}
		}()
	}
	if err != nil {
		result = multierror.Append(result, fmt.Errorf("error creating service account: %s", err))
		return result.ErrorOrNil()
	}

	resp, err := conf.sdk.IAM().AWSCompatibility().AccessKey().Create(conf.Context(), &awscompatibility.CreateAccessKeyRequest{
		ServiceAccountId: serviceAccountID,
		Description:      "Storage Bucket sweeper static key",
	})
	if err != nil {
		result = multierror.Append(result, fmt.Errorf("error creating service account static key: %s", err))
		return result.ErrorOrNil()
	}

	defer func() {
		_, err := conf.sdk.IAM().AWSCompatibility().AccessKey().Delete(conf.Context(), &awscompatibility.DeleteAccessKeyRequest{
			AccessKeyId: resp.AccessKey.Id,
		})
		if err != nil {
			result = multierror.Append(result, fmt.Errorf("error deleting service account static key: %s", err))
		}
	}()

	s3client, err := getS3ClientByKeys(resp.AccessKey.KeyId, resp.Secret, conf)
	if err != nil {
		result = multierror.Append(result, fmt.Errorf("error creating storage client: %s", err))
		return result.ErrorOrNil()
	}

	buckets, err := s3client.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		result = multierror.Append(result, fmt.Errorf("failed to list storage buckets: %s", err))
		return result.ErrorOrNil()
	}

	for _, b := range buckets.Buckets {
		_, err := s3client.DeleteBucket(&s3.DeleteBucketInput{
			Bucket: b.Name,
		})

		if err != nil {
			result = multierror.Append(result, fmt.Errorf("failed to delete bucket: %s, error: %s", *b.Name, err))
		}
	}

	return result.ErrorOrNil()
}

func TestAccStorageBucket_basic(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "yandex_storage_bucket.test"

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckStorageBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageBucketExists(resourceName),
					resource.TestCheckNoResourceAttr(resourceName, "website_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "bucket", testAccBucketName(rInt)),
					resource.TestCheckResourceAttr(resourceName, "bucket_domain_name", testAccBucketDomainName(rInt)),
				),
			},
		},
	})
}

func TestAccStorageBucket_namePrefix(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "yandex_storage_bucket.test"

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckStorageBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketConfigWithNamePrefix(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageBucketExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "bucket", regexp.MustCompile("^tf-test-")),
				),
			},
		},
	})
}

func TestAccStorageBucket_generatedName(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "yandex_storage_bucket.test"

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckStorageBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketConfigWithGeneratedName(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageBucketExists(resourceName),
				),
			},
		},
	})
}

func TestAccStorageBucket_Policy(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "yandex_storage_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:        func() { testAccPreCheck(t) },
		IDRefreshName:   resourceName,
		IDRefreshIgnore: []string{"access_key", "secret_key"},
		Providers:       testAccProviders,
		CheckDestroy:    testAccCheckStorageBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketConfigWithPolicy(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageBucketExists(resourceName),
					testAccCheckStorageBucketPolicy(resourceName, testAccStorageBucketPolicy(rInt)),
				),
			},
		},
	})
}

func TestAccStorageBucket_PolicyNone(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "yandex_storage_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:        func() { testAccPreCheck(t) },
		IDRefreshName:   resourceName,
		IDRefreshIgnore: []string{"access_key", "secret_key"},
		Providers:       testAccProviders,
		CheckDestroy:    testAccCheckStorageBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketBasic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageBucketExists(resourceName),
					testAccCheckStorageBucketPolicy(resourceName, ""),
				),
			},
		},
	})
}

func TestAccStorageBucket_PolicyEmpty(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "yandex_storage_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:        func() { testAccPreCheck(t) },
		IDRefreshName:   resourceName,
		IDRefreshIgnore: []string{"access_key", "secret_key"},
		Providers:       testAccProviders,
		CheckDestroy:    testAccCheckStorageBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketConfigWithEmptyPolicy(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageBucketExists(resourceName),
					testAccCheckStorageBucketPolicy(resourceName, ""),
				),
			},
		},
	})
}

func TestAccStorageBucket_updateAcl(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "yandex_storage_bucket.test"

	resource.Test(t, resource.TestCase{
		PreCheck:        func() { testAccPreCheck(t) },
		IDRefreshName:   resourceName,
		IDRefreshIgnore: []string{"access_key", "secret_key"},
		Providers:       testAccProviders,
		CheckDestroy:    testAccCheckStorageBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketAclPreConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "acl", "public-read"),
				),
			},
			{
				Config: testAccStorageBucketAclPostConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "acl", "private"),
				),
			},
		},
	})
}

func TestAccStorageBucket_Website_Simple(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "yandex_storage_bucket.test"

	resource.Test(t, resource.TestCase{
		PreCheck:        func() { testAccPreCheck(t) },
		IDRefreshName:   resourceName,
		IDRefreshIgnore: []string{"access_key", "secret_key"},
		Providers:       testAccProviders,
		CheckDestroy:    testAccCheckStorageBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketWebsiteConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageBucketExists(resourceName),
					wrapWithRetries(testAccCheckStorageBucketWebsite(resourceName, "index.html", "", "", "")),
					resource.TestCheckResourceAttr(resourceName, "website.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "website.0.index_document", "index.html"),
					resource.TestCheckResourceAttr(resourceName, "website.0.error_document", ""),
					resource.TestCheckResourceAttr(resourceName, "website_endpoint", testAccWebsiteEndpoint(rInt)),
				),
			},
			{
				Config: testAccStorageBucketWebsiteConfigWithError(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageBucketExists(resourceName),
					wrapWithRetries(testAccCheckStorageBucketWebsite(resourceName, "index.html", "error.html", "", "")),
					resource.TestCheckResourceAttr(resourceName, "website.0.index_document", "index.html"),
					resource.TestCheckResourceAttr(resourceName, "website.0.error_document", "error.html"),
					resource.TestCheckResourceAttr(resourceName, "website_endpoint", testAccWebsiteEndpoint(rInt)),
				),
			},
			{
				Config: testAccStorageBucketConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageBucketExists(resourceName),
					wrapWithRetries(testAccCheckStorageBucketWebsite(resourceName, "", "", "", "")),
					resource.TestCheckResourceAttr(resourceName, "website.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "website_endpoint", ""),
				),
			},
		},
	})
}

func TestAccStorageBucket_WebsiteRedirect(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "yandex_storage_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:        func() { testAccPreCheck(t) },
		IDRefreshName:   resourceName,
		IDRefreshIgnore: []string{"access_key", "secret_key"},
		Providers:       testAccProviders,
		CheckDestroy:    testAccCheckStorageBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketWebsiteConfigWithRedirect(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageBucketExists(resourceName),
					testAccCheckStorageBucketWebsite(resourceName, "", "", "", "hashicorp.com?my=query"),
					resource.TestCheckResourceAttr(resourceName, "website_endpoint", testAccWebsiteEndpoint(rInt)),
				),
			},
		},
	})
}

func TestAccStorageBucket_WebsiteHttpsRedirect(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "yandex_storage_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:        func() { testAccPreCheck(t) },
		IDRefreshName:   resourceName,
		IDRefreshIgnore: []string{"access_key", "secret_key"},
		Providers:       testAccProviders,
		CheckDestroy:    testAccCheckStorageBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketWebsiteConfigWithHttpsRedirect(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageBucketExists(resourceName),
					testAccCheckStorageBucketWebsite(resourceName, "", "", "https", "hashicorp.com?my=query"),
					resource.TestCheckResourceAttr(resourceName, "website_endpoint", testAccWebsiteEndpoint(rInt)),
				),
			},
		},
	})
}

func TestAccStorageBucket_WebsiteRoutingRules(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "yandex_storage_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:        func() { testAccPreCheck(t) },
		IDRefreshName:   resourceName,
		IDRefreshIgnore: []string{"access_key", "secret_key"},
		Providers:       testAccProviders,
		CheckDestroy:    testAccCheckStorageBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketWebsiteConfigWithRoutingRules(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageBucketExists(resourceName),
					testAccCheckStorageBucketWebsite(
						resourceName, "index.html", "error.html", "", ""),
					testAccCheckStorageBucketWebsiteRoutingRules(
						resourceName,
						[]*s3.RoutingRule{
							{
								Condition: &s3.Condition{
									KeyPrefixEquals: aws.String("docs/"),
								},
								Redirect: &s3.Redirect{
									HttpRedirectCode:     aws.String("301"),
									Protocol:             aws.String("http"),
									ReplaceKeyPrefixWith: aws.String("documents/"),
								},
							},
						},
					),
					resource.TestCheckResourceAttr(resourceName, "website_endpoint", testAccWebsiteEndpoint(rInt)),
				),
			},
		},
	})
}

// Test TestAccStorageBucket_shouldFailNotFound is designed to fail with a "plan
// not empty" error in Terraform, to check against regresssions.
// See https://github.com/hashicorp/terraform/pull/2925
func TestAccStorageBucket_shouldFailNotFound(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "yandex_storage_bucket.test"

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckStorageBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketDestroyedConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageBucketExists(resourceName),
					testAccCheckStorageDestroyBucket(resourceName),
					ensureBucketDeleted(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccStorageBucket_VersioningNone(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "yandex_storage_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckStorageBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketBasic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageBucketExists(resourceName),
					testAccCheckStorageBucketVersioning(resourceName, ""),
				),
			},
		},
	})
}

func TestAccStorageBucket_VersioningEnabled(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "yandex_storage_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckStorageBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketConfigWithVersioning(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageBucketExists(resourceName),
					testAccCheckStorageBucketVersioning(resourceName, s3.BucketVersioningStatusEnabled),
				),
			},
			{
				Config: testAccStorageBucketConfigWithDisableVersioning(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageBucketExists(resourceName),
					testAccCheckStorageBucketVersioning(resourceName, s3.BucketVersioningStatusSuspended),
				),
			},
		},
	})
}

func TestAccStorageBucket_cors_update(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "yandex_storage_bucket.test"

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckStorageBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketConfigWithCORS(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageBucketExists(resourceName),
					wrapWithRetries(testAccCheckStorageBucketCors(
						resourceName,
						[]*s3.CORSRule{
							{
								AllowedHeaders: []*string{aws.String("*")},
								AllowedMethods: []*string{aws.String("PUT"), aws.String("POST")},
								AllowedOrigins: []*string{aws.String("https://www.example.com")},
								ExposeHeaders:  []*string{aws.String("x-amz-server-side-encryption"), aws.String("ETag")},
								MaxAgeSeconds:  aws.Int64(3000),
							},
						},
					)),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_headers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_headers.0", "*"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_methods.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_methods.0", "PUT"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_methods.1", "POST"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_origins.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_origins.0", "https://www.example.com"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.max_age_seconds", "3000"),
				),
			},
			{
				Config: testAccStorageBucketConfigWithCORSUpdated(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageBucketExists(resourceName),
					wrapWithRetries(testAccCheckStorageBucketCors(
						resourceName,
						[]*s3.CORSRule{
							{
								AllowedHeaders: []*string{aws.String("*")},
								AllowedMethods: []*string{aws.String("GET")},
								AllowedOrigins: []*string{aws.String("https://www.example.ru")},
								ExposeHeaders:  []*string{aws.String("x-amz-server-side-encryption"), aws.String("ETag")},
								MaxAgeSeconds:  aws.Int64(2000),
							},
						},
					)),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_headers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_headers.0", "*"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_methods.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_methods.0", "GET"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_origins.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_origins.0", "https://www.example.ru"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.max_age_seconds", "2000"),
				),
			},
		},
	})
}

func TestAccStorageBucket_MaxSize(t *testing.T) {
	const (
		resourceName = "yandex_storage_bucket.test"
		maxSize      = 1024
	)

	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckStorageBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketMaxSize(rInt, maxSize),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "max_size", strconv.Itoa(maxSize)),
				),
			},
		},
	})
}

func TestAccStorageBucket_HTTPSConfig(t *testing.T) {
	const resourceName = "yandex_storage_bucket.test"

	externalCertificateID := os.Getenv("STORAGE_TEST_CERTIFICATE_ID")
	if externalCertificateID == "" {
		t.Logf("STORAGE_TEST_CERTIFICATE_ID not provided for test")
		t.Skip()
	}

	bucketName := os.Getenv("STORAGE_CERTIFICATE_BUCKET_NAME")
	if bucketName == "" {
		t.Logf("STORAGE_CERTIFICATE_BUCKET_NAME not provided for test")
		t.Skip()
	}

	resource.Test(t, resource.TestCase{
		PreCheck:        func() { testAccPreCheck(t) },
		IDRefreshName:   resourceName,
		IDRefreshIgnore: []string{"access_key", "secret_key"},
		Providers:       testAccProviders,
		CheckDestroy:    testAccCheckStorageBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketHTTPSConfig(bucketName, externalCertificateID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "https.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "https.0.certificate_id", externalCertificateID),
				),
			}, {
				Config: testAccStorageBucketWithCustomName(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "https.#", "0"),
				),
			},
		},
	})
}

func TestAccStorageBucket_AnonymousAccessFlags(t *testing.T) {
	const resourceName = "yandex_storage_bucket.test"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckStorageBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketAnonymousAccessFlags(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "anonymous_access_flags.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "anonymous_access_flags.0.read", "true"),
					resource.TestCheckResourceAttr(resourceName, "anonymous_access_flags.0.list", "true"),
				),
			},
		},
	})
}

func TestAccStorageBucket_StorageClass(t *testing.T) {
	const resourceName = "yandex_storage_bucket.test"

	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckStorageBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketDefaultStorageClassCold(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "default_storage_class", "COLD"),
				),
			},
		},
	})
}

func TestAccStorageBucket_FolderID(t *testing.T) {
	const resourceName = "yandex_storage_bucket.test"

	rInt := acctest.RandInt()
	folderID := testFolderID

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckStorageBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketWithFolderID(rInt, folderID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "folder_id", folderID),
				),
			},
		},
	})
}

func TestAccStorageBucket_cors_delete(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "yandex_storage_bucket.test"

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckStorageBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketConfigWithCORS(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageBucketExists(resourceName),
					wrapWithRetries(testAccCheckStorageBucketCors(
						resourceName,
						[]*s3.CORSRule{
							{
								AllowedHeaders: []*string{aws.String("*")},
								AllowedMethods: []*string{aws.String("PUT"), aws.String("POST")},
								AllowedOrigins: []*string{aws.String("https://www.example.com")},
								ExposeHeaders:  []*string{aws.String("x-amz-server-side-encryption"), aws.String("ETag")},
								MaxAgeSeconds:  aws.Int64(3000),
							},
						},
					)),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_headers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_headers.0", "*"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_methods.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_methods.0", "PUT"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_methods.1", "POST"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_origins.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_origins.0", "https://www.example.com"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.max_age_seconds", "3000"),
				),
			},
			{
				Config: testAccStorageBucketConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageBucketExists(resourceName),
					wrapWithRetries(testAccCheckStorageBucketCors(resourceName, nil)),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.#", "0"),
				),
			},
		},
	})
}

func TestAccStorageBucket_cors_emptyOrigin(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "yandex_storage_bucket.test"

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckStorageBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketConfigWithCORSEmptyOrigin(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageBucketExists(resourceName),
					testAccCheckStorageBucketCors(resourceName,
						[]*s3.CORSRule{
							{
								AllowedHeaders: []*string{aws.String("*")},
								AllowedMethods: []*string{aws.String("PUT"), aws.String("POST")},
								AllowedOrigins: []*string{aws.String("")},
								ExposeHeaders:  []*string{aws.String("x-amz-server-side-encryption"), aws.String("ETag")},
								MaxAgeSeconds:  aws.Int64(3000),
							},
						},
					),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_headers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_headers.0", "*"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_methods.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_methods.0", "PUT"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_methods.1", "POST"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_origins.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.allowed_origins.0", ""),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.0.max_age_seconds", "3000"),
				),
			},
		},
	})
}

func TestAccStorageBucket_SSE(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "yandex_storage_bucket.test"

	keyName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	var symmetricKey kms.SymmetricKey

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckStorageBucketSSEDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketSSEDefault(keyName, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKMSSymmetricKeyExists(
						"yandex_kms_symmetric_key.key-a", &symmetricKey),
					testAccCheckStorageBucketExists(resourceName),
					testAccCheckStorageBucketSSE(resourceName,
						&s3.ServerSideEncryptionConfiguration{
							Rules: []*s3.ServerSideEncryptionRule{
								{
									ApplyServerSideEncryptionByDefault: &s3.ServerSideEncryptionByDefault{
										KMSMasterKeyID: &symmetricKey.Id,
										SSEAlgorithm:   aws.String(s3.ServerSideEncryptionAwsKms),
									},
								},
							},
						},
					),
				),
			},
		},
	})
}

func TestStorageBucketName(t *testing.T) {
	validNames := []string{
		"foobar",
		"127.0.0.1",
		"foo..bar",
		"foo.bar.baz",
		"Foo.Bar",
		strings.Repeat("x", 63),
	}

	for _, v := range validNames {
		if err := validateS3BucketName(v); err != nil {
			t.Fatalf("%q should be a valid storage bucket name", v)
		}
	}

	invalidNames := []string{
		"foo_bar",
		"foo_bar_baz",
		"foo;bar",
		strings.Repeat("x", 64),
	}

	for _, v := range invalidNames {
		if err := validateS3BucketName(v); err == nil {
			t.Fatalf("%q should not be a valid storage bucket name", v)
		}
	}
}

func testAccCheckStorageBucketDestroy(s *terraform.State) error {
	return testAccCheckStorageBucketDestroyWithProvider(s, testAccProvider)
}

func testAccCheckStorageBucketSSEDestroy(s *terraform.State) error {
	err := testAccCheckStorageBucketDestroyWithProvider(s, testAccProvider)
	if err != nil {
		return err
	}
	return testAccCheckKMSSymmetricKeyDestroy(s)
}

func testAccCheckStorageBucketDestroyWithProvider(s *terraform.State, provider *schema.Provider) error {
	config := provider.Meta().(*Config)

	check := func(rs *terraform.ResourceState) error {
		// access and secret keys should be destroyed too and defaults may be not provided, so create temporary ones
		ak, sak, cleanup, err := createTemporaryStaticAccessKey("editor", config)
		if err != nil {
			return err
		}
		defer cleanup()

		conn, err := getS3ClientByKeys(ak, sak, config)
		if err != nil {
			return err
		}

		_, err = conn.DeleteBucket(&s3.DeleteBucketInput{
			Bucket: aws.String(rs.Primary.ID),
		})
		if err != nil {
			if isAWSErr(err, s3.ErrCodeNoSuchBucket, "") {
				return nil
			}
			return err
		}

		return nil
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_storage_bucket" {
			continue
		}

		err := check(rs)
		if err != nil {
			return err
		}
	}
	return nil
}

func testAccCheckStorageBucketExists(n string) resource.TestCheckFunc {
	return testAccCheckStorageBucketExistsWithProvider(n, func() *schema.Provider { return testAccProvider })
}

func testAccCheckStorageBucketExistsWithProvider(n string, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		provider := providerF()

		conn, err := getS3ClientByKeys(rs.Primary.Attributes["access_key"], rs.Primary.Attributes["secret_key"],
			provider.Meta().(*Config))
		if err != nil {
			return err
		}

		_, err = conn.HeadBucket(&s3.HeadBucketInput{
			Bucket: aws.String(rs.Primary.ID),
		})
		if err != nil {
			if isAWSErr(err, s3.ErrCodeNoSuchBucket, "") {
				return fmt.Errorf("bucket not found")
			}
			return err
		}

		return nil
	}
}

func testAccCheckStorageDestroyBucket(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no storage bucket ID is set")
		}

		conn, err := getS3ClientByKeys(rs.Primary.Attributes["access_key"], rs.Primary.Attributes["secret_key"],
			testAccProvider.Meta().(*Config))
		if err != nil {
			return err
		}

		_, err = conn.DeleteBucket(&s3.DeleteBucketInput{
			Bucket: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return fmt.Errorf("error destroying bucket (%s) in testAccCheckStorageDestroyBucket: %s", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccCheckStorageBucketPolicy(n string, policy string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn, err := getS3ClientByKeys(rs.Primary.Attributes["access_key"], rs.Primary.Attributes["secret_key"],
			testAccProvider.Meta().(*Config))
		if err != nil {
			return err
		}

		out, err := conn.GetBucketPolicy(&s3.GetBucketPolicyInput{
			Bucket: aws.String(rs.Primary.ID),
		})

		if policy == "" {
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NoSuchBucketPolicy" {
				// expected
				return nil
			}
			if err == nil {
				return fmt.Errorf("Expected no policy, got: %#v", *out.Policy)
			} else {
				return fmt.Errorf("GetBucketPolicy error: %v, expected %s", err, policy)
			}
		}
		if err != nil {
			return fmt.Errorf("GetBucketPolicy error: %v, expected %s", err, policy)
		}

		if v := out.Policy; v == nil {
			if policy != "" {
				return fmt.Errorf("bad policy, found nil, expected: %s", policy)
			}
		} else {
			expected := make(map[string]interface{})
			if err := json.Unmarshal([]byte(policy), &expected); err != nil {
				return err
			}
			actual := make(map[string]interface{})
			if err := json.Unmarshal([]byte(*v), &actual); err != nil {
				return err
			}

			if !reflect.DeepEqual(expected, actual) {
				return fmt.Errorf("bad policy, expected: %#v, got %#v", expected, actual)
			}
		}

		return nil
	}
}

func testAccCheckStorageBucketWebsite(n string, indexDoc string, errorDoc string, redirectProtocol string, redirectTo string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn, err := getS3ClientByKeys(rs.Primary.Attributes["access_key"], rs.Primary.Attributes["secret_key"],
			testAccProvider.Meta().(*Config))
		if err != nil {
			return err
		}

		out, err := conn.GetBucketWebsite(&s3.GetBucketWebsiteInput{
			Bucket: aws.String(rs.Primary.ID),
		})
		if err != nil {
			if indexDoc == "" {
				// If we want to assert that the website is not there, than
				// this error is expected
				return nil
			}

			return fmt.Errorf("S3BucketWebsite error: %v", err)
		}

		if v := out.IndexDocument; v == nil {
			if indexDoc != "" {
				return fmt.Errorf("bad index doc, found nil, expected: %s", indexDoc)
			}
		} else {
			if *v.Suffix != indexDoc {
				return fmt.Errorf("bad index doc, expected: %s, got %#v", indexDoc, out.IndexDocument)
			}
		}

		if v := out.ErrorDocument; v == nil {
			if errorDoc != "" {
				return fmt.Errorf("bad error doc, found nil, expected: %s", errorDoc)
			}
		} else {
			if *v.Key != errorDoc {
				return fmt.Errorf("bad error doc, expected: %s, got %#v", errorDoc, out.ErrorDocument)
			}
		}

		if v := out.RedirectAllRequestsTo; v == nil {
			if redirectTo != "" {
				return fmt.Errorf("bad redirect to, found nil, expected: %s", redirectTo)
			}
		} else {
			if *v.HostName != redirectTo {
				return fmt.Errorf("bad redirect to, expected: %s, got %#v", redirectTo, out.RedirectAllRequestsTo)
			}
			if redirectProtocol != "" && v.Protocol != nil && *v.Protocol != redirectProtocol {
				return fmt.Errorf("bad redirect protocol to, expected: %s, got %#v", redirectProtocol, out.RedirectAllRequestsTo)
			}
		}

		return nil
	}
}

func testAccCheckStorageBucketWebsiteRoutingRules(n string, routingRules []*s3.RoutingRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn, err := getS3ClientByKeys(rs.Primary.Attributes["access_key"], rs.Primary.Attributes["secret_key"],
			testAccProvider.Meta().(*Config))
		if err != nil {
			return err
		}

		out, err := conn.GetBucketWebsite(&s3.GetBucketWebsiteInput{
			Bucket: aws.String(rs.Primary.ID),
		})

		if err != nil {
			if routingRules == nil {
				return nil
			}
			return fmt.Errorf("GetBucketWebsite error: %v", err)
		}

		if !reflect.DeepEqual(out.RoutingRules, routingRules) {
			return fmt.Errorf("bad routing rule, expected: %v, got %v", routingRules, out.RoutingRules)
		}

		return nil
	}
}

func testAccCheckStorageBucketVersioning(n string, versioningStatus string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn, err := getS3ClientByKeys(rs.Primary.Attributes["access_key"], rs.Primary.Attributes["secret_key"],
			testAccProvider.Meta().(*Config))
		if err != nil {
			return err
		}
		out, err := conn.GetBucketVersioning(&s3.GetBucketVersioningInput{
			Bucket: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return fmt.Errorf("GetBucketVersioning error: %v", err)
		}

		if v := out.Status; v == nil {
			if versioningStatus != "" {
				return fmt.Errorf("bad error versioning status, found nil, expected: %s", versioningStatus)
			}
		} else {
			if *v != versioningStatus {
				return fmt.Errorf("bad error versioning status, expected: %s, got %s", versioningStatus, *v)
			}
		}

		return nil
	}
}

func testAccCheckStorageBucketCors(n string, corsRules []*s3.CORSRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn, err := getS3ClientByKeys(rs.Primary.Attributes["access_key"], rs.Primary.Attributes["secret_key"],
			testAccProvider.Meta().(*Config))
		if err != nil {
			return err
		}

		out, err := conn.GetBucketCors(&s3.GetBucketCorsInput{
			Bucket: aws.String(rs.Primary.ID),
		})
		if err != nil && !isAWSErr(err, "NoSuchCORSConfiguration", "") {
			return fmt.Errorf("func GetBucketCors error: %v", err)
		}

		if !reflect.DeepEqual(out.CORSRules, corsRules) {
			return fmt.Errorf("bad error cors rule, expected: %v, got %v", corsRules, out.CORSRules)
		}

		return nil
	}
}

func testAccCheckStorageBucketLogging(n, b, p string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn, err := getS3ClientByKeys(rs.Primary.Attributes["access_key"], rs.Primary.Attributes["secret_key"],
			testAccProvider.Meta().(*Config))
		if err != nil {
			return err
		}

		out, err := conn.GetBucketLogging(&s3.GetBucketLoggingInput{
			Bucket: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return fmt.Errorf("GetBucketLogging error: %v", err)
		}

		if out.LoggingEnabled == nil {
			return fmt.Errorf("logging not enabled for bucket: %s", rs.Primary.ID)
		}

		tb := s.RootModule().Resources[b]

		if v := out.LoggingEnabled.TargetBucket; v == nil {
			if tb.Primary.ID != "" {
				return fmt.Errorf("bad target bucket, found nil, expected: %s", tb.Primary.ID)
			}
		} else {
			if *v != tb.Primary.ID {
				return fmt.Errorf("bad target bucket, expected: %s, got %s", tb.Primary.ID, *v)
			}
		}

		if v := out.LoggingEnabled.TargetPrefix; v == nil {
			if p != "" {
				return fmt.Errorf("bad target prefix, found nil, expected: %s", p)
			}
		} else {
			if *v != p {
				return fmt.Errorf("bad target prefix, expected: %s, got %s", p, *v)
			}
		}

		return nil
	}
}

func testAccCheckStorageBucketSSE(n string, config *s3.ServerSideEncryptionConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn, err := getS3ClientByKeys(rs.Primary.Attributes["access_key"], rs.Primary.Attributes["secret_key"],
			testAccProvider.Meta().(*Config))
		if err != nil {
			return err
		}

		out, err := conn.GetBucketEncryption(&s3.GetBucketEncryptionInput{
			Bucket: aws.String(rs.Primary.ID),
		})
		if err != nil && !isAWSErr(err, "NoSuchEncryptionConfiguration", "") {
			return fmt.Errorf("func GetBucketCors error: %v", err)
		}

		if !reflect.DeepEqual(out.ServerSideEncryptionConfiguration, config) {
			return fmt.Errorf("bad error cors rule, expected: %v, got %v", config, out.ServerSideEncryptionConfiguration)
		}

		return nil
	}
}

//// These need a bit of randomness as the name can only be used once globally
func testAccBucketName(randInt int) string {
	return fmt.Sprintf("tf-test-bucket-%d", randInt)
}

func testAccBucketDomainName(randInt int) string {
	name, _ := bucketDomainName(fmt.Sprintf("tf-test-bucket-%d", randInt), getExampleStorageEndpoint())
	return name
}

func testAccWebsiteEndpoint(randInt int) string {
	return fmt.Sprintf("tf-test-bucket-%d.%s", randInt, WebsiteDomainURL())
}

func newBucketConfigBuilder(randInt int) testAccStorageBucketConfigBuilder {
	const (
		defaultStorageClass  = "STANDARD"
		defaultAnonymousRead = false
		defaultAnonymousList = false
	)
	return testAccStorageBucketConfigBuilder{
		bucketRandomNumber: randInt,
		storageClass:       defaultStorageClass,
		anonymousRead:      defaultAnonymousRead,
		anonymousList:      defaultAnonymousList,
	}
}

const (
	testAccStorageBucketConfigBuilderRoleEditor = "editor"
	testAccStorageBucketConfigBuilderRoleAdmin  = "admin"
)

type testAccStorageBucketConfigBuilder struct {
	bucketRandomNumber int
	customBucketName   string

	beforeBucket     []string
	bucketStatements []string
	afterBucket      []string
	role             string

	storageClass  string
	anonymousRead bool
	anonymousList bool
}

func (b testAccStorageBucketConfigBuilder) withCustomName(name string) testAccStorageBucketConfigBuilder {
	b.customBucketName = name

	return b
}

func (b testAccStorageBucketConfigBuilder) withStorageClass(class string) testAccStorageBucketConfigBuilder {
	b.storageClass = class

	return b
}

func (b testAccStorageBucketConfigBuilder) withAnonymousAccessFlags(read, list bool) testAccStorageBucketConfigBuilder {
	b.anonymousRead = read
	b.anonymousList = list

	return b
}

func (b testAccStorageBucketConfigBuilder) addStatement(statement string) testAccStorageBucketConfigBuilder {
	b.bucketStatements = append(b.bucketStatements, "\t"+statement)

	return b
}

func (b testAccStorageBucketConfigBuilder) before(statement string) testAccStorageBucketConfigBuilder {
	b.beforeBucket = append(b.beforeBucket, statement)

	return b
}

func (b testAccStorageBucketConfigBuilder) asEditor() testAccStorageBucketConfigBuilder {
	b.role = testAccStorageBucketConfigBuilderRoleEditor

	return b
}

func (b testAccStorageBucketConfigBuilder) asAdmin() testAccStorageBucketConfigBuilder {
	b.role = testAccStorageBucketConfigBuilderRoleAdmin

	return b
}

/*
render creates new bucket config. For visual representation, note the following
example of how it might look after calling this method:

resource "yandex_storage_bucket" "test" {
	bucket = "tf-test-bucket-%d"

	access_key = yandex_iam_service_account_static_access_key.sa-key.access_key
	secret_key = yandex_iam_service_account_static_access_key.sa-key.secret_key

	default_storage_class = "STANDARD"

	anonymous_access_flags {
		list = false
		read = false
	}

	{ bucket statements on each line }
}

{ after bucket statements on each line }

{ editor / admin IAM config if set }
*/
func (b testAccStorageBucketConfigBuilder) render() string {
	const (
		bucketNameTemplate = "tf-test-bucket-%d"
		baseTemplate       = `resource "yandex_storage_bucket" "test" {
	bucket = "%s"

	access_key = yandex_iam_service_account_static_access_key.sa-key.access_key
	secret_key = yandex_iam_service_account_static_access_key.sa-key.secret_key`
		extendedTemplate = `
	default_storage_class = %s

	anonymous_access_flags {
		list = %t
		read = %t
	}`
	)

	var bucketName string
	if b.customBucketName != "" {
		bucketName = b.customBucketName
	} else {
		bucketName = fmt.Sprintf(bucketNameTemplate, b.bucketRandomNumber)
	}

	var out strings.Builder
	if len(b.beforeBucket) > 0 {
		out.WriteString(strings.Join(b.beforeBucket, "\n"))
		out.WriteString("\n")
	}

	out.WriteString(fmt.Sprintf(baseTemplate, bucketName))
	out.WriteString("\n")

	out.WriteString(fmt.Sprintf(
		extendedTemplate,
		strconv.Quote(b.storageClass),
		b.anonymousList,
		b.anonymousRead,
	))
	out.WriteString("\n")

	if len(b.bucketStatements) > 0 {
		out.WriteString(strings.Join(b.bucketStatements, "\n"))
		out.WriteString("\n")
	}

	out.WriteString("}\n")

	if len(b.afterBucket) > 0 {
		out.WriteString(strings.Join(b.afterBucket, "\n"))
		out.WriteString("\n")
	}

	switch b.role {
	case testAccStorageBucketConfigBuilderRoleEditor:
		out.WriteString(testAccCommonIamDependenciesEditorConfig(b.bucketRandomNumber))
	case testAccStorageBucketConfigBuilderRoleAdmin:
		out.WriteString(testAccCommonIamDependenciesAdminConfig(b.bucketRandomNumber))
	}

	return out.String()
}

func testAccStorageBucketConfig(randInt int) string {
	return newBucketConfigBuilder(randInt).
		asEditor().
		render()
}

func testAccStorageBucketAclPreConfig(randInt int) string {
	const acl = `acl = "public-read"`

	return newBucketConfigBuilder(randInt).
		addStatement(acl).
		asAdmin().
		render()
}

func testAccStorageBucketAclPostConfig(randInt int) string {
	const acl = `acl = "private"`

	return newBucketConfigBuilder(randInt).
		addStatement(acl).
		asAdmin().
		render()
}

func testAccStorageBucketWebsiteConfig(randInt int) string {
	const website = `website {
		index_document = "index.html"
	}`

	return newBucketConfigBuilder(randInt).
		addStatement(website).
		asEditor().
		render()
}

func testAccStorageBucketWebsiteConfigWithError(randInt int) string {
	const website = `website {
		index_document = "index.html"
		error_document = "error.html"
	}`

	return newBucketConfigBuilder(randInt).
		addStatement(website).
		asEditor().
		render()
}

func testAccStorageBucketWebsiteConfigWithRedirect(randInt int) string {
	const website = `website {
		redirect_all_requests_to = "http://hashicorp.com?my=query"
	}`

	return newBucketConfigBuilder(randInt).
		addStatement(website).
		asEditor().
		render()
}

func testAccStorageBucketWebsiteConfigWithHttpsRedirect(randInt int) string {
	const website = `website {
		redirect_all_requests_to = "https://hashicorp.com?my=query"
	}`

	return newBucketConfigBuilder(randInt).
		addStatement(website).
		asEditor().
		render()
}

func testAccStorageBucketWebsiteConfigWithRoutingRules(randInt int) string {
	const website = `website {
		index_document = "index.html"
		error_document = "error.html"

		routing_rules = <<EOF
		[
			{
				"Condition": {
					"KeyPrefixEquals": "docs/"
				},
				"Redirect": {
					"Protocol": "http",
					"HttpRedirectCode": "301",
					"ReplaceKeyPrefixWith": "documents/"
				}
			}
		]
		EOF
	}`

	return newBucketConfigBuilder(randInt).
		addStatement(website).
		asEditor().
		render()
}

func testAccStorageBucketDestroyedConfig(randInt int) string {
	return newBucketConfigBuilder(randInt).
		asEditor().
		render()
}

func testAccStorageBucketConfigWithVersioning(randInt int) string {
	const versioning = `versioning {
		enabled = true
	}`
	return newBucketConfigBuilder(randInt).
		addStatement(versioning).
		asAdmin().
		render()
}

func testAccStorageBucketConfigWithDisableVersioning(randInt int) string {
	const versioning = `versioning {
		enabled = false
	}`

	return newBucketConfigBuilder(randInt).
		addStatement(versioning).
		asAdmin().
		render()
}

func testAccStorageBucketConfigWithCORS(randInt int) string {
	const cors = `cors_rule {
		allowed_headers = ["*"]
		allowed_methods = ["PUT","POST"]
		allowed_origins = ["https://www.example.com"]
		expose_headers  = ["x-amz-server-side-encryption","ETag"]
		max_age_seconds = 3000
	}`

	return newBucketConfigBuilder(randInt).
		addStatement(cors).
		asEditor().
		render()
}

func testAccStorageBucketConfigWithCORSUpdated(randInt int) string {
	const cors = `cors_rule {
		allowed_headers = ["*"]
		allowed_methods = ["GET"]
		allowed_origins = ["https://www.example.ru"]
		expose_headers  = ["x-amz-server-side-encryption","ETag"]
		max_age_seconds = 2000
	}`

	return newBucketConfigBuilder(randInt).
		addStatement(cors).
		asEditor().
		render()
}

func testAccStorageBucketConfigWithCORSEmptyOrigin(randInt int) string {
	const cors = `cors_rule {
		allowed_headers = ["*"]
		allowed_methods = ["PUT","POST"]
		allowed_origins = [""]
		expose_headers = ["x-amz-server-side-encryption","ETag"]
		max_age_seconds = 3000
	}`

	return newBucketConfigBuilder(randInt).
		addStatement(cors).
		asEditor().
		render()
}

func testAccStorageBucketConfigWithNamePrefix(randInt int) string {
	// do not use render here because it use prefix here.
	return `resource "yandex_storage_bucket" "test" {
	bucket_prefix = "tf-test-"

	access_key = yandex_iam_service_account_static_access_key.sa-key.access_key
	secret_key = yandex_iam_service_account_static_access_key.sa-key.secret_key

	default_storage_class = "STANDARD"

	anonymous_access_flags {
		list = false
		read = false
	}
}
` + testAccCommonIamDependenciesEditorConfig(randInt)
}

func testAccStorageBucketConfigWithGeneratedName(randInt int) string {
	// do not use render here because name will be generated.
	return `resource "yandex_storage_bucket" "test" {
	access_key = yandex_iam_service_account_static_access_key.sa-key.access_key
	secret_key = yandex_iam_service_account_static_access_key.sa-key.secret_key

	default_storage_class = "STANDARD"

	anonymous_access_flags {
		list = false
		read = false
	}
}
` + testAccCommonIamDependenciesEditorConfig(randInt)
}

func testAccStorageBucketConfigWithLogging(randInt int) string {
	const stmt = `logging {
    		target_bucket = yandex_storage_bucket.log_bucket.id
		target_prefix = "log/"
  	}`

	before := fmt.Sprintf(`resource "yandex_storage_bucket" "log_bucket" {
  	bucket = "tf-test-bucket-%[1]d-log"

	access_key = yandex_iam_service_account_static_access_key.sa-key.access_key
	secret_key = yandex_iam_service_account_static_access_key.sa-key.secret_key

	default_storage_class = "STANDARD"

	anonymous_access_flags {
		list = false
		read = false
	}
}`, randInt)

	return newBucketConfigBuilder(randInt).
		before(before).
		addStatement(stmt).
		asAdmin().
		render()
}

func testAccStorageBucketConfigWithLifecycle(randInt int) string {
	const acl = `acl = "private"`
	const stmt = `lifecycle_rule {
		id      = "id1"
		prefix  = "path1/"
		enabled = true

		expiration {
			days = 365
		}
	}`

	return newBucketConfigBuilder(randInt).
		addStatement(acl).
		addStatement(stmt).
		asAdmin().
		render()
}

func testAccStorageBucketConfigWithVersioningLifecycle(randInt int) string {
	const acl = `acl = "private"`
	const lifecycle = `lifecycle_rule {
		id      = "id1"
		prefix  = "path1/"
		enabled = true

		noncurrent_version_expiration {
			days = 365
		}
	}

	lifecycle_rule {
		id      = "id2"
		prefix  = "path2/"
		enabled = false

		noncurrent_version_expiration {
			days = 365
		}
	}`

	return newBucketConfigBuilder(randInt).
		addStatement(acl).
		addStatement(lifecycle).
		asAdmin().
		render()
}

func testAccStorageBucketConfigLifecycleRuleExpirationEmptyConfigurationBlock(randInt int) string {
	const stmt = `lifecycle_rule {
	enabled = true
	id      = "id1"

	expiration {}
}`

	return newBucketConfigBuilder(randInt).
		addStatement(stmt).
		asAdmin().
		render()
}

func testAccStorageBucketConfigLifecycleRuleAbortIncompleteMultipartUploadDays(randInt int) string {
	const stmt = `lifecycle_rule {
	abort_incomplete_multipart_upload_days = 7
	enabled                                = true
	id                                     = "id1"
}`

	return newBucketConfigBuilder(randInt).
		addStatement(stmt).
		asAdmin().
		render()
}

func testAccStorageBucketSSEDefault(keyName string, randInt int) string {
	const sse = `server_side_encryption_configuration {
		rule {
			apply_server_side_encryption_by_default {
				kms_master_key_id = yandex_kms_symmetric_key.key-a.id
				sse_algorithm     = "aws:kms"
			}
		}
	}`

	before := fmt.Sprintf(`resource "yandex_kms_symmetric_key" "key-a" {
	name              = "%s"
	description       = "description for key-a"
	default_algorithm = "AES_128"
	rotation_period   = "24h"

	labels = {
		tf-label    = "tf-label-value-a"
		empty-label = ""
	}
}`, keyName)

	return newBucketConfigBuilder(randInt).
		before(before).
		addStatement(sse).
		asAdmin().
		render()
}

func testAccStorageBucketBasic(randInt int) string {
	return newBucketConfigBuilder(randInt).
		asAdmin().
		render()
}

func testAccStorageBucketPolicy(randInt int) string {
	return fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "TestPolicySid",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:*",
      "Resource": [
        "arn:aws:s3:::tf-test-bucket-%[1]d/*",
        "arn:aws:s3:::tf-test-bucket-%[1]d"
      ]
    },
    {
      "Sid": "TestPolicySid",
      "Effect": "Deny",
      "Principal": "*",
      "Action": "s3:PutObject",
      "Resource": [
        "arn:aws:s3:::tf-test-bucket-%[1]d/*",
        "arn:aws:s3:::tf-test-bucket-%[1]d"
      ]
    }
  ]
}`, randInt)
}

func testAccStorageBucketConfigWithPolicy(randInt int) string {
	const acl = `acl = "public-read"`
	policy := "policy = " + strconv.Quote(testAccStorageBucketPolicy(randInt))
	return newBucketConfigBuilder(randInt).
		addStatement(policy).
		addStatement(acl).
		asAdmin().
		render()
}

func testAccStorageBucketConfigWithEmptyPolicy(randInt int) string {
	const acl = `acl = "public-read"`

	return newBucketConfigBuilder(randInt).
		addStatement(acl).
		asAdmin().
		render()
}

func testAccStorageBucketMaxSize(randInt int, maxSize int) string {
	maxSizeStmt := fmt.Sprintf(`max_size = %d`, maxSize)

	return newBucketConfigBuilder(randInt).
		asEditor().
		addStatement(maxSizeStmt).
		render()
}

func testAccStorageBucketWithCustomName(name string) string {
	return newBucketConfigBuilder(0).
		withCustomName(name).
		asEditor().
		render()
}

func testAccStorageBucketHTTPSConfig(bucketName, certID string) string {
	httpsStmt := fmt.Sprintf(`https {
		certificate_id = "%s"
	}`, certID)

	return newBucketConfigBuilder(0).
		withCustomName(bucketName).
		asEditor().
		addStatement(httpsStmt).
		render()
}

func testAccStorageBucketDefaultStorageClassCold(randInt int) string {
	return newBucketConfigBuilder(randInt).
		asEditor().
		withStorageClass("COLD").
		render()
}

func testAccStorageBucketWithFolderID(randInt int, folderID string) string {
	folderStmt := fmt.Sprintf("folder_id = %q", folderID)

	return newBucketConfigBuilder(randInt).
		asEditor().
		addStatement(folderStmt).
		render()
}

func testAccStorageBucketAnonymousAccessFlags(randInt int) string {
	return newBucketConfigBuilder(randInt).
		asEditor().
		withAnonymousAccessFlags(true, true).
		render()
}

func TestAccStorageBucket_Logging(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "yandex_storage_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckStorageBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketConfigWithLogging(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageBucketExists(resourceName),
					testAccCheckStorageBucketLogging(resourceName, "yandex_storage_bucket.log_bucket", "log/"),
				),
			},
		},
	})
}

func TestAccStorageBucket_LifecycleBasic(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "yandex_storage_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckStorageBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketConfigWithLifecycle(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.id", "id1"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.prefix", "path1/"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.expiration.0.days", "365"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.expiration.0.date", ""),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.expiration.0.expired_object_delete_marker", "false"),
				),
			},
		},
	})
}

func TestAccStorageBucket_LifecycleVersioning(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "yandex_storage_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckStorageBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketConfigWithVersioningLifecycle(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.id", "id1"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.prefix", "path1/"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.noncurrent_version_expiration.0.days", "365"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.1.id", "id2"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.1.prefix", "path2/"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.1.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.1.noncurrent_version_expiration.0.days", "365"),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/11420
func TestAccStorageBucket_LifecycleRule_Expiration_EmptyConfigurationBlock(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "yandex_storage_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckStorageBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketConfigLifecycleRuleExpirationEmptyConfigurationBlock(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageBucketExists(resourceName),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/15138
func TestAccStorageBucket_LifecycleRule_AbortIncompleteMultipartUploadDays_NoExpiration(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "yandex_storage_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckStorageBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketConfigLifecycleRuleAbortIncompleteMultipartUploadDays(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageBucketExists(resourceName),
				),
			},
		},
	})
}

// Test yandex_storage_bucket import with policy operation
func TestAccStorageBucket_ImportBasic(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "yandex_storage_bucket.test"

	resource.Test(t, resource.TestCase{
		PreCheck:        func() { testAccPreCheck(t) },
		IDRefreshName:   resourceName,
		IDRefreshIgnore: []string{"access_key", "secret_key"},
		Providers:       testAccProviders,
		CheckDestroy:    testAccCheckStorageBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketConfigWithPolicy(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageBucketExists(resourceName),
					testAccCheckStorageBucketPolicy(resourceName, testAccStorageBucketPolicy(rInt)),
				),
			},
		},
	})
}

func wrapWithRetries(f resource.TestCheckFunc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		err := f(s)
		for i := 0; err != nil && i < 6; i++ {
			time.Sleep(time.Second * 20)
			err = f(s)
		}
		return err
	}
}

func ensureBucketDeleted(n string) resource.TestCheckFunc {
	return wrapWithRetries(func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn, err := getS3ClientByKeys(rs.Primary.Attributes["access_key"], rs.Primary.Attributes["secret_key"],
			testAccProvider.Meta().(*Config))
		if err != nil {
			return err
		}
		return checkBucketDeleted(rs.Primary.ID, conn)
	})
}

func checkBucketDeleted(ID string, conn *s3.S3) error {
	_, err := conn.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(ID),
	})

	if err == nil {
		return fmt.Errorf("expected NoSuchBucket error, got none")
	}

	awsErr, ok := err.(awserr.RequestFailure)

	if !ok {
		return fmt.Errorf("got unexpected error type: %v", err)
	}

	if awsErr.StatusCode() != 404 {
		return fmt.Errorf("expected NotFound error, got: %v", err)
	}

	return nil
}
