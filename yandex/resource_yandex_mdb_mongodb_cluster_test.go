package yandex

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mongodb/v1"
)

const mongodbResource = "yandex_mdb_mongodb_cluster.foo"

const mongodbVPCDependencies = `
resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_vpc_subnet" "bar" {
  zone           = "ru-central1-b"
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["10.2.0.0/24"]
}

resource "yandex_vpc_subnet" "baz" {
  zone           = "ru-central1-c"
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["10.3.0.0/24"]
}

resource "yandex_vpc_security_group" "sg-x" {
  network_id     = "${yandex_vpc_network.foo.id}"
  ingress {
    protocol          = "ANY"
    description       = "Allow incoming traffic from members of the same security group"
    from_port         = 0
    to_port           = 65535
    v4_cidr_blocks    = ["0.0.0.0/0"]
  }
  egress {
    protocol          = "ANY"
    description       = "Allow outgoing traffic to members of the same security group"
    from_port         = 0
    to_port           = 65535
    v4_cidr_blocks    = ["0.0.0.0/0"]
  }
}

resource "yandex_vpc_security_group" "sg-y" {
  network_id     = "${yandex_vpc_network.foo.id}"
  
  ingress {
    protocol          = "ANY"
    description       = "Allow incoming traffic from members of the same security group"
    from_port         = 0
    to_port           = 65535
    v4_cidr_blocks    = ["0.0.0.0/0"]
  }
  egress {
    protocol          = "ANY"
    description       = "Allow outgoing traffic to members of the same security group"
    from_port         = 0
    to_port           = 65535
    v4_cidr_blocks    = ["0.0.0.0/0"]
  }
}
`

const resourceYandexMdbMongodbClusterTemplateText = mongodbVPCDependencies + `
resource "yandex_mdb_mongodb_cluster" "foo" {
  name        = "{{.ClusterName}}"
{{- if .ClusterDescription}}
  description = "{{.ClusterDescription}}"
{{- end}}
  environment = "{{.Environment}}"
  network_id  = "${yandex_vpc_network.foo.id}"

{{if .Lables}}
  labels = {
{{- range $key, $value := .Lables}}
    {{ $key }} = "{{ $value }}"
{{- end}}
  }
{{end}}

  cluster_config {
    version = "{{.Version}}"
    feature_compatibility_version = "{{dropSuffix .Version "-enterprise"}}"
{{if .Access}}
	access {
	{{- range $key, $value := .Access}}
		{{ $key }} = "{{ $value }}"
	{{- end}}
	} 
{{end}}
    backup_window_start {
      hours = {{.BackupWindow.hours}}
      minutes = {{.BackupWindow.minutes}}
    }
{{if .Mongod}}
    mongod {
{{if .Mongod.AuditLog}}
      audit_log {
        filter = "{{escapeQuotations .Mongod.AuditLog.Filter}}"
      }
      set_parameter {
        audit_authorization_success = {{.Mongod.SetParameter.AuditAuthorizationSuccess}}
      }
{{end}}
    }
{{end}}
  }

{{range $i, $r := .Databases}}
  database {
    name = "{{.}}"
  }
{{- end}}

{{range $i, $r := .Users}}
  user {
    name     = "{{$r.Name}}"
    password = "{{$r.Password}}"
{{range $ii, $rr := $r.Permissions}}
    permission {
      database_name = "{{$rr.DatabaseName}}"
      {{if $rr.Roles -}}
      roles = [{{range $iii, $rrr := $rr.Roles}}{{if $iii}}, {{end}}"{{.}}"{{end}}]
      {{- end}}
    }
{{- end}}
  }
{{- end}}

  resources {
    resource_preset_id = "{{.Resources.ResourcePresetId}}"
    disk_size          = {{.Resources.DiskSize}}
    disk_type_id       = "{{.Resources.DiskTypeId}}"
  }

{{range $i, $r := .Hosts}}
  host {
    zone_id   = "{{$r.ZoneId}}"
    subnet_id = "{{$r.SubnetId}}"
  }
{{end}}

  security_group_ids = [{{range $i, $r := .SecurityGroupIds}}{{if $i}}, {{end}}"{{.}}"{{end}}]

  maintenance_window {
    type = "{{.MaintenanceWindow.Type}}"
    {{with .MaintenanceWindow.Day}}day  = "{{.}}"{{end}}
    {{with .MaintenanceWindow.Hour}}hour = {{.}}{{end}}
  }

  {{if isNotNil .DeletionProtection}}deletion_protection = {{.DeletionProtection}}{{end}}
}
`

