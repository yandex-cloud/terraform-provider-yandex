package yandex

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"testing"
	"text/template"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/audittrails/v1"
)

func init() {
	resource.AddTestSweepers("yandex_audit_trails_trail", &resource.Sweeper{
		Name: "yandex_audit_trails_trail",
		F:    testSweepAuditTrails,
	})
}

// Sweep function deletes all trails (used resources are deleted by their sweepers)
func testSweepAuditTrails(_ string) error {
	log.Printf("[DEBUG] Sweeping Audit Trails Trail")
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	result := &multierror.Error{}

	req := &audittrails.ListTrailsRequest{FolderId: conf.FolderID}
	it := conf.sdk.AuditTrails().Trail().TrailIterator(conf.Context(), req)
	for it.Next() {
		id := it.Value().GetId()

		if !sweepAuditTrailsTrail(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Audit Trails Trail %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepAuditTrailsTrail(conf *Config, id string) bool {
	return sweepWithRetry(sweepAuditTrailsTrailOnce, conf, "Audit Trails Trail", id)
}

func sweepAuditTrailsTrailOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(5 * time.Minute)
	defer cancel()

	op, err := conf.sdk.AuditTrails().Trail().Delete(ctx, &audittrails.DeleteTrailRequest{
		TrailId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

// Tests for Storage with Any/Some filters trail create/update/import/delete operations
// Requires YC_STORAGE_ACCESS_KEY/YC_STORAGE_SECRET_KEY environment variables
func TestAccResourceAuditTrailsTrail_storage(t *testing.T) {
	t.Parallel()

	saName := acctest.RandomWithPrefix("tf-acc-trail-storage-sa")
	bucketTestName := acctest.RandomWithPrefix("tf-acc-trail-initial-bucket")
	updatedBucketTestName := acctest.RandomWithPrefix("tf-acc-trail-updated-bucket")
	trailTestName := acctest.RandomWithPrefix("tf-acc-trail")

	// base config describes required resources for this test - we will reuse it to check only trail update logic
	tfBaseConfig := auditTrailsServiceAccountConfig(saName) + auditTrailsStorageResourceConfig(bucketTestName) + auditTrailsStorageResourceConfig(updatedBucketTestName)
	initialTrail := auditTrailsStorageConfig(trailTestName, bucketTestName, saName)

	updatedTrail := initialTrail
	updatedTrail.StorageDestination = trailStorageDestination{
		BucketName:   updatedBucketTestName,
		ObjectPrefix: "some-prefix",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckYandexAuditTrailsTrailAllDestroyed, // delete is called for each resource and checked
		Steps: []resource.TestStep{
			// create base infrastructure
			{
				Config: tfBaseConfig,
			},
			// create base storage destination trail with full filter
			{
				Config: tfBaseConfig + initialTrail.toTerraformResource(),
				Check:  checkTrail(initialTrail, false),
			},
			// check that it could be correctly imported
			auditTrailsTrailImportStep(trailTestName),
			// update the trail - change destination bucket
			{
				Config: tfBaseConfig + updatedTrail.toTerraformResource(),
				Check:  checkTrail(updatedTrail, false),
			},
			// check that it could be correctly imported
			auditTrailsTrailImportStep(trailTestName),
			// delete the trail to ensure that after cleaning the bucket, we won't receive any new events
			{
				Config: tfBaseConfig,
			},
			// clean used buckets before finishing the test
			{
				PreConfig: cleanBuckets,
				Config:    tfBaseConfig,
			},
		},
	})
}

func cleanBuckets() {
	err := testSweepStorageObject("")
	if err != nil {
		fmt.Printf("Failed to clean buckets: %s\n", err.Error())
	}
}

// Tests for Logging trail create/update/import/delete operations
func TestAccResourceAuditTrailsTrail_logging(t *testing.T) {
	t.Parallel()

	saName := acctest.RandomWithPrefix("tf-acc-trail-logging-sa")
	loggingTestName := acctest.RandomWithPrefix("tf-acc-trail-initial-group")
	updatedLoggingTestName := acctest.RandomWithPrefix("tf-acc-trail-updated-group")
	trailTestName := acctest.RandomWithPrefix("tf-acc-trail")

	// base config describes required resources for this test - we will reuse it to check only trail update logic
	tfBaseConfig := auditTrailsServiceAccountConfig(saName) + auditTrailsLoggingResourceConfig(loggingTestName) + auditTrailsLoggingResourceConfig(updatedLoggingTestName)
	initialTrail := auditTrailsLoggingConfig(trailTestName, loggingTestName, saName)

	updatedTrail := initialTrail
	updatedTrail.LoggingDestination = trailLoggingDestination{
		LogGroupName: updatedLoggingTestName,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckYandexAuditTrailsTrailAllDestroyed, // delete is called for each resource and checked
		Steps: []resource.TestStep{
			// create base logging infrastructure
			{
				Config: tfBaseConfig,
			},
			// create base logging destination trail with minimal filter
			{
				Config: tfBaseConfig + initialTrail.toTerraformResource(),
				Check:  checkTrail(initialTrail, false),
			},
			// check that it could be correctly imported
			auditTrailsTrailImportStep(trailTestName),
			// update the trail - change log group
			{
				Config: tfBaseConfig + updatedTrail.toTerraformResource(),
				Check:  checkTrail(updatedTrail, false),
			},
			// check that it could be correctly imported
			auditTrailsTrailImportStep(trailTestName),
		},
	})
}

func TestAccResourceAuditTrailsTrail_dataStream(t *testing.T) {
	t.Parallel()

	saName := acctest.RandomWithPrefix("tf-acc-trail-yds-sa")
	ydbTestName := acctest.RandomWithPrefix("tf-acc-trail-ydb")
	streamTestName := acctest.RandomWithPrefix("tf-acc-trail-stream")
	updatedStreamTestName := acctest.RandomWithPrefix("tf-acc-trail-updated-stream")
	trailTestName := acctest.RandomWithPrefix("tf-acc-trail")

	// base config describes required resources for this test - we will reuse it to check only trail update logic
	tfYdbConfig := auditTrailsYdbResourceConfig(ydbTestName)
	tfBaseConfig := tfYdbConfig + auditTrailsServiceAccountConfig(saName) + auditTrailsYdsResourceConfig(ydbTestName, streamTestName) + auditTrailsYdsResourceConfig(ydbTestName, updatedStreamTestName)
	initialTrail := auditTrailsYdsConfig(trailTestName, ydbTestName, streamTestName, saName)

	updatedTrail := initialTrail
	updatedTrail.YDSDestination = trailDataStreamDestination{
		YdbName:    ydbTestName,
		StreamName: updatedStreamTestName,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckYandexAuditTrailsTrailAllDestroyed, // delete is called for each resource and checked
		Steps: []resource.TestStep{
			// create YDB infrastructure
			{
				Config: tfYdbConfig,
			},
			// create base infrastructure
			{
				PreConfig: waitForYdb,
				Config:    tfBaseConfig,
			},
			// create base yds destination trail with minimal filter
			{
				Config: tfBaseConfig + initialTrail.toTerraformResource(),
				Check:  checkTrail(initialTrail, false),
			},
			// check that it could be correctly imported
			auditTrailsTrailImportStep(trailTestName),
			// update the trail - change ydb and stream name
			{
				Config: tfBaseConfig + updatedTrail.toTerraformResource(),
				Check:  checkTrail(updatedTrail, false),
			},
			// check that it could be correctly imported
			auditTrailsTrailImportStep(trailTestName),
		},
	})
}

// we will wait until YDB cluster is ready to be used
// this is needed because serverless_database resource is created when DB is transferred into PROVISIONING state
func waitForYdb() {
	time.Sleep(1 * time.Minute)
}

func checkTrail(trail yandexAuditTrailsTrail, dataSourceCheck bool) resource.TestCheckFunc {
	resourceName := "yandex_audit_trails_trail." + trail.Name
	if dataSourceCheck {
		resourceName = "data." + resourceName
	}

	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr(resourceName, "name", trail.Name),
		resource.TestCheckResourceAttr(resourceName, "description", trail.Description),
		resolveAndCheckSaID(resourceName, trail.ServiceAccountName),
		resource.TestCheckResourceAttr(resourceName, "folder_id", trail.FolderID),
	}

	for key, value := range trail.Labels {
		checks = append(checks, resource.TestCheckResourceAttr(resourceName, "labels."+key, value))
	}

	if storageConfig := trail.StorageDestination; storageConfig != (trailStorageDestination{}) {
		checks = append(checks, resource.TestCheckResourceAttr(resourceName,
			"storage_destination.0.bucket_name", storageConfig.BucketName))
		checks = append(checks, resource.TestCheckResourceAttr(resourceName,
			"storage_destination.0.object_prefix", storageConfig.ObjectPrefix))
	} else {
		checks = append(checks, resource.TestCheckResourceAttr(resourceName, "storage_destination.#", "0"))
	}

	if loggingConfig := trail.LoggingDestination; loggingConfig != (trailLoggingDestination{}) {
		checks = append(checks, resolveAndCheckLogGroupID(resourceName, trail.LoggingDestination.LogGroupName))
	} else {
		checks = append(checks, resource.TestCheckResourceAttr(resourceName, "logging_destination.#", "0"))
	}

	if dataStreamConfig := trail.YDSDestination; dataStreamConfig != (trailDataStreamDestination{}) {
		checks = append(checks, resolveAndCheckYdbID(resourceName, trail.YDSDestination.YdbName))
		checks = append(checks, resource.TestCheckResourceAttr(resourceName,
			"data_stream_destination.0.stream_name", dataStreamConfig.StreamName))
	} else {
		checks = append(checks, resource.TestCheckResourceAttr(resourceName, "data_stream_destination.#", "0"))
	}

	managementFilter := trail.FilteringPolicy.ManagementFilter
	checks = append(checks, checkResourceScopes(resourceName, "filtering_policy.0.management_events_filter.0.resource_scope.", managementFilter.ResourceScope)...)

	dataEventFilters := trail.FilteringPolicy.DataEventFilters
	for i, dataEventFilter := range dataEventFilters {
		statePrefix := fmt.Sprintf("filtering_policy.0.data_events_filter.%d.", i)
		checks = append(checks, checkDataEventFilter(resourceName, statePrefix, dataEventFilter)...)
	}

	return resource.ComposeTestCheckFunc(checks...)
}

func checkDataEventFilter(resourceName string, statePrefix string, dataEventFilter trailDataEventFilter) []resource.TestCheckFunc {
	var checks []resource.TestCheckFunc

	checks = append(checks, resource.TestCheckResourceAttr(resourceName, statePrefix+"service", dataEventFilter.Service))
	checks = append(checks, checkResourceScopes(resourceName, statePrefix+"resource_scope.", dataEventFilter.ResourceScope)...)

	for i, includedEvent := range dataEventFilter.IncludedEvents {
		statePath := fmt.Sprintf("%sincluded_events.%d", statePrefix, i)
		checks = append(checks, resource.TestCheckResourceAttr(resourceName, statePath, includedEvent))
	}

	for i, excludedEvent := range dataEventFilter.ExcludedEvents {
		statePath := fmt.Sprintf("%sexcluded_events.%d", statePrefix, i)
		checks = append(checks, resource.TestCheckResourceAttr(resourceName, statePath, excludedEvent))
	}

	return checks
}

func checkResourceScopes(resourceName string, statePrefix string, expected []trailResourceEntry) []resource.TestCheckFunc {
	var checks []resource.TestCheckFunc

	resourceEntryLen := len(expected)
	checks = append(checks, resource.TestCheckResourceAttr(resourceName, statePrefix+"#", fmt.Sprint(resourceEntryLen)))

	for i := 0; i < resourceEntryLen; i++ {
		nestedResourceStatePrefix := fmt.Sprintf("%s%d.", statePrefix, i)
		checks = append(checks, checkResource(nestedResourceStatePrefix, resourceName, expected[i])...)
	}

	return checks
}

func checkResource(statePrefix string, resourceName string, resourceEntry trailResourceEntry) []resource.TestCheckFunc {
	var checks []resource.TestCheckFunc

	checks = append(checks, resource.TestCheckResourceAttr(resourceName, statePrefix+"resource_id", resourceEntry.ResourceId))
	checks = append(checks, resource.TestCheckResourceAttr(resourceName, statePrefix+"resource_type", resourceEntry.ResourceType))
	return checks
}

func getResourceAttr(state *terraform.State, resourceName, attrName string) (string, error) {
	resourceState, ok := state.RootModule().Resources[resourceName]
	if !ok {
		return "", fmt.Errorf("can't find %s resource", resourceName)
	}

	attrValue, ok := resourceState.Primary.Attributes[attrName]
	if !ok {
		return "", fmt.Errorf("can't find '%s' attr for %s resource", attrName, resourceName)
	}
	return attrValue, nil
}

func resolveAndCheckSaID(trailResource, serviceAccountName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		serviceAccountID, err := getResourceAttr(state, "yandex_iam_service_account."+serviceAccountName, "id")
		if err != nil {
			return err
		}

		trailSaID, err := getResourceAttr(state, trailResource, "service_account_id")
		if err != nil {
			return err
		}

		if serviceAccountID != trailSaID {
			return fmt.Errorf("service account IDs from main state and trail state do not match: %s %s",
				serviceAccountID, trailSaID)
		}
		return nil
	}
}

