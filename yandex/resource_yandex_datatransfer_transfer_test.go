package yandex

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"text/template"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/datatransfer/v1"
	"google.golang.org/grpc/codes"
)

const (
	yandexDataTransferTransferDeleteTimeout = 5 * time.Minute

	sourceEndpointResourceName = "yandex_datatransfer_endpoint.pg_source"
	targetEndpointResourceName = "yandex_datatransfer_endpoint.pg_target"
	transferResourceName       = "yandex_datatransfer_transfer.pgpg_transfer"
)

var (
	randomPostfix = strconv.Itoa(rand.New(rand.NewSource(time.Now().Unix())).Int())
)

func init() {
	resource.AddTestSweepers("yandex_datatransfer", &resource.Sweeper{
		Name: "yandex_datatransfer",
		F:    testSweepDataTransfer,
	})
}

func testSweepDataTransfer(_ string) error {
	config, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting config: %s", err)
	}
	listTransfersResponse, err := config.sdk.DataTransfer().Transfer().List(config.Context(), &datatransfer.ListTransfersRequest{FolderId: config.FolderID})
	if err != nil {
		return fmt.Errorf("error getting DataTransfer transfers: %s", err)
	}

	resultError := &multierror.Error{}
	for _, transfer := range listTransfersResponse.Transfers {
		if !sweepDataTransferTransfer(config, transfer.Id) {
			resultError = multierror.Append(resultError, fmt.Errorf("failed to sweep DataTransfer transfer %q", transfer.Id))
		}
	}
	if resultError.Len() > 0 {
		return resultError
	}

	listEndpointsResponse, err := config.sdk.DataTransfer().Endpoint().List(config.Context(), &datatransfer.ListEndpointsRequest{FolderId: config.FolderID})
	if err != nil {
		return fmt.Errorf("error getting DataTransfer transfers: %s", err)
	}
	for _, endpoint := range listEndpointsResponse.Endpoints {
		if !sweepDataTransferEndpoint(config, endpoint.Id) {
			resultError = multierror.Append(resultError, fmt.Errorf("failed to sweep "))
		}
	}
	return resultError.ErrorOrNil()
}

func sweepDataTransferTransfer(config *Config, transferID string) bool {
	return sweepWithRetry(sweepDataTransferTransferOnce, config, "DataTransfer transfer", transferID)
}

func sweepDataTransferTransferOnce(config *Config, transferID string) error {
	ctx, cancel := config.ContextWithTimeout(yandexDataTransferTransferDeleteTimeout)
	defer cancel()

	deleteOperation, err := config.sdk.DataTransfer().Transfer().Delete(ctx, &datatransfer.DeleteTransferRequest{TransferId: transferID})
	return handleSweepOperation(ctx, config, deleteOperation, err)
}

func sweepDataTransferEndpoint(config *Config, transferID string) bool {
	return sweepWithRetry(sweepDataTransferEndpointOnce, config, "DataTransfer endpoint", transferID)
}

func sweepDataTransferEndpointOnce(config *Config, endpointID string) error {
	ctx, cancel := config.ContextWithTimeout(yandexDataTransferTransferDeleteTimeout)
	defer cancel()

	deleteOperation, err := config.sdk.DataTransfer().Endpoint().Delete(ctx, &datatransfer.DeleteEndpointRequest{EndpointId: endpointID})
	return handleSweepOperation(ctx, config, deleteOperation, err)
}

func dataTransferSourceEndpointImportStep() resource.TestStep {
	return resource.TestStep{
		ResourceName:            sourceEndpointResourceName,
		ImportState:             true,
		ImportStateVerify:       true,
		ImportStateVerifyIgnore: []string{"settings.0.postgres_source.0.password"},
	}
}

func dataTransferTargetEndpointImportStep() resource.TestStep {
	return resource.TestStep{
		ResourceName:            targetEndpointResourceName,
		ImportState:             true,
		ImportStateVerify:       true,
		ImportStateVerifyIgnore: []string{"settings.0.postgres_target.0.password"},
	}
}

func dataTransferTransferImportStep() resource.TestStep {
	return resource.TestStep{
		ResourceName:      transferResourceName,
		ImportState:       true,
		ImportStateVerify: true,
	}
}