func makeConfigFromTemplateText(t *testing.T, templateText string, data *map[string]interface{}, patch *map[string]interface{}) string {
	if patch != nil {
		for k, v := range *patch {
			if v != nil {
				(*data)[k] = v
			} else {
				delete(*data, k)
			}
		}
	}

	tmpl, err := template.New("config").Funcs(template.FuncMap{
		"isNotNil":         func(v interface{}) bool { return v != nil },
		"escapeQuotations": func(v string) string { return strings.Replace(v, "\"", "\\\"", -1) },
		"dropSuffix": func(v string, suffix string) string {
			if strings.HasSuffix(v, suffix) {
				return v[0 : len(v)-len(suffix)]
			}
			return v
		},
	}).Parse(templateText)
	if err != nil {
		t.Fatal(err)
		return ""
	}
	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, data)
	if err != nil {
		t.Fatal(err)
		return ""
	}
	result := buf.String()
	return result
}

func makeConfig(t *testing.T, data *map[string]interface{}, patch *map[string]interface{}) string {
	return makeConfigFromTemplateText(t, resourceYandexMdbMongodbClusterTemplateText, data, patch)
}

var s2Micro16hdd = mongodb.Resources{
	ResourcePresetId: "s2.micro",
	DiskSize:         toBytes(16),
	DiskTypeId:       "network-hdd",
}

var s2Small26hdd = mongodb.Resources{
	ResourcePresetId: "s2.small",
	DiskSize:         toBytes(26),
	DiskTypeId:       "network-hdd",
}

func init() {
	resource.AddTestSweepers("yandex_mdb_mongodb_cluster", &resource.Sweeper{
		Name: "yandex_mdb_mongodb_cluster",
		F:    testSweepMDBMongoDBCluster,
	})
}

