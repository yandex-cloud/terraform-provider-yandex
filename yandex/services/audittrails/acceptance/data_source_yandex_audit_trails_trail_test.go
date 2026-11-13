package acceptance

import (
	"bytes"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
)

func TestAccDataSourceAuditTrailsTrail_storageByID(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	saName := acctest.RandomWithPrefix("tf-data-acc-trail-storage-sa")
	bucketTestName := acctest.RandomWithPrefix("tf-data-acc-trail-bucket")
	trailTestName := acctest.RandomWithPrefix("tf-data-acc-trail")

	tfBaseConfig := auditTrailsServiceAccountConfig(saName) + auditTrailsStorageResourceConfig(bucketTestName)
	trailConfig := auditTrailsStorageConfig(trailTestName, bucketTestName, saName, "1m")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckYandexAuditTrailsTrailAllDestroyed, // delete is called for each resource and checked
		Steps: []resource.TestStep{
			// create base infrastructure
			{
				Config: tfBaseConfig,
			},
			// create base storage destination trail with full filter
			// and import it in the data block
			{
				Config: tfBaseConfig + trailConfig.toTerraformResource() + trailConfig.toTerraformData(),
				Check:  checkTrail(trailConfig, true),
			},
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

func TestAccDataSourceAuditTrailsTrail_loggingByID(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	saName := acctest.RandomWithPrefix("tf-data-acc-trail-logging-sa")
	loggingTestName := acctest.RandomWithPrefix("tf-data-acc-trail-group")
	trailTestName := acctest.RandomWithPrefix("tf-data-acc-trail")

	tfBaseConfig := auditTrailsServiceAccountConfig(saName) + auditTrailsLoggingResourceConfig(loggingTestName)
	trailConfig := auditTrailsLoggingConfig(trailTestName, loggingTestName, saName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckYandexAuditTrailsTrailAllDestroyed, // delete is called for each resource and checked
		Steps: []resource.TestStep{
			// create base logging infrastructure
			{
				Config: tfBaseConfig,
			},
			// create base logging destination trail with minimal filter
			// and import it in the data block
			{
				Config: tfBaseConfig + trailConfig.toTerraformResource() + trailConfig.toTerraformData(),
				Check:  checkTrail(trailConfig, true),
			},
		},
	})
}

func TestAccDataSourceAuditTrailsTrail_moniumByID(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	saName := acctest.RandomWithPrefix("tf-data-acc-trail-monium-sa")
	trailTestName := acctest.RandomWithPrefix("tf-data-acc-trail")
	tfBaseConfig := auditTrailsServiceAccountConfig(saName)
	trailConfig := auditTrailsMoniumConfig(trailTestName, saName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckYandexAuditTrailsTrailAllDestroyed,
		Steps: []resource.TestStep{
			{
				Config: tfBaseConfig,
			},
			{
				Config: tfBaseConfig + trailConfig.toTerraformResource() + trailConfig.toTerraformData(),
				Check:  checkTrail(trailConfig, true),
			},
		},
	})
}

func TestAccDataSourceAuditTrailsTrail_dataStreamByID(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	saName := acctest.RandomWithPrefix("tf-data-acc-trail-yds-sa")
	ydbTestName := acctest.RandomWithPrefix("tf-data-acc-trail-ydb")
	streamTestName := acctest.RandomWithPrefix("tf-data-acc-trail-stream")
	trailTestName := acctest.RandomWithPrefix("tf-data-acc-trail")

	// base config describes required resources for this test - we will reuse it to check only trail update logic
	tfYdbConfig := auditTrailsYdbResourceConfig(ydbTestName)
	tfBaseConfig := tfYdbConfig + auditTrailsServiceAccountConfig(saName) + auditTrailsYdsResourceConfig(ydbTestName, streamTestName)
	trailConfig := auditTrailsYdsConfig(trailTestName, ydbTestName, streamTestName, saName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckYandexAuditTrailsTrailAllDestroyed, // delete is called for each resource and checked
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
			// and import it in the data block
			{
				Config: tfBaseConfig + trailConfig.toTerraformResource() + trailConfig.toTerraformData(),
				Check:  checkTrail(trailConfig, true),
			},
		},
	})
}

const trailDataTemplate = `
data "yandex_audit_trails_trail" "{{.Name}}" {
  trail_id = yandex_audit_trails_trail.{{.Name}}.id
}
`

func (t *yandexAuditTrailsTrail) toTerraformData() string {
	tmpl := template.Must(template.New("auditTrailsTrailData").Parse(trailDataTemplate))
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, t); err != nil {
		panic(err)
	}
	return buf.String()
}