func resolveAndCheckLogGroupID(trailResource, logGroupName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		logGroupID, err := getResourceAttr(state, "yandex_logging_group."+logGroupName, "id")
		if err != nil {
			return err
		}

		trailLogGroupID, err := getResourceAttr(state, trailResource, "logging_destination.0.log_group_id")
		if err != nil {
			return err
		}

		if logGroupID != trailLogGroupID {
			return fmt.Errorf("log group IDs from main state and trail state do not match: %s %s",
				logGroupID, trailLogGroupID)
		}
		return nil
	}
}

func resolveAndCheckYdbID(trailResource, ydbName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		databaseID, err := getResourceAttr(state, "yandex_ydb_database_serverless."+ydbName, "id")
		if err != nil {
			return err
		}

		trailDatabaseID, err := getResourceAttr(state, trailResource, "data_stream_destination.0.database_id")
		if err != nil {
			return err
		}

		if databaseID != trailDatabaseID {
			return fmt.Errorf("database IDs from main state and trail state do not match: %s %s",
				databaseID, trailDatabaseID)
		}
		return nil
	}
}

func auditTrailsYdsConfig(trailResourceName, ydbName, streamName, saName string) yandexAuditTrailsTrail {
	return yandexAuditTrailsTrail{
		Name:               trailResourceName,
		FolderID:           getExampleFolderID(),
		Description:        "some desc",
		Labels:             map[string]string{"a": "b"},
		ServiceAccountName: saName,
		YDSDestination: trailDataStreamDestination{
			YdbName:    ydbName,
			StreamName: streamName,
		},
		FilteringPolicy: trailFilteringPolicy{
			ManagementFilter: trailManagementFilter{
				ResourceScope: []trailResourceEntry{
					{
						ResourceId:   getExampleFolderID(),
						ResourceType: "resource-manager.folder",
					},
				},
			},
			DataEventFilters: []trailDataEventFilter{},
		},
	}
}