// Test that a DataTransfer Transfer can be created, updated and destroyed along with the endpoints
func TestAccDataTransferTransfer_full(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDataTransferDestroy,
		Steps: []resource.TestStep{
			//Create DataTransfer transfer and two endpoints
			{
				Config: testAccDataTransferConfigMain(defaultTemplateParams),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(sourceEndpointResourceName, "name", defaultTemplateParams.SourceEndpointName),
					resource.TestCheckResourceAttr(sourceEndpointResourceName, "description", defaultTemplateParams.SourceEndpointDescription),
					resource.TestCheckResourceAttr(sourceEndpointResourceName, "settings.0.postgres_source.0.password.0.raw", defaultTemplateParams.SourceEndpointPassword),
					resource.TestCheckResourceAttr(sourceEndpointResourceName, "settings.0.postgres_source.0.connection.0.on_premise.0.hosts.0", defaultTemplateParams.SourceEndpointHostName),
					resource.TestCheckResourceAttr(sourceEndpointResourceName, "settings.0.postgres_source.0.connection.0.on_premise.0.port", strconv.Itoa(defaultTemplateParams.SourceEndpointPort)),
					resource.TestCheckResourceAttr(sourceEndpointResourceName, "settings.0.postgres_source.0.slot_gigabyte_lag_limit", strconv.Itoa(defaultTemplateParams.SourceEndpointSlotGigabyteLagLimit)),

					resource.TestCheckResourceAttr(targetEndpointResourceName, "name", defaultTemplateParams.TargetEndpointName),
					resource.TestCheckResourceAttr(targetEndpointResourceName, "description", defaultTemplateParams.TargetEndpointDescription),
					resource.TestCheckResourceAttr(targetEndpointResourceName, "settings.0.postgres_target.0.password.0.raw", defaultTemplateParams.TargetEndpointPassword),
					resource.TestCheckResourceAttr(targetEndpointResourceName, "settings.0.postgres_target.0.connection.0.on_premise.0.hosts.0", defaultTemplateParams.TargetEndpointHostName),
					resource.TestCheckResourceAttr(targetEndpointResourceName, "settings.0.postgres_target.0.connection.0.on_premise.0.port", strconv.Itoa(defaultTemplateParams.TargetEndpointPort)),
					resource.TestCheckResourceAttr(targetEndpointResourceName, "settings.0.postgres_target.0.cleanup_policy", defaultTemplateParams.CleanupPolicy),

					resource.TestCheckResourceAttr(transferResourceName, "name", defaultTemplateParams.TransferName),
					resource.TestCheckResourceAttr(transferResourceName, "description", defaultTemplateParams.TransferDescription),
					resource.TestCheckResourceAttrSet(transferResourceName, "source_id"),
					resource.TestCheckResourceAttrSet(transferResourceName, "target_id"),
				),
			},
			// Update transfer name, expect that description stays the same
			{
				Config: testAccDataTransferConfigMain(defaultTemplateParams.withTransferName("new-transfer-name" + randomPostfix)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(transferResourceName, "name", "new-transfer-name"+randomPostfix),
					resource.TestCheckResourceAttr(transferResourceName, "description", defaultTemplateParams.TransferDescription),
				),
			},
			// Update endpoints, set back old transfer name
			{
				Config: testAccDataTransferConfigMain(
					defaultTemplateParams.
						withSourceEndpointName("new-source-name" + randomPostfix).
						withSourceEndpointSlotGigabyteLagLimit(200).
						withTargetEndpointName("new-target-name" + randomPostfix).
						withTargetEndpointPassword("12345"),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(sourceEndpointResourceName, "name", "new-source-name"+randomPostfix),
					resource.TestCheckResourceAttr(sourceEndpointResourceName, "description", defaultTemplateParams.SourceEndpointDescription),
					resource.TestCheckResourceAttr(sourceEndpointResourceName, "settings.0.postgres_source.0.slot_gigabyte_lag_limit", "200"),
					resource.TestCheckResourceAttr(sourceEndpointResourceName, "settings.0.postgres_source.0.password.0.raw", defaultTemplateParams.SourceEndpointPassword),

					resource.TestCheckResourceAttr(targetEndpointResourceName, "name", "new-target-name"+randomPostfix),
					resource.TestCheckResourceAttr(targetEndpointResourceName, "description", defaultTemplateParams.TargetEndpointDescription),
					resource.TestCheckResourceAttr(targetEndpointResourceName, "settings.0.postgres_target.0.password.0.raw", "12345"),

					resource.TestCheckResourceAttr(transferResourceName, "name", defaultTemplateParams.TransferName),
					resource.TestCheckResourceAttr(transferResourceName, "description", defaultTemplateParams.TransferDescription),
				),
			},
			{
				Config: testAccDataTransferConfigMain(
					defaultTemplateParams.
						withTransferType("SNAPSHOT_AND_INCREMENT").
						withActivateMode(dontActivateMode),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(transferResourceName, "on_create_activate_mode", internalMessageActivateMode),
				),
			},
			{
				Config: testAccDataTransferConfigMain(
					defaultTemplateParams.
						withTransferType("SNAPSHOT_AND_INCREMENT").
						withActivateMode(asyncActivateMode),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(transferResourceName, "on_create_activate_mode", internalMessageActivateMode),
				),
			},
			dataTransferSourceEndpointImportStep(),
			dataTransferTargetEndpointImportStep(),
			dataTransferTransferImportStep(),
		},
	})
}

func testAccCheckDataTransferDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type == "yandex_datatransfer_transfer" {
			_, err := config.sdk.DataTransfer().Transfer().Get(context.Background(), &datatransfer.GetTransferRequest{TransferId: rs.Primary.ID})
			if err == nil {
				return fmt.Errorf("DataTransfer Transfer %s still exists", rs.Primary.ID)
			}
			if !isStatusWithCode(err, codes.NotFound) {
				return fmt.Errorf("Cannot get Transfer %s: %s", rs.Primary.ID, err)
			}
		}
	}

	return nil
}

type dataTransferTerraformTemplateParams struct {
	SourceEndpointName                 string
	SourceEndpointDescription          string
	SourceEndpointPassword             string
	SourceEndpointHostName             string
	SourceEndpointPort                 int
	SourceEndpointSlotGigabyteLagLimit int
	TargetEndpointName                 string
	TargetEndpointDescription          string
	TargetEndpointPassword             string
	TargetEndpointHostName             string
	TargetEndpointPort                 int
	TransferName                       string
	TransferDescription                string
	TransferType                       string
	TransferActivateMode               string
	CleanupPolicy                      string
}

var defaultTemplateParams = dataTransferTerraformTemplateParams{
	SourceEndpointName:                 "datatransfer-src-endpoint" + randomPostfix,
	SourceEndpointDescription:          "src description",
	SourceEndpointPassword:             "src password",
	SourceEndpointHostName:             "src hostname",
	SourceEndpointPort:                 5432,
	SourceEndpointSlotGigabyteLagLimit: 10,
	TargetEndpointName:                 "datatransfer-dst-endpoint" + randomPostfix,
	TargetEndpointDescription:          "dst description",
	TargetEndpointPassword:             "dst password",
	TargetEndpointHostName:             "dst hostname",
	TargetEndpointPort:                 5432,
	TransferName:                       "datatransfer-transfer" + randomPostfix,
	TransferDescription:                "transfer description",
	TransferType:                       "SNAPSHOT_ONLY",
	TransferActivateMode:               syncActivateMode,
	CleanupPolicy:                      "DROP",
}

func (p dataTransferTerraformTemplateParams) withSourceEndpointName(sourceEndpointName string) dataTransferTerraformTemplateParams {
	p.SourceEndpointName = sourceEndpointName
	return p
}

func (p dataTransferTerraformTemplateParams) withSourceEndpointSlotGigabyteLagLimit(sourceEndpointSlotGigabyteLagLimit int) dataTransferTerraformTemplateParams {
	p.SourceEndpointSlotGigabyteLagLimit = sourceEndpointSlotGigabyteLagLimit
	return p
}

func (p dataTransferTerraformTemplateParams) withTargetEndpointName(targetEndpointName string) dataTransferTerraformTemplateParams {
	p.TargetEndpointName = targetEndpointName
	return p
}

func (p dataTransferTerraformTemplateParams) withTargetEndpointPassword(targetEndpointPassword string) dataTransferTerraformTemplateParams {
	p.TargetEndpointPassword = targetEndpointPassword
	return p
}

func (p dataTransferTerraformTemplateParams) withTransferName(transferName string) dataTransferTerraformTemplateParams {
	p.TransferName = transferName
	return p
}

func (p dataTransferTerraformTemplateParams) withActivateMode(activateMode string) dataTransferTerraformTemplateParams {
	p.TransferActivateMode = activateMode
	return p
}

