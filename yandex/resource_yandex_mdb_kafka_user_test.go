package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/kafka/v1"
	"google.golang.org/grpc/codes"
)

func TestNoCrashOnNilPermissions(t *testing.T) {
	raw := map[string]interface{}{
		"name":       "events_user",
		"password":   "test_pwd",
		"permission": []interface{}{},
	}
	resourceData := schema.TestResourceDataRaw(t, resourceYandexMDBKafkaUser().Schema, raw)

	userSpec, err := buildKafkaUserSpec(resourceData)
	if err != nil {
		require.NoError(t, err)
	}

	expected := &kafka.UserSpec{
		Name:        "events_user",
		Password:    "test_pwd",
		Permissions: nil,
	}

	assert.Equal(t, expected, userSpec)
}

func TestBuildKafkaUserSpec(t *testing.T) {
	raw := map[string]interface{}{
		"name":     "events_user",
		"password": "test_pwd",
		"permission": []interface{}{
			map[string]interface{}{
				"topic_name":  "topic1",
				"role":        "ACCESS_ROLE_PRODUCER",
				"allow_hosts": []interface{}{"host1", "host2"},
			},
			map[string]interface{}{
				"topic_name":  "topic2",
				"role":        "ACCESS_ROLE_CONSUMER",
				"allow_hosts": []interface{}{"host3", "host4"},
			},
		},
	}
	resourceData := schema.TestResourceDataRaw(t, resourceYandexMDBKafkaUser().Schema, raw)

	userSpec, err := buildKafkaUserSpec(resourceData)
	if err != nil {
		require.NoError(t, err)
	}

	expected := &kafka.UserSpec{
		Name:     "events_user",
		Password: "test_pwd",
		Permissions: []*kafka.Permission{
			{
				TopicName:  "topic1",
				Role:       kafka.Permission_ACCESS_ROLE_PRODUCER,
				AllowHosts: []string{"host1", "host2"},
			},
			{
				TopicName:  "topic2",
				Role:       kafka.Permission_ACCESS_ROLE_CONSUMER,
				AllowHosts: []string{"host3", "host4"},
			},
		},
	}

	assert.Equal(t, expected, userSpec)
}

func TestAccMDBKafkaUser(t *testing.T) {
	t.Parallel()
	clusterName := acctest.RandomWithPrefix("tf-kafka")
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBKafkaUserConfigStep1(clusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBKafkaClusterHasUser("events-user"),
					testAccCheckMDBKafkaClusterHasUser("another-user"),
					testAccCheckMDBKafkaUserHasPermissions("events-user", []*kafka.Permission{
						{
							TopicName:  "raw_events",
							Role:       kafka.Permission_ACCESS_ROLE_PRODUCER,
							AllowHosts: []string{"host1.db.yandex.net", "host2.db.yandex.net"},
						},
						{
							TopicName:  "raw_events",
							Role:       kafka.Permission_ACCESS_ROLE_CONSUMER,
							AllowHosts: []string{"host3.db.yandex.net"},
						},
					}),
					testAccCheckMDBKafkaUserHasPermissions("another-user", []*kafka.Permission{
						{
							TopicName:  "raw_events",
							Role:       kafka.Permission_ACCESS_ROLE_PRODUCER,
							AllowHosts: []string{"host3.db.yandex.net"},
						},
					}),
				),
			},
			mdbKafkaUserImportStep("yandex_mdb_kafka_user.events_user"),
			mdbKafkaUserImportStep("yandex_mdb_kafka_user.another_user"),
			{
				Config: testAccMDBKafkaUserConfigStep2(clusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBKafkaClusterHasUser("events-user"),
					testAccCheckMDBKafkaClusterDoesNotHaveUser("another-user"),
					testAccCheckMDBKafkaUserHasPermissions("events-user", []*kafka.Permission{
						{
							TopicName:  "raw_events",
							Role:       kafka.Permission_ACCESS_ROLE_PRODUCER,
							AllowHosts: []string{"host1.db.yandex.net", "host2.db.yandex.net", "host3.db.yandex.net", "host4.db.yandex.net"},
						},
						{
							TopicName: "raw_events",
							Role:      kafka.Permission_ACCESS_ROLE_CONSUMER,
						},
					}),
				),
			},
		},
	})
}

func mdbKafkaUserImportStep(name string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      name,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateVerifyIgnore: []string{
			"password", // password are not returned
		},
	}
}