func auditTrailsLoggingConfig(trailResourceName, logGroupName, saName string) yandexAuditTrailsTrail {
	return yandexAuditTrailsTrail{
		Name:               trailResourceName,
		FolderID:           getExampleFolderID(),
		Description:        "some desc",
		Labels:             map[string]string{"a": "b"},
		ServiceAccountName: saName,
		LoggingDestination: trailLoggingDestination{
			LogGroupName: logGroupName,
		},
		FilteringPolicy: trailFilteringPolicy{
			ManagementFilter: trailManagementFilter{
				ResourceScope: []trailResourceEntry{
					{
						ResourceId:   getExampleFolderID(),
						ResourceType: "resource-manager.folder",
					},
				},
			},
			DataEventFilters: []trailDataEventFilter{
				{
					Service: "kms",
					ResourceScope: []trailResourceEntry{
						{
							ResourceId:   getExampleFolderID(),
							ResourceType: "resource-manager.folder",
						},
					},
					IncludedEvents: []string{
						"yandex.cloud.audit.kms.Encrypt",
						"yandex.cloud.audit.kms.Decrypt",
					},
				},
				{
					Service: "iam",
					ResourceScope: []trailResourceEntry{
						{
							ResourceId:   getExampleFolderID(),
							ResourceType: "resource-manager.folder",
						},
					},
					ExcludedEvents: []string{
						"yandex.cloud.audit.iam.oslogin.CheckSshPolicy",
					},
				},
			},
		},
	}
}