func (p dataTransferTerraformTemplateParams) withTransferType(transferType string) dataTransferTerraformTemplateParams {
	p.TransferType = transferType
	return p
}
func testAccDataTransferConfigMain(templateParams dataTransferTerraformTemplateParams) string {
	template := template.Must(template.New("main.tf").Parse(`
		resource "yandex_datatransfer_endpoint" "pg_source" {
		  name = "{{.SourceEndpointName}}"
		  description = "{{.SourceEndpointDescription}}"
		  settings {
			postgres_source {
			  connection {
				on_premise {
				  hosts = [
					"{{.SourceEndpointHostName}}"
				  ]
				  port = {{.SourceEndpointPort}}
				}
			  }
			  slot_gigabyte_lag_limit = {{.SourceEndpointSlotGigabyteLagLimit}}
			  database = "postgres"
			  user = "postgres"
			  password {
				raw = "{{.SourceEndpointPassword}}"
			  }
			}
		  }
		}

		resource "yandex_datatransfer_endpoint" "pg_target" {
		  name = "{{.TargetEndpointName}}"
		  description = "{{.TargetEndpointDescription}}"
		  settings {
			postgres_target {
			  connection {
				on_premise {
				  hosts = [
					"{{.TargetEndpointHostName}}"
				  ]
				  port = {{.TargetEndpointPort}}
				}
			  }
			  database = "postgres"
			  user = "postgres"
			  password {
				raw = "{{.TargetEndpointPassword}}"
			  }
			  cleanup_policy = "DROP"
			}
		  }
		}

		resource "yandex_datatransfer_transfer" "pgpg_transfer" {
		  name = "{{.TransferName}}"
		  description = "{{.TransferDescription}}"
		  source_id = yandex_datatransfer_endpoint.pg_source.id
		  target_id = yandex_datatransfer_endpoint.pg_target.id
		  type = "{{.TransferType}}"
          on_create_activate_mode = "{{.TransferActivateMode}}"
		}
	`))
	buffer := bytes.NewBuffer(nil)
	_ = template.Execute(buffer, templateParams)
	return buffer.String()
}

func TestAccDataTransferKafkaSourceEndpoint(t *testing.T) {
	t.Parallel()
	const kafkaSourceEndpointResourceName = "kafka-source"
	const fullResourceName = "yandex_datatransfer_endpoint.kafka_source"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDataTransferDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataTransferConfigKafkaSource(kafkaSourceEndpointResourceName+randomPostfix, "TestAccDataTransfer"+randomPostfix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResourceName, "name", kafkaSourceEndpointResourceName+randomPostfix),
					resource.TestCheckResourceAttr(fullResourceName, "description", "TestAccDataTransfer"+randomPostfix),
				),
			},
			{
				Config: testAccDataTransferConfigKafkaSource("new-kafka-source-name"+randomPostfix, "TestAccDataTransfer"+randomPostfix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResourceName, "name", "new-kafka-source-name"+randomPostfix),
					resource.TestCheckResourceAttr(fullResourceName, "description", "TestAccDataTransfer"+randomPostfix),
				),
			},
			{
				ResourceName:      fullResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccDataTransferConfigKafkaSource(name, description string) string {
	return fmt.Sprintf(`
		resource "yandex_datatransfer_endpoint" "kafka_source" {
  name        = "%s"
  description = "%s"
  settings {
    kafka_source {
      security_groups = []
      topic_name      = "topic-name"

      auth {
        no_auth {}
      }
      connection {
        on_premise {
          broker_urls = [
            "localhost:1234",
          ]
          tls_mode {
            disabled {}
          }
        }
      }
      parser {
        json_parser {
          add_rest_column   = false
          null_keys_allowed = false
          data_schema {
            fields {
              fields {
                key      = false
                name     = "123123"
                required = false
                type     = "ANY"
              }
            }
          }
        }
      }
    }
  }
}`, name, description)
}

func TestAccDataTransferKafkaTargetEndpoint(t *testing.T) {
	t.Parallel()
	const kafkaTargetEndpointResourceName = "kafka-target"
	const fullResourceName = "yandex_datatransfer_endpoint.kafka_target"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDataTransferDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataTransferConfigKafkaTarget(kafkaTargetEndpointResourceName+randomPostfix, "TestAccDataTransfer"+randomPostfix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResourceName, "name", kafkaTargetEndpointResourceName+randomPostfix),
					resource.TestCheckResourceAttr(fullResourceName, "description", "TestAccDataTransfer"+randomPostfix),
				),
			},
			{
				Config: testAccDataTransferConfigKafkaTarget("new-kafka-target-name"+randomPostfix, "TestAccDataTransfer"+randomPostfix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResourceName, "name", "new-kafka-target-name"+randomPostfix),
					resource.TestCheckResourceAttr(fullResourceName, "description", "TestAccDataTransfer"+randomPostfix),
				),
			},
			{
				ResourceName:            fullResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"settings.0.kafka_target.0.auth.0.sasl.0.password."},
			},
		},
	})
}