func testSweepMDBMongoDBCluster(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := conf.sdk.MDB().MongoDB().Cluster().List(conf.Context(), &mongodb.ListClustersRequest{
		FolderId: conf.FolderID,
		PageSize: defaultMDBPageSize,
	})
	if err != nil {
		return fmt.Errorf("error getting MongoDB clusters: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Clusters {
		if !sweepMDBMongoDBCluster(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep MongoDB cluster %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepMDBMongoDBCluster(conf *Config, id string) bool {
	return sweepWithRetry(sweepMDBMongoDBClusterOnce, conf, "MongoDB cluster", id)
}

func sweepMDBMongoDBClusterOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(*schema.DefaultTimeout(30 * time.Minute))
	defer cancel()

	mask := field_mask.FieldMask{Paths: []string{"deletion_protection"}}
	op, err := conf.sdk.MDB().MongoDB().Cluster().Update(ctx, &mongodb.UpdateClusterRequest{
		ClusterId:          id,
		DeletionProtection: false,
		UpdateMask:         &mask,
	})
	err = handleSweepOperation(ctx, conf, op, err)
	if err != nil && !strings.EqualFold(errorMessage(err), "no changes detected") {
		return err
	}

	op, err = conf.sdk.MDB().MongoDB().Cluster().Delete(ctx, &mongodb.DeleteClusterRequest{
		ClusterId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

func mdbMongoDBClusterImportStep() resource.TestStep {
	return resource.TestStep{
		ResourceName:      mongodbResource,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateVerifyIgnore: []string{
			"user",
			"health", // volatile value
			"host",   // order may differ
		},
	}
}

func create4_2ConfigData() map[string]interface{} {
	return map[string]interface{}{
		"Version":     "4.2",
		"ClusterName": acctest.RandomWithPrefix("test-acc-tf-mongodb"),
		"Environment": "PRESTABLE",
		"Lables":      map[string]string{"test_key": "test_value"},
		"BackupWindow": map[string]int64{
			"hours":   3,
			"minutes": 4,
		},
		"Access": map[string]bool{
			"data_lens":     true,
			"data_transfer": true,
		},
		"Databases": []string{"testdb"},
		"Users": []*mongodb.UserSpec{
			{
				Name:     "john",
				Password: "password",
				Permissions: []*mongodb.Permission{
					{
						DatabaseName: "testdb",
					},
				},
			},
		},
		"Resources": &mongodb.Resources{
			ResourcePresetId: s2Micro16hdd.ResourcePresetId,
			DiskSize:         s2Micro16hdd.DiskSize >> 30,
			DiskTypeId:       s2Micro16hdd.DiskTypeId,
		},
		"Hosts": []map[string]interface{}{
			{"ZoneId": "ru-central1-a", "SubnetId": "${yandex_vpc_subnet.foo.id}"},
			{"ZoneId": "ru-central1-b", "SubnetId": "${yandex_vpc_subnet.bar.id}"},
		},
		"SecurityGroupIds": []string{"${yandex_vpc_security_group.sg-x.id}"},
		"MaintenanceWindow": map[string]interface{}{
			"Type": "WEEKLY",
			"Day":  "FRI",
			"Hour": 20,
		},
		"DeletionProtection": true,
	}
}

// Test that a MongoDB Cluster can be created, updated and destroyed
func TestAccMDBMongoDBCluster_4_2(t *testing.T) {
	t.Parallel()

	configData := create4_2ConfigData()
	clusterName := configData["ClusterName"].(string)

	var r mongodb.Cluster
	folderID := getExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckVPCNetworkDestroy,
		Steps: []resource.TestStep{
			// Create MongoDB Cluster
			{
				Config: makeConfig(t, &configData, nil),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMongoDBClusterExists(mongodbResource, &r, 2),
					resource.TestCheckResourceAttr(mongodbResource, "name", clusterName),
					resource.TestCheckResourceAttr(mongodbResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(mongodbResource, "cluster_config.0.access.0.data_lens", "true"),
					resource.TestCheckResourceAttr(mongodbResource, "cluster_config.0.access.0.data_transfer", "true"),
					testAccCheckMDBMongoDBClusterHasRightVersion(&r, configData["Version"].(string)),
					testAccCheckMDBMongoDBClusterHasMongodSpec(&r, map[string]interface{}{"Resources": &s2Micro16hdd}),
					testAccCheckMDBMongoDBClusterHasDatabases(mongodbResource, []string{"testdb"}),
					testAccCheckMDBMongoDBClusterHasUsers(mongodbResource, map[string][]string{"john": {"testdb"}}),
					testAccCheckMDBMongoDBClusterContainsLabel(&r, "test_key", "test_value"),
					testAccCheckCreatedAtAttr(mongodbResource),
					resource.TestCheckResourceAttr(mongodbResource, "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(mongodbResource, "maintenance_window.0.type", "WEEKLY"),
					resource.TestCheckResourceAttr(mongodbResource, "maintenance_window.0.day", "FRI"),
					resource.TestCheckResourceAttr(mongodbResource, "maintenance_window.0.hour", "20"),
					resource.TestCheckResourceAttr(mongodbResource, "deletion_protection", "true"),
				),
			},
			mdbMongoDBClusterImportStep(),
			// uncheck 'deletion_protection'
			{
				Config: makeConfig(t, &configData, &map[string]interface{}{"DeletionProtection": false}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMongoDBClusterExists(mongodbResource, &r, 2),
					resource.TestCheckResourceAttr(mongodbResource, "deletion_protection", "false"),
				),
			},
			mdbMongoDBClusterImportStep(),
			// check 'deletion_protection'
			{
				Config: makeConfig(t, &configData, &map[string]interface{}{
					"DeletionProtection": true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMongoDBClusterExists(mongodbResource, &r, 2),
					resource.TestCheckResourceAttr(mongodbResource, "deletion_protection", "true"),
				),
			},
			mdbMongoDBClusterImportStep(),
			// trigger deletion by changing environment
			{
				Config: makeConfig(t, &configData, &map[string]interface{}{
					"Environment": "PRODUCTION",
				}),
				ExpectError: regexp.MustCompile(".*The operation was rejected because cluster has 'deletion_protection' = ON.*"),
			},
			// uncheck 'deletion_protection'
			{
				Config: makeConfig(t, &configData, &map[string]interface{}{
					"Environment":        "PRESTABLE",
					"DeletionProtection": false,
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMongoDBClusterExists(mongodbResource, &r, 2),
					resource.TestCheckResourceAttr(mongodbResource, "deletion_protection", "false"),
				),
			},
			mdbMongoDBClusterImportStep(),
			{
				Config: makeConfig(t, &configData, &map[string]interface{}{
					"MaintenanceWindow": map[string]interface{}{"Type": "ANYTIME"},
					"SecurityGroupIds": []string{
						"${yandex_vpc_security_group.sg-x.id}",
						"${yandex_vpc_security_group.sg-y.id}",
					},
					"Users": []*mongodb.UserSpec{
						{
							Name:     "john",
							Password: "password",
							Permissions: []*mongodb.Permission{
								{
									DatabaseName: "admin",
									Roles:        []string{"mdbMonitor"},
								},
							},
						},
					},
					"DeletionProtection": nil,
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMongoDBClusterExists(mongodbResource, &r, 2),
					resource.TestCheckResourceAttr(mongodbResource, "name", clusterName),
					resource.TestCheckResourceAttr(mongodbResource, "folder_id", folderID),
					testAccCheckMDBMongoDBClusterHasMongodSpec(&r, map[string]interface{}{"Resources": &s2Micro16hdd}),
					testAccCheckMDBMongoDBClusterHasUsers(mongodbResource, map[string][]string{"john": {"admin"}}),
					testAccCheckMDBMongoDBClusterHasDatabases(mongodbResource, []string{"testdb"}),
					testAccCheckCreatedAtAttr(mongodbResource),
					resource.TestCheckResourceAttr(mongodbResource, "security_group_ids.#", "2"),
					resource.TestCheckResourceAttr(mongodbResource, "maintenance_window.0.type", "ANYTIME"),
				),
			},
			mdbMongoDBClusterImportStep(),
			{
				Config: makeConfig(t, &configData, &map[string]interface{}{
					"ClusterName":        clusterName + "-changed",
					"ClusterDescription": "Updated MongDB cluster",
					"Lables":             map[string]string{"new_key": "new_value"},
					"Databases":          []string{"testdb", "newdb"},
					"Users": []*mongodb.UserSpec{
						{
							Name:     "john",
							Password: "password",
							Permissions: []*mongodb.Permission{
								{
									DatabaseName: "admin",
									Roles:        []string{"mdbMonitor"},
								},
							},
						},
						{
							Name:     "mary",
							Password: "qwerty123",
							Permissions: []*mongodb.Permission{
								{
									DatabaseName: "newdb",
								},
								{
									DatabaseName: "admin",
									Roles:        []string{"mdbMonitor"},
								},
							},
						},
					},
					"Resources": &mongodb.Resources{
						ResourcePresetId: s2Small26hdd.ResourcePresetId,
						DiskSize:         s2Small26hdd.DiskSize >> 30,
						DiskTypeId:       s2Small26hdd.DiskTypeId,
					},
					"Hosts": []map[string]interface{}{
						{"ZoneId": "ru-central1-c", "SubnetId": "${yandex_vpc_subnet.baz.id}"},
						{"ZoneId": "ru-central1-b", "SubnetId": "${yandex_vpc_subnet.bar.id}"},
					},
					"SecurityGroupIds": []string{"${yandex_vpc_security_group.sg-y.id}"},
					"MaintenanceWindow": map[string]interface{}{
						"Type": "WEEKLY",
						"Day":  "FRI",
						"Hour": 20,
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMongoDBClusterExists(mongodbResource, &r, 2),
					resource.TestCheckResourceAttr(mongodbResource, "name", clusterName+"-changed"),
					resource.TestCheckResourceAttr(mongodbResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(mongodbResource, "description", "Updated MongDB cluster"),
					resource.TestCheckResourceAttrSet(mongodbResource, "host.0.name"),
					testAccCheckMDBMongoDBClusterContainsLabel(&r, "new_key", "new_value"),
					testAccCheckMDBMongoDBClusterHasMongodSpec(&r, map[string]interface{}{"Resources": &s2Small26hdd}),
					testAccCheckMDBMongoDBClusterHasUsers(mongodbResource, map[string][]string{"john": {"admin"}, "mary": {"newdb", "admin"}}),
					testAccCheckMDBMongoDBClusterHasDatabases(mongodbResource, []string{"testdb", "newdb"}),
					testAccCheckCreatedAtAttr(mongodbResource),
					resource.TestCheckResourceAttr(mongodbResource, "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(mongodbResource, "maintenance_window.0.type", "WEEKLY"),
					resource.TestCheckResourceAttr(mongodbResource, "maintenance_window.0.day", "FRI"),
					resource.TestCheckResourceAttr(mongodbResource, "maintenance_window.0.hour", "20"),
				),
			},
			mdbMongoDBClusterImportStep(),
		},
	})
}

func create5_0_enterpriseConfigData() map[string]interface{} {
	return map[string]interface{}{
		"Version":     "5.0-enterprise",
		"ClusterName": acctest.RandomWithPrefix("test-acc-tf-mongodb"),
		"Environment": "PRESTABLE",
		"Lables":      map[string]string{"test_key": "test_value"},
		"BackupWindow": map[string]int64{
			"hours":   3,
			"minutes": 4,
		},
		"Mongod": map[string]interface{}{
			"AuditLog": map[string]interface{}{
				"Filter": "{ \"atype\": { \"$in\": [ \"createCollection\", \"dropCollection\" ] } }",
			},
			"SetParameter": map[string]interface{}{
				"AuditAuthorizationSuccess": true,
			},
		},
		"Databases": []string{"testdb"},
		"Users": []*mongodb.UserSpec{
			{
				Name:     "john",
				Password: "password",
				Permissions: []*mongodb.Permission{
					{
						DatabaseName: "testdb",
					},
				},
			},
		},
		"Resources": &mongodb.Resources{
			ResourcePresetId: s2Micro16hdd.ResourcePresetId,
			DiskSize:         s2Micro16hdd.DiskSize >> 30,
			DiskTypeId:       s2Micro16hdd.DiskTypeId,
		},
		"Hosts": []map[string]interface{}{
			{"ZoneId": "ru-central1-a", "SubnetId": "${yandex_vpc_subnet.foo.id}"},
			{"ZoneId": "ru-central1-b", "SubnetId": "${yandex_vpc_subnet.bar.id}"},
		},
		"SecurityGroupIds": []string{"${yandex_vpc_security_group.sg-x.id}"},
		"MaintenanceWindow": map[string]interface{}{
			"Type": "WEEKLY",
			"Day":  "FRI",
			"Hour": 20,
		},
		"DeletionProtection": true,
	}
}

// Test that a MongoDB Cluster can be created, updated and destroyed
func TestAccMDBMongoDBCluster_5_0_enterprise(t *testing.T) {
	t.Parallel()

	configData := create5_0_enterpriseConfigData()
	clusterName := configData["ClusterName"].(string)
	version := configData["Version"].(string)
	auditLogFilter := ((configData["Mongod"].(map[string]interface{}))["AuditLog"].(map[string]interface{}))["Filter"].(string)

	var testCluster mongodb.Cluster
	folderID := getExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckMDBMongoDBClusterDestroy,
			testAccCheckVPCNetworkDestroy,
		),
		Steps: []resource.TestStep{
			// Create
			{
				Config: makeConfig(t, &configData, nil),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMongoDBClusterExists(mongodbResource, &testCluster, 2),
					resource.TestCheckResourceAttr(mongodbResource, "name", clusterName),
					resource.TestCheckResourceAttr(mongodbResource, "folder_id", folderID),
					testAccCheckMDBMongoDBClusterHasRightVersion(&testCluster, version),
					testAccCheckMDBMongoDBClusterHasMongodSpec(&testCluster, map[string]interface{}{
						"Resources":                 &s2Micro16hdd,
						"AuditLogFilter":            auditLogFilter,
						"AuditAuthorizationSuccess": true,
					}),
					testAccCheckMDBMongoDBClusterHasDatabases(mongodbResource, []string{"testdb"}),
					testAccCheckMDBMongoDBClusterHasUsers(mongodbResource, map[string][]string{"john": {"testdb"}}),
					testAccCheckMDBMongoDBClusterContainsLabel(&testCluster, "test_key", "test_value"),
					testAccCheckCreatedAtAttr(mongodbResource),
					resource.TestCheckResourceAttr(mongodbResource, "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(mongodbResource, "maintenance_window.0.type", "WEEKLY"),
					resource.TestCheckResourceAttr(mongodbResource, "maintenance_window.0.day", "FRI"),
					resource.TestCheckResourceAttr(mongodbResource, "maintenance_window.0.hour", "20"),
					resource.TestCheckResourceAttr(mongodbResource, "deletion_protection", "true"),
					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.audit_log.0.filter", auditLogFilter),
					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.set_parameter.0.audit_authorization_success", "true"),
				),
			},
			mdbMongoDBClusterImportStep(),
			{ // Uncheck deletion_protection
				Config: makeConfig(t, &configData, &map[string]interface{}{
					"DeletionProtection": false,
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMongoDBClusterExists(mongodbResource, &testCluster, 2),
					resource.TestCheckResourceAttr(mongodbResource, "deletion_protection", "false"),
				),
			},
			mdbMongoDBClusterImportStep(),
			{ // Update: remove filter and uncheck AuditAuthorizationSuccess
				Config: makeConfig(t, &configData, &map[string]interface{}{
					"Mongod": map[string]interface{}{
						"AuditLog": map[string]interface{}{
							"Filter": "{}",
						},
						"SetParameter": map[string]interface{}{
							"AuditAuthorizationSuccess": false,
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMongoDBClusterExists(mongodbResource, &testCluster, 2),
					resource.TestCheckResourceAttr(mongodbResource, "name", clusterName),
					resource.TestCheckResourceAttr(mongodbResource, "folder_id", folderID),
					testAccCheckMDBMongoDBClusterHasRightVersion(&testCluster, version),
					testAccCheckMDBMongoDBClusterHasMongodSpec(&testCluster, map[string]interface{}{
						"Resources":                 &s2Micro16hdd,
						"AuditLogFilter":            "{}",
						"AuditAuthorizationSuccess": false,
					}),
					testAccCheckMDBMongoDBClusterHasDatabases(mongodbResource, []string{"testdb"}),
					testAccCheckMDBMongoDBClusterHasUsers(mongodbResource, map[string][]string{"john": {"testdb"}}),
					testAccCheckMDBMongoDBClusterContainsLabel(&testCluster, "test_key", "test_value"),
					testAccCheckCreatedAtAttr(mongodbResource),
					resource.TestCheckResourceAttr(mongodbResource, "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(mongodbResource, "maintenance_window.0.type", "WEEKLY"),
					resource.TestCheckResourceAttr(mongodbResource, "maintenance_window.0.day", "FRI"),
					resource.TestCheckResourceAttr(mongodbResource, "maintenance_window.0.hour", "20"),
					resource.TestCheckResourceAttr(mongodbResource, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.audit_log.0.filter", "{}"),
					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.set_parameter.0.audit_authorization_success", "false"),
				),
			},
			mdbMongoDBClusterImportStep(),
		},
	})
}

func testAccCheckMDBMongoDBClusterDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_mdb_mongodb_cluster" {
			continue
		}

		_, err := config.sdk.MDB().MongoDB().Cluster().Get(context.Background(), &mongodb.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})

		if err == nil {
			return fmt.Errorf("MongoDB Cluster still exists")
		}
	}

	return nil
}

func testAccCheckMDBMongoDBClusterExists(n string, r *mongodb.Cluster, hosts int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.MDB().MongoDB().Cluster().Get(context.Background(), &mongodb.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("MongoDB Cluster not found")
		}

		//goland:noinspection GoVetCopyLock (this comment suppress warning in Idea IDE about coping sync.Mutex)
		*r = *found

		resp, err := config.sdk.MDB().MongoDB().Cluster().ListHosts(context.Background(), &mongodb.ListClusterHostsRequest{
			ClusterId: rs.Primary.ID,
			PageSize:  defaultMDBPageSize,
		})
		if err != nil {
			return err
		}

		if len(resp.Hosts) != hosts {
			return fmt.Errorf("Expected %d hosts, got %d", hosts, len(resp.Hosts))
		}

		return nil
	}
}

func testAccCheckMDBMongoDBClusterHasRightVersion(r *mongodb.Cluster, version string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if r.Config.Version != version {
			return fmt.Errorf("Expected version '%s', got '%s'", version, r.Config.Version)
		}

		return nil
	}
}

func supportTestResources(actual *mongodb.Resources, expected *mongodb.Resources) error {
	if actual.ResourcePresetId != expected.ResourcePresetId {
		return fmt.Errorf("Expected resource preset id '%s', got '%s'",
			expected.ResourcePresetId, actual.ResourcePresetId)
	}
	if actual.DiskSize != expected.DiskSize {
		return fmt.Errorf("Expected size '%d', got '%d'", expected.DiskSize, actual.DiskSize)
	}

	if actual.DiskTypeId != expected.DiskTypeId {
		return fmt.Errorf("Expected disk type id '%s', got '%s'", expected.DiskTypeId, actual.DiskTypeId)
	}

	return nil
}

func testAccCheckMDBMongoDBClusterHasMongodSpec(r *mongodb.Cluster, expected map[string]interface{}) resource.TestCheckFunc {
	//TODO for future updates: test for different resources (mongod, mongos and mongocfg)
	expectedResources := expected["Resources"].(*mongodb.Resources)
	return func(s *terraform.State) error {
		switch r.Config.Version {
		case "5.0-enterprise":
			{
				mongo := r.Config.Mongodb.(*mongodb.ClusterConfig_Mongodb_5_0Enterprise).Mongodb_5_0Enterprise
				d := mongo.Mongod
				if d != nil {
					err := supportTestResources(d.Resources, expectedResources)
					if err != nil {
						return err
					}
					if expectedValue, ok := expected["AuditLogFilter"]; ok {
						actual := d.Config.UserConfig.AuditLog.Filter
						expected := expectedValue.(string)
						if actual != expected {
							return fmt.Errorf("Expected audit log filter '%s', got '%s'", expected, actual)
						}
					}
					if expectedValue, ok := expected["AuditAuthorizationSuccess"]; ok {
						expected := expectedValue.(bool)
						actual := d.Config.UserConfig.SetParameter.AuditAuthorizationSuccess.Value
						if actual != expected {
							return fmt.Errorf("Expected audit_authorization_success '%t', got '%t'", expected, actual)
						}
					}
				}

				s := mongo.Mongos
				if s != nil {
					err := supportTestResources(s.Resources, expectedResources)

					if err != nil {
						return err
					}
				}

				cfg := mongo.Mongocfg
				if cfg != nil {
					err := supportTestResources(cfg.Resources, expectedResources)

					if err != nil {
						return err
					}
				}
			}
		case "4.4-enterprise":
			{
				mongo := r.Config.Mongodb.(*mongodb.ClusterConfig_Mongodb_4_4Enterprise).Mongodb_4_4Enterprise
				d := mongo.Mongod
				if d != nil {
					err := supportTestResources(d.Resources, expectedResources)

					if err != nil {
						return err
					}
				}

				s := mongo.Mongos
				if s != nil {
					err := supportTestResources(s.Resources, expectedResources)

					if err != nil {
						return err
					}
				}

				cfg := mongo.Mongocfg
				if cfg != nil {
					err := supportTestResources(cfg.Resources, expectedResources)

					if err != nil {
						return err
					}
				}
			}
		case "5.0":
			{
				mongo := r.Config.Mongodb.(*mongodb.ClusterConfig_Mongodb_5_0).Mongodb_5_0
				d := mongo.Mongod
				if d != nil {
					err := supportTestResources(d.Resources, expectedResources)

					if err != nil {
						return err
					}
				}

				s := mongo.Mongos
				if s != nil {
					err := supportTestResources(s.Resources, expectedResources)

					if err != nil {
						return err
					}
				}

				cfg := mongo.Mongocfg
				if cfg != nil {
					err := supportTestResources(cfg.Resources, expectedResources)

					if err != nil {
						return err
					}
				}
			}
		case "4.4":
			{
				mongo := r.Config.Mongodb.(*mongodb.ClusterConfig_Mongodb_4_4).Mongodb_4_4
				d := mongo.Mongod
				if d != nil {
					err := supportTestResources(d.Resources, expectedResources)

					if err != nil {
						return err
					}
				}

				s := mongo.Mongos
				if s != nil {
					err := supportTestResources(s.Resources, expectedResources)

					if err != nil {
						return err
					}
				}

				cfg := mongo.Mongocfg
				if cfg != nil {
					err := supportTestResources(cfg.Resources, expectedResources)

					if err != nil {
						return err
					}
				}
			}
		case "4.2":
			{
				mongo := r.Config.Mongodb.(*mongodb.ClusterConfig_Mongodb_4_2).Mongodb_4_2
				d := mongo.Mongod
				if d != nil {
					err := supportTestResources(d.Resources, expectedResources)

					if err != nil {
						return err
					}
				}

				s := mongo.Mongos
				if s != nil {
					err := supportTestResources(s.Resources, expectedResources)

					if err != nil {
						return err
					}
				}

				cfg := mongo.Mongocfg
				if cfg != nil {
					err := supportTestResources(cfg.Resources, expectedResources)

					if err != nil {
						return err
					}
				}
			}
		case "4.0":
			{
				mongo := r.Config.Mongodb.(*mongodb.ClusterConfig_Mongodb_4_0).Mongodb_4_0
				d := mongo.Mongod
				if d != nil {
					err := supportTestResources(d.Resources, expectedResources)

					if err != nil {
						return err
					}
				}

				s := mongo.Mongos
				if s != nil {
					err := supportTestResources(s.Resources, expectedResources)

					if err != nil {
						return err
					}
				}

				cfg := mongo.Mongocfg
				if cfg != nil {
					err := supportTestResources(cfg.Resources, expectedResources)

					if err != nil {
						return err
					}
				}
			}
		case "3.6":
			{
				mongo := r.Config.Mongodb.(*mongodb.ClusterConfig_Mongodb_3_6).Mongodb_3_6
				d := mongo.Mongod
				if d != nil {
					err := supportTestResources(d.Resources, expectedResources)

					if err != nil {
						return err
					}
				}

				s := mongo.Mongos
				if s != nil {
					err := supportTestResources(s.Resources, expectedResources)

					if err != nil {
						return err
					}
				}

				cfg := mongo.Mongocfg
				if cfg != nil {
					err := supportTestResources(cfg.Resources, expectedResources)

					if err != nil {
						return err
					}
				}
			}
		}
		return nil
	}
}

func testAccCheckMDBMongoDBClusterContainsLabel(r *mongodb.Cluster, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v, ok := r.Labels[key]
		if !ok {
			return fmt.Errorf("Expected label with key '%s' not found", key)
		}
		if v != value {
			return fmt.Errorf("Incorrect label value for key '%s': expected '%s' but found '%s'", key, value, v)
		}
		return nil
	}
}

func testAccCheckMDBMongoDBClusterHasUsers(r string, perms map[string][]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("Not found: %s", r)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		resp, err := config.sdk.MDB().MongoDB().User().List(context.Background(), &mongodb.ListUsersRequest{
			ClusterId: rs.Primary.ID,
			PageSize:  defaultMDBPageSize,
		})
		if err != nil {
			return err
		}
		users := resp.Users

		if len(users) != len(perms) {
			return fmt.Errorf("Expected %d users, found %d", len(perms), len(users))
		}

		for _, u := range users {
			ps, ok := perms[u.Name]
			if !ok {
				return fmt.Errorf("Unexpected user: %s", u.Name)
			}

			var ups []string
			for _, p := range u.Permissions {
				ups = append(ups, p.DatabaseName)
			}

			sort.Strings(ps)
			sort.Strings(ups)
			if fmt.Sprintf("%v", ps) != fmt.Sprintf("%v", ups) {
				return fmt.Errorf("User %s has wrong permissions, %v. Expected %v", u.Name, ups, ps)
			}
		}

		return nil
	}
}

func testAccCheckMDBMongoDBClusterHasDatabases(r string, databases []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("Not found: %s", r)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		resp, err := config.sdk.MDB().MongoDB().Database().List(context.Background(), &mongodb.ListDatabasesRequest{
			ClusterId: rs.Primary.ID,
			PageSize:  defaultMDBPageSize,
		})
		if err != nil {
			return err
		}
		var dbs []string
		for _, d := range resp.Databases {
			dbs = append(dbs, d.Name)
		}

		if len(dbs) != len(databases) {
			return fmt.Errorf("Expected %d dbs, found %d", len(databases), len(dbs))
		}

		sort.Strings(dbs)
		sort.Strings(databases)
		if fmt.Sprintf("%v", dbs) != fmt.Sprintf("%v", databases) {
			return fmt.Errorf("Cluster has wrong databases, %v. Expected %v", dbs, databases)
		}

		return nil
	}
}