func auditTrailsStorageConfig(trailResourceName, bucketName, saName string) yandexAuditTrailsTrail {
	return yandexAuditTrailsTrail{
		Name:               trailResourceName,
		FolderID:           getExampleFolderID(),
		Description:        "some desc",
		Labels:             map[string]string{"a": "b"},
		ServiceAccountName: saName,
		StorageDestination: trailStorageDestination{
			BucketName: bucketName,
		},
		FilteringPolicy: trailFilteringPolicy{
			ManagementFilter: trailManagementFilter{
				ResourceScope: []trailResourceEntry{
					{
						ResourceId:   getExampleFolderID(),
						ResourceType: "resource-manager.folder",
					},
				},
			},
			DataEventFilters: []trailDataEventFilter{
				{
					Service: "storage",
					ResourceScope: []trailResourceEntry{
						{
							ResourceId:   bucketName,
							ResourceType: "storage.bucket",
						},
					},
				},
			},
		},
	}
}

func auditTrailsTrailImportStep(trailResourceName string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      "yandex_audit_trails_trail." + trailResourceName,
		ImportState:       true,
		ImportStateVerify: true,
	}
}

func testAccCheckYandexAuditTrailsTrailAllDestroyed(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_audit_trails_trail" {
			continue
		}
		if err := testAccCheckYandexAuditTrailsTrailDestroyed(rs.Primary.Attributes["trail_id"]); err != nil {
			return err
		}
	}
	return nil
}