func testAccDataTransferConfigKafkaTarget(name, description string) string {
	return fmt.Sprintf(`resource "yandex_datatransfer_endpoint" "kafka_target" {
    name        = "%s"
  	description = "%s"
    settings {
        kafka_target {
            security_groups = []
            auth {
                sasl {
                    mechanism = "KAFKA_MECHANISM_SHA256"
                    user      = "user"
					password  {
						raw = "password"
					  }
                }
            }
            connection {
                on_premise {
                    broker_urls = [
                        "localhost:9999",
                    ]
                    tls_mode {
                        enabled {
                            ca_certificate = "123123123123"
                        }
                    }
                }
            }
            topic_settings {
                topic_prefix = "topic-prefix"
            }
			serializer {
				serializer_json{
                }
			}
        }
    }
}`, name, description)
}

func TestAccDataTransferYDBSourceEndpoint(t *testing.T) {
	t.Parallel()
	const ydbSourceEndpointResourceName = "ydb-source"
	const fullResourceName = "yandex_datatransfer_endpoint.ydb_source"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDataTransferDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataTransferConfigYdbSource(ydbSourceEndpointResourceName+randomPostfix, "TestAccDataTransfer"+randomPostfix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResourceName, "name", ydbSourceEndpointResourceName+randomPostfix),
					resource.TestCheckResourceAttr(fullResourceName, "description", "TestAccDataTransfer"+randomPostfix),
				),
			},
			{
				Config: testAccDataTransferConfigYdbSource("new-ydb-source-name"+randomPostfix, "TestAccDataTransfer"+randomPostfix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResourceName, "name", "new-ydb-source-name"+randomPostfix),
					resource.TestCheckResourceAttr(fullResourceName, "description", "TestAccDataTransfer"+randomPostfix),
				),
			},
			{
				ResourceName:            fullResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"settings.0.ydb_source.0.sa_key_content."},
			},
		},
	})
}

func testAccDataTransferConfigYdbSource(name, description string) string {
	return fmt.Sprintf(`
resource "yandex_iam_service_account" "ydb_sa_%[1]s" {
  name        = "ydb-sa-%[1]s"
  description = "service account for %[1]s ydb endpoint"
}

resource "yandex_datatransfer_endpoint" "ydb_source" {
  name        = "%[1]s"
  description = "%[2]s"
  settings {
    ydb_source {
      database = "xyz"
      instance = "my-cute-ydb.yandex.cloud:2135"
      paths = [
        "path1/a/b/c",
        "path2/a/b/c",
        "path3/a/b/c",
      ]
      security_groups = []
	  service_account_id = yandex_iam_service_account.ydb_sa_%[1]s.id
    }
  }
}
`, name, description)
}

func TestAccDataTransferYdbTargetEndpoint(t *testing.T) {
	t.Parallel()
	const ydbTargetEndpointResourceName = "ydb-target"
	const fullResourceName = "yandex_datatransfer_endpoint.ydb_target"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDataTransferDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataTransferConfigYdbTarget(ydbTargetEndpointResourceName+randomPostfix, "TestAccDataTransfer"+randomPostfix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResourceName, "name", ydbTargetEndpointResourceName+randomPostfix),
					resource.TestCheckResourceAttr(fullResourceName, "description", "TestAccDataTransfer"+randomPostfix),
					resource.TestCheckResourceAttr(fullResourceName, "settings.0.ydb_target.0.is_table_column_oriented", "true"),
				),
			},
			{
				Config: testAccDataTransferConfigYdbTarget("new-ydb-target-name"+randomPostfix, "TestAccDataTransfer"+randomPostfix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResourceName, "name", "new-ydb-target-name"+randomPostfix),
					resource.TestCheckResourceAttr(fullResourceName, "description", "TestAccDataTransfer"+randomPostfix),
					resource.TestCheckResourceAttr(fullResourceName, "settings.0.ydb_target.0.is_table_column_oriented", "true"),
				),
			},
			{
				ResourceName:            fullResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"settings.0.ydb_target.0.sa_key_content."},
			},
		},
	})
}