func testAccMDBKafkaUserConfigStep0(name string) string {
	return fmt.Sprintf(kfVPCDependencies+`
resource "yandex_mdb_kafka_cluster" "foo" {
	name        = "%s"
	description = "Kafka User Terraform Test"
	environment = "PRODUCTION"
	network_id  = yandex_vpc_network.mdb-kafka-test-net.id
	subnet_ids = [yandex_vpc_subnet.mdb-kafka-test-subnet-a.id]

	config {
	  version          = "%s"
	  brokers_count    = 1
	  zones            = ["ru-central1-a"]
	  kafka {
		resources {
		  resource_preset_id = "s2.micro"
		  disk_type_id       = "network-hdd"
		  disk_size          = 16
		}

		kafka_config {
		  log_segment_bytes = 104857600
		}
	  }
	}

	topic {
	  name               = "raw_events"
	  partitions         = 1
	  replication_factor = 1
	  topic_config {
		cleanup_policy    = "CLEANUP_POLICY_COMPACT_AND_DELETE"
		max_message_bytes = 777216
		segment_bytes     = 134217728
		flush_ms          = 9223372036854775807
	  }
	}
}
`, name, currentDefaultKafkaVersion)
}

func testAccMDBKafkaUserConfigStep1(name string) string {
	return testAccMDBKafkaUserConfigStep0(name) + `
resource "yandex_mdb_kafka_user" events_user {
  cluster_id = yandex_mdb_kafka_cluster.foo.id
  name       = "events-user"
  password   = "test-password-123"
  permission {
	topic_name  = "raw_events"
	role        = "ACCESS_ROLE_PRODUCER"
    allow_hosts = ["host1.db.yandex.net", "host2.db.yandex.net"] 
  }
  permission {
	topic_name  = "raw_events"
	role        = "ACCESS_ROLE_CONSUMER"
    allow_hosts = ["host3.db.yandex.net"]
  }
}

resource "yandex_mdb_kafka_user" another_user {
  cluster_id = yandex_mdb_kafka_cluster.foo.id
  name       = "another-user"
  password   = "test-password-123"
  permission {
	topic_name  = "raw_events"
	role        = "ACCESS_ROLE_PRODUCER"
    allow_hosts = ["host3.db.yandex.net"]
  }
}
`
}

func testAccMDBKafkaUserConfigStep2(name string) string {
	return testAccMDBKafkaUserConfigStep0(name) + `
resource "yandex_mdb_kafka_user" events_user {
  cluster_id = yandex_mdb_kafka_cluster.foo.id
  name       = "events-user"
  password   = "test-password-1234"
  permission {
	topic_name  = "raw_events"
	role        = "ACCESS_ROLE_PRODUCER"
    allow_hosts = ["host1.db.yandex.net", "host2.db.yandex.net", "host3.db.yandex.net", "host4.db.yandex.net"]
  }
  permission {
	topic_name = "raw_events"
	role       = "ACCESS_ROLE_CONSUMER"
  }
}
`
}

func testAccLoadKafkaUser(s *terraform.State, userName string) (*kafka.User, error) {
	rs, ok := s.RootModule().Resources[kafkaClusterResourceName]
	if !ok {
		return nil, fmt.Errorf("resource %q not found", kafkaClusterResourceName)
	}

	if rs.Primary.ID == "" {
		return nil, fmt.Errorf("no ID is set")
	}

	config := testAccProvider.Meta().(*Config)
	return config.sdk.MDB().Kafka().User().Get(context.Background(), &kafka.GetUserRequest{
		ClusterId: rs.Primary.ID,
		UserName:  userName,
	})
}

func testAccCheckMDBKafkaClusterDoesNotHaveUser(userName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := testAccLoadKafkaUser(s, userName)
		if err == nil {
			return fmt.Errorf("expected user %q to be absent but it exists", userName)
		}
		if !isStatusWithCode(err, codes.NotFound) {
			return err
		}
		return nil
	}
}

func testAccCheckMDBKafkaClusterHasUser(userName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := testAccLoadKafkaUser(s, userName)
		return err
	}
}

func testAccCheckMDBKafkaUserHasPermissions(userName string, permissions []*kafka.Permission) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		user, err := testAccLoadKafkaUser(s, userName)
		if err != nil {
			return err
		}
		actualUserPermissions := user.GetPermissions()
		permissionsStr := UserPermissionsToStr(permissions)
		actualUserPermissionsStr := UserPermissionsToStr(actualUserPermissions)
		if permissionsStr != actualUserPermissionsStr {
			return fmt.Errorf("user %q has permissions %q, expected: %q", userName, actualUserPermissionsStr, permissionsStr)
		}
		return nil
	}
}