func testAccCheckYandexAuditTrailsTrailDestroyed(id string) error {
	config := testAccProvider.Meta().(*Config)
	_, err := config.sdk.AuditTrails().Trail().Get(context.Background(), &audittrails.GetTrailRequest{
		TrailId: id,
	})
	if err == nil {
		return fmt.Errorf("AuditTrailTrail %s still exists", id)
	}
	return nil
}

func auditTrailsServiceAccountConfig(saName string) string {
	return templateConfig(saResourceTemplate, map[string]interface{}{"SaName": saName, "FolderId": getExampleFolderID()})
}

// for_each not allowed https://github.com/hashicorp/terraform-plugin-sdk/issues/536
const saResourceTemplate = `
resource "yandex_iam_service_account" "{{.SaName}}" {
  name = "{{.SaName}}"
  description = "trail acceptance tests SA"
}

resource "yandex_resourcemanager_folder_iam_member" "role-1-{{.SaName}}" {
  folder_id = "{{.FolderId}}"
  role      = "audit-trails.viewer"
  member    = "serviceAccount:${yandex_iam_service_account.{{.SaName}}.id}"
}

resource "yandex_resourcemanager_folder_iam_member" "role-2-{{.SaName}}" {
  folder_id = "{{.FolderId}}"
  role      =  "storage.uploader"
  member    = "serviceAccount:${yandex_iam_service_account.{{.SaName}}.id}"
}

resource "yandex_resourcemanager_folder_iam_member" "role-3-{{.SaName}}" {
  folder_id = "{{.FolderId}}"
  role      = "logging.writer"
  member    = "serviceAccount:${yandex_iam_service_account.{{.SaName}}.id}"
}

resource "yandex_resourcemanager_folder_iam_member" "role-4-{{.SaName}}" {
  folder_id = "{{.FolderId}}"
  role      = "yds.writer"
  member    = "serviceAccount:${yandex_iam_service_account.{{.SaName}}.id}"
}

resource "yandex_resourcemanager_folder_iam_member" "role-5-{{.SaName}}" {
  folder_id = "{{.FolderId}}"
  role      = "ydb.viewer"
  member    = "serviceAccount:${yandex_iam_service_account.{{.SaName}}.id}"
}

resource "yandex_resourcemanager_folder_iam_member" "role-6-{{.SaName}}" {
  folder_id = "{{.FolderId}}"
  role      = "logging.viewer"
  member    = "serviceAccount:${yandex_iam_service_account.{{.SaName}}.id}"
}
`

func auditTrailsStorageResourceConfig(bucketName string) string {
	return templateConfig(storageResourcesTemplate, map[string]interface{}{"BucketName": bucketName, "FolderId": getExampleFolderID()})
}

const storageResourcesTemplate = `
resource "yandex_storage_bucket" "{{.BucketName}}" {
  bucket = "{{.BucketName}}"
  folder_id = "{{.FolderId}}"
}
`

func auditTrailsLoggingResourceConfig(logGroupName string) string {
	return templateConfig(loggingResourcesTemplate, map[string]interface{}{"LogGroupName": logGroupName})
}

const loggingResourcesTemplate = `
resource "yandex_logging_group" "{{.LogGroupName}}" {
  name      = "{{.LogGroupName}}"
}
`

func auditTrailsYdbResourceConfig(ydbName string) string {
	return templateConfig(ydbResourceTemplate, map[string]interface{}{"YdbName": ydbName})
}