func testAccDataTransferConfigYdbTarget(name, description string) string {
	return fmt.Sprintf(`
resource "yandex_iam_service_account" "ydb_sa_%[1]s" {
  name        = "ydb-sa-%[1]s"
  description = "service account for %[1]s ydb endpoint"
}

resource "yandex_datatransfer_endpoint" "ydb_target" {
    name        = "%[1]s"
  	description = "%[2]s"
    settings {
        ydb_target {
          database = "xyz"
          instance = "my-cute-ydb.yandex.cloud"
          path = "/bushido/logs"
          security_groups = []
          cleanup_policy = "YDB_CLEANUP_POLICY_DROP"
		  is_table_column_oriented = true
          service_account_id = yandex_iam_service_account.ydb_sa_%[1]s.id
        }
    }
}`, name, description)
}

func TestAccDataTransferTransferWithTransformation(t *testing.T) {
	t.Parallel()
	const ydsTargetEndpointResourceName = "yds-target-with-transformation"
	const mysqlSourceEndpointResourceName = "mysql-source-with-transformation"
	const transferResourceName = "transfer-with-transformation"
	const fullTransferResourceName = "yandex_datatransfer_transfer.transfer_with_transformation"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDataTransferDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataTransferConfigTransferWithTransformation(
					ydsTargetEndpointResourceName+randomPostfix,
					mysqlSourceEndpointResourceName+randomPostfix,
					transferResourceName+randomPostfix,
					"new_name",
					"TestAccDataTransfer"+randomPostfix,
					3,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullTransferResourceName, "name", transferResourceName+randomPostfix),
					resource.TestCheckResourceAttr(fullTransferResourceName, "runtime.0.yc_runtime.0.job_count", "3"),
					resource.TestCheckResourceAttr(fullTransferResourceName, "runtime.0.yc_runtime.0.upload_shard_params.0.job_count", "3"),
					resource.TestCheckResourceAttr(fullTransferResourceName, "runtime.0.yc_runtime.0.upload_shard_params.0.process_count", "3"),
					resource.TestCheckResourceAttr(fullTransferResourceName, "transformation.0.transformers.0.filter_columns.0.columns.0.exclude_columns.0", "col1"),
					resource.TestCheckResourceAttr(fullTransferResourceName, "transformation.0.transformers.1.rename_tables.0.rename_tables.0.new_name.0.name", "new_name1"),
					resource.TestCheckResourceAttr(fullTransferResourceName, "transformation.0.transformers.1.rename_tables.0.rename_tables.0.original_name.0.name", "original_name1"),
					resource.TestCheckResourceAttr(fullTransferResourceName, "transformation.0.transformers.1.rename_tables.0.rename_tables.1.new_name.0.name", "new_name2"),
					resource.TestCheckResourceAttr(fullTransferResourceName, "transformation.0.transformers.1.rename_tables.0.rename_tables.1.original_name.0.name", "original_name2"),
				),
			},
			{
				Config: testAccDataTransferConfigTransferWithTransformation(
					ydsTargetEndpointResourceName+randomPostfix,
					mysqlSourceEndpointResourceName+randomPostfix,
					transferResourceName+randomPostfix+"2",
					"new_name2",
					"TestAccDataTransfer"+randomPostfix,
					2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullTransferResourceName, "name", transferResourceName+randomPostfix+"2"),
					resource.TestCheckResourceAttr(fullTransferResourceName, "runtime.0.yc_runtime.0.job_count", "2"),
					resource.TestCheckResourceAttr(fullTransferResourceName, "runtime.0.yc_runtime.0.upload_shard_params.0.job_count", "2"),
					resource.TestCheckResourceAttr(fullTransferResourceName, "runtime.0.yc_runtime.0.upload_shard_params.0.process_count", "3"),
					resource.TestCheckResourceAttr(fullTransferResourceName, "transformation.0.transformers.0.filter_columns.0.columns.0.exclude_columns.0", "col1"),
					resource.TestCheckResourceAttr(fullTransferResourceName, "transformation.0.transformers.1.rename_tables.0.rename_tables.0.new_name.0.name", "new_name21"),
					resource.TestCheckResourceAttr(fullTransferResourceName, "transformation.0.transformers.1.rename_tables.0.rename_tables.0.original_name.0.name", "original_name1"),
					resource.TestCheckResourceAttr(fullTransferResourceName, "transformation.0.transformers.1.rename_tables.0.rename_tables.1.new_name.0.name", "new_name22"),
					resource.TestCheckResourceAttr(fullTransferResourceName, "transformation.0.transformers.1.rename_tables.0.rename_tables.1.original_name.0.name", "original_name2"),
				),
			},
			{
				ResourceName:      fullTransferResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccDataTransferConfigTransferWithTransformation(
	ydsName,
	mysqlName,
	transferName,
	newName,
	description string,
	jobCount int,
) string {
	return fmt.Sprintf(`
resource "yandex_datatransfer_endpoint" "mysql_source_with_transformation" {
  name      = "%[1]s"
  settings {
    mysql_source {
      connection {
		on_premise {
		  hosts = [
			"src hostname"
		  ]
		  port = 3306
		}
	  }
      database = "db1"
      user     = "user1"
      password {
        raw = "pass"
      }
    }
  }
}

resource "yandex_iam_service_account" "yds_sa_with_transformation" {
  name        = "yds-sa-%[2]s"
  description = "service account for %[2]s yds endpoint"
}

resource "yandex_datatransfer_endpoint" "yds_target_with_transformation" {
  name      = "%[2]s"
  settings {
    yds_target {
      endpoint           = "endpoint"
      stream             = "test"
      database           = "database"
      service_account_id = yandex_iam_service_account.yds_sa_with_transformation.id
      serializer {
        serializer_auto {

        }
      }
    }
  }
}

resource "yandex_datatransfer_transfer" "transfer_with_transformation" {
  name        = "%[3]s"
  description = "%[4]s"
  source_id   = yandex_datatransfer_endpoint.mysql_source_with_transformation.id
  target_id   = yandex_datatransfer_endpoint.yds_target_with_transformation.id
  type        = "SNAPSHOT_AND_INCREMENT"
  runtime {
    yc_runtime {
      job_count = %[5]d
      upload_shard_params {
        job_count = %[5]d
        process_count = 3
      }
    }
  }
  transformation {
    transformers {
      filter_columns {
        columns {
          exclude_columns = ["col1"]
        }
		tables {}
      }
    }
    transformers {
      rename_tables {
        rename_tables {
          new_name {
            name = "%[6]s1"
          }
          original_name {
            name = "original_name1"
          }
        }
        rename_tables {
          new_name {
            name = "%[6]s2"
          }
          original_name {
            name = "original_name2"
          }
        }
      }
    }
  }
}
`, ydsName, mysqlName, transferName, description, jobCount, newName)
}

func TestAccDataTransferYDSSourceEndpoint(t *testing.T) {
	t.Parallel()
	const ydsSourceEndpointResourceName = "yds-source"
	const fullResourceName = "yandex_datatransfer_endpoint.yds_source"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDataTransferDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataTransferConfigYdsSource(ydsSourceEndpointResourceName+randomPostfix, "TestAccDataTransfer"+randomPostfix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResourceName, "name", ydsSourceEndpointResourceName+randomPostfix),
					resource.TestCheckResourceAttr(fullResourceName, "description", "TestAccDataTransfer"+randomPostfix),
					resource.TestCheckResourceAttr(fullResourceName, "settings.0.yds_source.0.supported_codecs.0", "YDS_COMPRESSION_CODEC_GZIP"),
					resource.TestCheckResourceAttr(fullResourceName, "settings.0.yds_source.0.supported_codecs.1", "YDS_COMPRESSION_CODEC_RAW"),
				),
			},
			{
				Config: testAccDataTransferConfigYdsSource("new-yds-source-name"+randomPostfix, "TestAccDataTransfer"+randomPostfix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResourceName, "name", "new-yds-source-name"+randomPostfix),
					resource.TestCheckResourceAttr(fullResourceName, "description", "TestAccDataTransfer"+randomPostfix),
				),
			},
			{
				ResourceName:      fullResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccDataTransferConfigYdsSource(name, description string) string {
	return fmt.Sprintf(`
resource "yandex_iam_service_account" "yds_sa_%[1]s" {
  name        = "yds-sa-%[1]s"
  description = "service account for %[1]s yds endpoint"
}

resource "yandex_datatransfer_endpoint" "yds_source" {
  name        = "%[1]s"
  description = "%[2]s"
  settings {
    yds_source {
      endpoint           = "endpoint"
      stream             = "stream"
      database           = "database"
      consumer           = "consumer"
      parser {
        cloud_logging_parser {}
      }
      supported_codecs = ["YDS_COMPRESSION_CODEC_GZIP", "YDS_COMPRESSION_CODEC_RAW"]
      service_account_id = yandex_iam_service_account.yds_sa_%[1]s.id
    }
  }
}
`, name, description)
}

func TestAccDataTransferYdsTargetEndpoint(t *testing.T) {
	t.Parallel()
	const ydsTargetEndpointResourceName = "yds-target"
	const fullResourceName = "yandex_datatransfer_endpoint.yds_target"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDataTransferDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataTransferConfigYdsTarget(ydsTargetEndpointResourceName+randomPostfix, "TestAccDataTransfer"+randomPostfix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResourceName, "name", ydsTargetEndpointResourceName+randomPostfix),
					resource.TestCheckResourceAttr(fullResourceName, "description", "TestAccDataTransfer"+randomPostfix),
				),
			},
			{
				Config: testAccDataTransferConfigYdsTarget("new-yds-target-name"+randomPostfix, "TestAccDataTransfer"+randomPostfix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResourceName, "name", "new-yds-target-name"+randomPostfix),
					resource.TestCheckResourceAttr(fullResourceName, "description", "TestAccDataTransfer"+randomPostfix),
				),
			},
			{
				ResourceName:      fullResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccDataTransferConfigYdsTarget(name, description string) string {
	return fmt.Sprintf(`
resource "yandex_iam_service_account" "yds_sa_%[1]s" {
  name        = "yds-sa-%[1]s"
  description = "service account for %[1]s yds endpoint"
}

resource "yandex_datatransfer_endpoint" "yds_target" {
  name        = "%[1]s"
  description = "%[2]s"
  settings {
    yds_target {
      endpoint           = "endpoint"
      stream             = "stream"
      database           = "database"
      service_account_id = yandex_iam_service_account.yds_sa_%[1]s.id
      serializer {
        serializer_auto {

        }
      }
    }
  }
}
`, name, description)
}

func TestAccDataTransferMetrikaSourceEndpoint(t *testing.T) {
	t.Parallel()
	const metrikaSourceEndpointResourceName = "metrika-source"
	const fullResourceName = "yandex_datatransfer_endpoint.metrika_source"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDataTransferDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataTransferConfigMetrikaSource(metrikaSourceEndpointResourceName+randomPostfix, "TestAccDataTransfer"+randomPostfix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResourceName, "name", metrikaSourceEndpointResourceName+randomPostfix),
					resource.TestCheckResourceAttr(fullResourceName, "description", "TestAccDataTransfer"+randomPostfix),
				),
			},
			{
				Config: testAccDataTransferConfigMetrikaSource("new-metrika-source-name"+randomPostfix, "TestAccDataTransfer"+randomPostfix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResourceName, "name", "new-metrika-source-name"+randomPostfix),
					resource.TestCheckResourceAttr(fullResourceName, "description", "TestAccDataTransfer"+randomPostfix),
					resource.TestCheckResourceAttr(fullResourceName, "settings.0.metrika_source.0.counter_ids.0", "1"),
					resource.TestCheckResourceAttr(fullResourceName, "settings.0.metrika_source.0.counter_ids.1", "2"),
					resource.TestCheckResourceAttr(fullResourceName, "settings.0.metrika_source.0.counter_ids.2", "3"),
				),
			},
			{
				ResourceName:            fullResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"settings.0.metrika_source.0.token"},
			},
		},
	})
}

func testAccDataTransferConfigMetrikaSource(name, description string) string {
	return fmt.Sprintf(`
resource "yandex_datatransfer_endpoint" "metrika_source" {
  name        = "%[1]s"
  description = "%[2]s"
  settings {
    metrika_source {
      counter_ids = [1, 2, 3]
      streams {
        columns = ["col1", "col2"]
        type = "METRIKA_STREAM_TYPE_HITS"
      }
      streams {
        columns = ["col1", "col2"]
        type = "METRIKA_STREAM_TYPE_HITS_V2"
      }
      streams {
        columns = ["col1", "col2"]
        type = "METRIKA_STREAM_TYPE_VISITS"
      }
      token {
        raw = "my_token"
      }
    }
  }
}
`, name, description)
}