func auditTrailsYdsResourceConfig(ydbName, topicName string) string {
	return templateConfig(ydsResourcesTemplate, map[string]interface{}{"YdbName": ydbName, "TopicName": topicName})
}

const ydbResourceTemplate = `
resource "yandex_ydb_database_serverless" "{{.YdbName}}" {
  name = "{{.YdbName}}"
  location_id = "ru-central1"
}
`

const ydsResourcesTemplate = `
resource "yandex_ydb_topic" "{{.TopicName}}" {
  database_endpoint = yandex_ydb_database_serverless.{{.YdbName}}.ydb_full_endpoint
  name = "{{.TopicName}}"

  supported_codecs = ["raw", "gzip"]
  partitions_count = 1
}
`

const trailResourceTemplate = `
resource "yandex_audit_trails_trail" "{{.Name}}" {
 name = "{{.Name}}"
 folder_id = "{{.FolderID}}"
 description = "{{.Description}}"
  
 labels = {
    {{range $key, $value := .Labels}}
    {{$key}} = "{{$value}}"
    {{end}}
 }
  
 service_account_id = yandex_iam_service_account.{{.ServiceAccountName}}.id
  
 {{if .LoggingDestination.LogGroupName}}
 {{with .LoggingDestination}}
 logging_destination {
    log_group_id = yandex_logging_group.{{.LogGroupName}}.id
 }
 {{end}}
 {{end}}
 
 {{if .YDSDestination.StreamName}}
 {{with .YDSDestination}}
 data_stream_destination {
    database_id = yandex_ydb_database_serverless.{{.YdbName}}.id
    stream_name = "{{.StreamName}}"
 }
 {{end}}
 {{end}}

 {{if .StorageDestination.BucketName}}
 {{with .StorageDestination}}
 storage_destination {
    bucket_name = "{{.BucketName}}"
    {{if .ObjectPrefix}}
    {{with .ObjectPrefix}}
    object_prefix = "{{.}}"
    {{end}}
    {{end}}
 }
 {{end}}
 {{end}}

 filtering_policy {
    {{if .FilteringPolicy.ManagementFilter}}
	management_events_filter {
	  {{range .FilteringPolicy.ManagementFilter.ResourceScope}}
	  resource_scope {
	    resource_id = "{{.ResourceId}}"
		resource_type = "{{.ResourceType}}"
	  }
	  {{end}}
	}
	{{end}}
    {{range .FilteringPolicy.DataEventFilters}}
	data_events_filter {
	  service = "{{.Service}}"

	  {{range .ResourceScope}}
	  resource_scope {
	    resource_id = "{{.ResourceId}}"
		resource_type = "{{.ResourceType}}"
	  }
	  {{end}}

	  {{if .IncludedEvents}}
	  included_events = [
	  {{range .IncludedEvents}}
	    "{{.}}",
	  {{end}}
	  ]
	  {{end}}

	  {{if .ExcludedEvents}}
	  excluded_events = [
	  {{range .ExcludedEvents}}
	    "{{.}}",
	  {{end}}
	  ]
	  {{end}}
	}
	{{end}}
 }
}
`

type trailLoggingDestination struct {
	LogGroupName string
}

type trailDataStreamDestination struct {
	YdbName    string
	StreamName string
}

type trailStorageDestination struct {
	BucketName   string
	ObjectPrefix string
}

type trailManagementFilter struct {
	ResourceScope []trailResourceEntry
}

type trailDataEventFilter struct {
	Service        string
	ResourceScope  []trailResourceEntry
	IncludedEvents []string
	ExcludedEvents []string
}

type trailResourceEntry struct {
	ResourceId   string
	ResourceType string
}

type trailFilteringPolicy struct {
	ManagementFilter trailManagementFilter
	DataEventFilters []trailDataEventFilter
}

type yandexAuditTrailsTrail struct {
	Name               string
	FolderID           string
	Description        string
	Labels             map[string]string
	ServiceAccountName string
	LoggingDestination trailLoggingDestination
	YDSDestination     trailDataStreamDestination
	StorageDestination trailStorageDestination
	FilteringPolicy    trailFilteringPolicy
}

func (t yandexAuditTrailsTrail) toTerraformResource() string {
	tmpl := template.Must(template.New("auditTrailsTrail").Parse(trailResourceTemplate))
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, t); err != nil {
		panic(err)
	}
	return buf.String()
}
