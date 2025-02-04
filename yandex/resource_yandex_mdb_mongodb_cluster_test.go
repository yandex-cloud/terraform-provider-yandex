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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mongodb/v1"

	"golang.org/x/exp/maps"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const mongodbRestoreBackupId = "c9qvb4o0gnrh8ene82l7:c9qhh0gi4hn06qkdoqke"

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
  zone           = "ru-central1-d"
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

{{if .Restore}}
  restore {
	backup_id = "{{.Restore.BackupId}}"
	{{if .Restore.Time}}
	time = "{{.Restore.Time}}"
	{{end}}
  }
{{end}}

{{if .Lables}}
  labels = {
{{- range $key, $value := .Lables}}
    {{ $key }} = "{{ $value }}"
{{- end}}
  }
{{end}}

  cluster_config {
    version = "{{.Version}}"
{{ if .CompatibilityVersion}}
    feature_compatibility_version = "{{.CompatibilityVersion}}"
{{end}}
{{if .Access}}
	access {
	{{- range $key, $value := .Access}}
		{{ $key }} = "{{ $value }}"
	{{- end}}
	} 
{{end}}
{{if .PerformanceDiagnostics}}
	performance_diagnostics {
		enabled = "{{.PerformanceDiagnostics.enabled}}"
	} 
{{end}}
{{if .BackupWindow}}
    backup_window_start {
      hours = {{.BackupWindow.hours}}
      minutes = {{.BackupWindow.minutes}}
    }
{{end}}
{{if .BackupRetainPeriodDays}}
    backup_retain_period_days = {{.BackupRetainPeriodDays}}
{{end}}
{{if .Mongod}}
    mongod {
{{if .Mongod.AuditLog}}
      audit_log {
        filter = "{{escapeQuotations .Mongod.AuditLog.Filter}}"
      }
{{end}}
{{if .Mongod.SetParameter}}
      set_parameter {
				{{if .Mongod.SetParameter.EnableFlowControl}}
					enable_flow_control = {{.Mongod.SetParameter.EnableFlowControl}}
				{{end}}
				{{if .Mongod.SetParameter.AuditAuthorizationSuccess}}
					audit_authorization_success = {{.Mongod.SetParameter.AuditAuthorizationSuccess}}
				{{end}}
				{{if .Mongod.SetParameter.MinSnapshotHistoryWindowInSeconds}}
				    min_snapshot_history_window_in_seconds = {{.Mongod.SetParameter.MinSnapshotHistoryWindowInSeconds}}
				{{end}}
			}
{{end}}
{{if .Mongod.Net}}
      net {
        max_incoming_connections = "{{.Mongod.Net.MaxConnections}}"
      {{if .Mongod.Net.Compressors}}
				compressors = {{.Mongod.Net.Compressors}}
      {{end}}
      }
{{end}}
{{if .Mongod.Storage}}
	storage {
	{{if .Mongod.Storage.Journal}}
      journal {
        commit_interval = "{{.Mongod.Storage.Journal.CommitInterval}}"
      }
	{{end}}
	{{if .Mongod.Storage.WiredTiger}}
      wired_tiger {
        block_compressor = "{{.Mongod.Storage.WiredTiger.Compressor}}"
				prefix_compression = "{{.Mongod.Storage.WiredTiger.PrefixCompression}}"
      }
	{{end}}
	}
{{end}}
{{if .Mongod.OperationProfiling}}
	operation_profiling {
				{{if .Mongod.OperationProfiling.Mode}}
        		mode = "{{.Mongod.OperationProfiling.Mode}}"
				{{end}}
        slow_op_threshold = "{{.Mongod.OperationProfiling.OpThreshold}}"
				slow_op_sample_rate = "{{.Mongod.OperationProfiling.OpSampleRate}}"
	}
{{end}}
    }
{{end}}

{{if .Mongos}}
    mongos {
{{if .Mongos.Net}}
      net {
        max_incoming_connections = "{{.Mongos.Net.MaxConnections}}"
				{{if .Mongos.Net.Compressors}}
					compressors = {{.Mongos.Net.Compressors}}
      	{{end}}
      }
{{end}}
    }
{{end}}

{{if .MongoCfg}}
    mongocfg {
{{if .MongoCfg.Net}}
      net {
        max_incoming_connections = "{{.MongoCfg.Net.MaxConnections}}"
      }
{{end}}
{{if .MongoCfg.OperationProfiling}}
      operation_profiling {
        {{if .MongoCfg.OperationProfiling.Mode}}
        		mode = "{{.MongoCfg.OperationProfiling.Mode}}"
				{{end}}
        slow_op_threshold = "{{.MongoCfg.OperationProfiling.OpThreshold}}"
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

{{if .Resources}}
  resources {
    resource_preset_id = "{{.Resources.ResourcePresetId}}"
    disk_size          = {{.Resources.DiskSize}}
    disk_type_id       = "{{.Resources.DiskTypeId}}"
  }
{{end}}

{{if .ResourcesMongod}}
  resources_mongod {
    resource_preset_id = "{{.ResourcesMongod.ResourcePresetId}}"
    disk_size          = {{.ResourcesMongod.DiskSize}}
    disk_type_id       = "{{.ResourcesMongod.DiskTypeId}}"
  }
{{end}}

{{if .ResourcesMongoCfg}}
  resources_mongocfg {
    resource_preset_id = "{{.ResourcesMongoCfg.ResourcePresetId}}"
    disk_size          = {{.ResourcesMongoCfg.DiskSize}}
    disk_type_id       = "{{.ResourcesMongoCfg.DiskTypeId}}"
  }
{{end}}

{{if .ResourcesMongos}}
  resources_mongos {
    resource_preset_id = "{{.ResourcesMongos.ResourcePresetId}}"
    disk_size          = {{.ResourcesMongos.DiskSize}}
    disk_type_id       = "{{.ResourcesMongos.DiskTypeId}}"
  }
{{end}}

{{if .ResourcesMongoInfra}}
  resources_mongoinfra {
    resource_preset_id = "{{.ResourcesMongoInfra.ResourcePresetId}}"
    disk_size          = {{.ResourcesMongoInfra.DiskSize}}
    disk_type_id       = "{{.ResourcesMongoInfra.DiskTypeId}}"
  }
{{end}}

{{if .DiskSizeAutoscalingMongod}}
  disk_size_autoscaling_mongod {
    disk_size_limit           = {{.DiskSizeAutoscalingMongod.DiskSizeLimit.Value}}
	{{if .DiskSizeAutoscalingMongod.PlannedUsageThreshold}}
    planned_usage_threshold   = {{.DiskSizeAutoscalingMongod.PlannedUsageThreshold.Value}}
	{{end}}
    emergency_usage_threshold = {{.DiskSizeAutoscalingMongod.EmergencyUsageThreshold.Value}}
  }

{{end}}
{{if .DiskSizeAutoscalingMongoCfg}}
  disk_size_autoscaling_mongocfg {
    disk_size_limit           = {{.DiskSizeAutoscalingMongoCfg.DiskSizeLimit.Value}}
	{{if .DiskSizeAutoscalingMongoCfg.PlannedUsageThreshold}}
    planned_usage_threshold   = {{.DiskSizeAutoscalingMongoCfg.PlannedUsageThreshold.Value}}
	{{end}}
    emergency_usage_threshold = {{.DiskSizeAutoscalingMongoCfg.EmergencyUsageThreshold.Value}}
  }
{{end}}

{{if .DiskSizeAutoscalingMongos}}
  disk_size_autoscaling_mongos {
    disk_size_limit           = {{.DiskSizeAutoscalingMongos.DiskSizeLimit.Value}}
	{{if .DiskSizeAutoscalingMongos.PlannedUsageThreshold}}
    planned_usage_threshold   = {{.DiskSizeAutoscalingMongos.PlannedUsageThreshold.Value}}
	{{end}}
    emergency_usage_threshold = {{.DiskSizeAutoscalingMongos.EmergencyUsageThreshold.Value}}
  }
{{end}}

{{if .DiskSizeAutoscalingMongoInfra}}
  disk_size_autoscaling_mongoinfra {
    disk_size_limit           = {{.DiskSizeAutoscalingMongoInfra.DiskSizeLimit.Value}}
	{{if .DiskSizeAutoscalingMongoInfra.PlannedUsageThreshold}}
    planned_usage_threshold   = {{.DiskSizeAutoscalingMongoInfra.PlannedUsageThreshold.Value}}
	{{end}}
    emergency_usage_threshold = {{.DiskSizeAutoscalingMongoInfra.EmergencyUsageThreshold.Value}}
  }
{{end}}

{{range $i, $r := .Hosts}}
  host {
    zone_id   = "{{$r.ZoneId}}"
    subnet_id = "{{$r.SubnetId}}"
{{if $r.Type}}
	type 	  = "{{$r.Type}}"
{{end}}
{{if $r.ShardName}}
	shard_name 	  = "{{$r.ShardName}}"
{{end}}
{{if $r.AssignPublicIp}}
	assign_public_ip = "{{$r.AssignPublicIp}}"
{{end}}
{{if $r.HostParameters}}
	host_parameters {
{{if $r.HostParameters.Hidden}}
		hidden = "{{$r.HostParameters.Hidden}}"
{{end}}
{{if $r.HostParameters.Priority}}
		priority = "{{$r.HostParameters.Priority}}"
{{end}}
{{if $r.HostParameters.SecondaryDelaySecs}}
		secondary_delay_secs = "{{$r.HostParameters.SecondaryDelaySecs}}"
{{end}}
{{if $r.HostParameters.Tags}}
		tags = {
			{{- range $k, $v := $r.HostParameters.Tags}}
				{{$k}} = "{{$v}}"
			{{- end}}
		}
{{end}}
	}
{{end}}

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

var mongoHosts = []mongodb.Host{
	{
		ZoneId:         "ru-central1-a",
		SubnetId:       "${yandex_vpc_subnet.foo.id}",
		AssignPublicIp: false,
	},
	{
		ZoneId:         "ru-central1-b",
		SubnetId:       "${yandex_vpc_subnet.bar.id}",
		AssignPublicIp: true,
	},
}

var shardedMongoInfraHosts = []mongodb.Host{
	{
		ZoneId:   "ru-central1-a",
		SubnetId: "${yandex_vpc_subnet.foo.id}",
		Type:     mongodb.Host_MONGOD,
	},
	{
		ZoneId:   "ru-central1-b",
		SubnetId: "${yandex_vpc_subnet.bar.id}",
		Type:     mongodb.Host_MONGOD,
	},
	{
		ZoneId:   "ru-central1-a",
		SubnetId: "${yandex_vpc_subnet.foo.id}",
		Type:     mongodb.Host_MONGOINFRA,
	},
	{
		ZoneId:   "ru-central1-b",
		SubnetId: "${yandex_vpc_subnet.bar.id}",
		Type:     mongodb.Host_MONGOINFRA,
	},
	{
		ZoneId:   "ru-central1-a",
		SubnetId: "${yandex_vpc_subnet.foo.id}",
		Type:     mongodb.Host_MONGOINFRA,
	},
}

var shardedMongoCfgHosts = []mongodb.Host{
	{
		ZoneId:   "ru-central1-a",
		SubnetId: "${yandex_vpc_subnet.foo.id}",
		Type:     mongodb.Host_MONGOD,
	},
	{
		ZoneId:   "ru-central1-b",
		SubnetId: "${yandex_vpc_subnet.bar.id}",
		Type:     mongodb.Host_MONGOD,
	},
	{
		ZoneId:   "ru-central1-a",
		SubnetId: "${yandex_vpc_subnet.foo.id}",
		Type:     mongodb.Host_MONGOCFG,
	},
	{
		ZoneId:   "ru-central1-b",
		SubnetId: "${yandex_vpc_subnet.bar.id}",
		Type:     mongodb.Host_MONGOCFG,
	},
	{
		ZoneId:   "ru-central1-a",
		SubnetId: "${yandex_vpc_subnet.foo.id}",
		Type:     mongodb.Host_MONGOCFG,
	},
	{
		ZoneId:   "ru-central1-b",
		SubnetId: "${yandex_vpc_subnet.bar.id}",
		Type:     mongodb.Host_MONGOS,
	},
	{
		ZoneId:   "ru-central1-a",
		SubnetId: "${yandex_vpc_subnet.foo.id}",
		Type:     mongodb.Host_MONGOS,
	},
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

	result := &multierror.Group{}
	for _, c := range resp.Clusters {
		id := c.Id
		result.Go(func() error { return sweepMDBMongoDBCluster(conf, id) })
	}

	return result.Wait().ErrorOrNil()
}

func sweepMDBMongoDBCluster(conf *Config, id string) error {
	if !sweepWithRetry(sweepMDBMongoDBClusterOnce, conf, "MongoDB cluster", id) {
		return fmt.Errorf("failed to sweep MongoDB cluster %q", id)
	}
	return nil
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

func create6_0ConfigData() map[string]interface{} {
	return map[string]interface{}{
		"Version":              "6.0",
		"CompatibilityVersion": "6.0",
		"ClusterName":          acctest.RandomWithPrefix("test-acc-tf-mongodb"),
		"Environment":          "PRESTABLE",
		"Lables":               map[string]string{"test_key": "test_value"},
		"BackupWindow": map[string]int64{
			"hours":   3,
			"minutes": 4,
		},
		"BackupRetainPeriodDays": 10,
		"Access": map[string]bool{
			"data_lens":     true,
			"data_transfer": true,
		},
		"PerformanceDiagnostics": map[string]bool{
			"enabled": true,
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
		"ResourcesMongod": &mongodb.Resources{
			ResourcePresetId: s2Micro16hdd.ResourcePresetId,
			DiskSize:         s2Micro16hdd.DiskSize >> 30,
			DiskTypeId:       s2Micro16hdd.DiskTypeId,
		},
		"Hosts":            mongoHosts,
		"SecurityGroupIds": []string{"${yandex_vpc_security_group.sg-x.id}"},
		"MaintenanceWindow": map[string]interface{}{
			"Type": "WEEKLY",
			"Day":  "FRI",
			"Hour": 20,
		},
		"DeletionProtection": true,
	}
}

func createRestoreConfigData() map[string]interface{} {
	return map[string]interface{}{
		"Version":     "6.0",
		"ClusterName": acctest.RandomWithPrefix("test-acc-tf-mongodb"),
		"Restore": map[string]string{
			"BackupId": mongodbRestoreBackupId,
		},
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
		"PerformanceDiagnostics": map[string]bool{
			"enabled": true,
		},
		"Databases": []string{"db1"},
		"Users": []*mongodb.UserSpec{
			{
				Name:     "user1",
				Password: "password",
				Permissions: []*mongodb.Permission{
					{
						DatabaseName: "db1",
						Roles:        []string{"readWrite"},
					},
					{
						DatabaseName: "admin",
						Roles:        []string{"mdbShardingManager", "mdbMonitor"},
					},
				},
			},
		},
		"ResourcesMongod": &mongodb.Resources{
			ResourcePresetId: s2Micro16hdd.ResourcePresetId,
			DiskSize:         s2Micro16hdd.DiskSize >> 30,
			DiskTypeId:       s2Micro16hdd.DiskTypeId,
		},
		"Hosts":            mongoHosts,
		"SecurityGroupIds": []string{"${yandex_vpc_security_group.sg-x.id}"},
		"MaintenanceWindow": map[string]interface{}{
			"Type": "ANYTIME",
		},
		"DeletionProtection": true,
	}
}

// Test that a MongoDB Cluster can be created, updated and destroyed
func TestAccMDBMongoDBCluster_6_0(t *testing.T) {
	t.Parallel()

	configData := create6_0ConfigData()
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
					//todo set compatibility_version doesn't work in create method , change test after fix in api
					resource.TestCheckResourceAttr(mongodbResource, "cluster_config.0.feature_compatibility_version", "6.0"),
					resource.TestCheckResourceAttr(mongodbResource, "cluster_config.0.access.0.data_lens", "true"),
					resource.TestCheckResourceAttr(mongodbResource, "cluster_config.0.access.0.data_transfer", "true"),
					resource.TestCheckResourceAttr(mongodbResource, "cluster_config.0.backup_retain_period_days", "10"),
					resource.TestCheckResourceAttr(mongodbResource, "cluster_config.0.performance_diagnostics.0.enabled", "true"),
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
					resource.TestCheckResourceAttr(mongodbResource, "host.0.assign_public_ip", "false"),
					resource.TestCheckResourceAttr(mongodbResource, "host.1.assign_public_ip", "true"),
					resource.TestCheckResourceAttr(mongodbResource, "host.0.host_parameters.#", "1"),
					resource.TestCheckResourceAttr(mongodbResource, "host.1.host_parameters.#", "1"),
					testAccCheckMDBMongoDBClusterHasHostParameters(mongodbResource, map[string]interface{}{"tags": map[string]string{}, "hidden": false, "priority": 1.0, "secondary_delay_secs": int64(0)}, 2),
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
					"MaintenanceWindow":      map[string]interface{}{"Type": "ANYTIME"},
					"BackupRetainPeriodDays": "20",
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
					"DeletionProtection":   nil,
					"CompatibilityVersion": "5.0",
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
					resource.TestCheckResourceAttr(mongodbResource, "cluster_config.0.backup_retain_period_days", "20"),
					resource.TestCheckResourceAttr(mongodbResource, "cluster_config.0.feature_compatibility_version", "5.0"),
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
					"ResourcesMongod": &mongodb.Resources{
						ResourcePresetId: s2Small26hdd.ResourcePresetId,
						DiskSize:         s2Small26hdd.DiskSize >> 30,
						DiskTypeId:       s2Small26hdd.DiskTypeId,
					},
					"Hosts": []map[string]interface{}{
						{
							"ZoneId":   "ru-central1-d",
							"SubnetId": "${yandex_vpc_subnet.baz.id}",
							"HostParameters": map[string]interface{}{
								"Priority": 1.05,
							},
						},
						{
							"ZoneId":   "ru-central1-d",
							"SubnetId": "${yandex_vpc_subnet.baz.id}",
							"HostParameters": map[string]interface{}{
								"Priority": 1.15,
							},
						},
						{"ZoneId": "ru-central1-d", "SubnetId": "${yandex_vpc_subnet.baz.id}"},
						{
							"ZoneId":   "ru-central1-b",
							"SubnetId": "${yandex_vpc_subnet.bar.id}",
							"HostParameters": map[string]interface{}{
								"Priority": 0.0,
								"Hidden":   true,
								"Tags": map[string]interface{}{
									"abc": "def",
								},
							},
							"Type":      "MONGOD",
							"ShardName": "rs01",
						},
					},
					"SecurityGroupIds": []string{"${yandex_vpc_security_group.sg-y.id}"},
					"MaintenanceWindow": map[string]interface{}{
						"Type": "WEEKLY",
						"Day":  "FRI",
						"Hour": 20,
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMongoDBClusterExists(mongodbResource, &r, 4),
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
					testAccCheckMDBMongoDBClusterHasHostParameters(mongodbResource, map[string]interface{}{"tags": map[string]string{}, "hidden": false, "priority": 1.05, "secondary_delay_secs": int64(0)}, 1),
					testAccCheckMDBMongoDBClusterHasHostParameters(mongodbResource, map[string]interface{}{"tags": map[string]string{}, "hidden": false, "priority": 1.15, "secondary_delay_secs": int64(0)}, 1),
					testAccCheckMDBMongoDBClusterHasHostParameters(mongodbResource, map[string]interface{}{"tags": map[string]string{}, "hidden": false, "priority": 1.0, "secondary_delay_secs": int64(0)}, 1),
					testAccCheckMDBMongoDBClusterHasHostParameters(mongodbResource, map[string]interface{}{"tags": map[string]string{"abc": "def"}, "priority": 0.0, "hidden": true, "secondary_delay_secs": int64(0)}, 1),
				),
			},
			mdbMongoDBClusterImportStep(),
			// Check if description can be set to null
			{
				Config: makeConfig(t, &configData, &map[string]interface{}{
					"ClusterDescription": "",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMongoDBClusterExists(mongodbResource, &r, 4),
					resource.TestCheckResourceAttr(mongodbResource, "description", ""),
				),
			},
			mdbMongoDBClusterImportStep(),
			// Update disk size autoscaling
			{
				Config: makeConfig(t, &configData, &map[string]interface{}{
					"DiskSizeAutoscalingMongod": &mongodb.DiskSizeAutoscaling{
						DiskSizeLimit:           &wrapperspb.Int64Value{Value: (s2Small26hdd.DiskSize * 2) >> 30},
						PlannedUsageThreshold:   &wrapperspb.Int64Value{Value: 80},
						EmergencyUsageThreshold: &wrapperspb.Int64Value{Value: 90},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMongoDBClusterExists(mongodbResource, &r, 4),
					testAccCheckMDBMongoDBClusterHasMongodSpec(&r, map[string]interface{}{"DiskSizeAutoscalingMongod": &mongodb.DiskSizeAutoscaling{
						DiskSizeLimit:           &wrapperspb.Int64Value{Value: s2Small26hdd.DiskSize * 2},
						PlannedUsageThreshold:   &wrapperspb.Int64Value{Value: 80},
						EmergencyUsageThreshold: &wrapperspb.Int64Value{Value: 90},
					}}),
				),
			},
			mdbMongoDBClusterImportStep(),
			// web_sql enable
			{
				Config: makeConfig(t, &configData, &map[string]interface{}{
					"Access": map[string]bool{
						"web_sql":       true,
						"data_lens":     true,
						"data_transfer": true,
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMongoDBClusterExists(mongodbResource, &r, 4),
					resource.TestCheckResourceAttr(mongodbResource, "cluster_config.0.access.0.web_sql", "true"),
				),
			},
			mdbMongoDBClusterImportStep(),
		},
	})
}

// minimal configs for creation mongodb cluster
func create6_0V1ConfigData() map[string]interface{} {
	return map[string]interface{}{
		"Version":     "6.0",
		"ClusterName": acctest.RandomWithPrefix("test-acc-tf-mongodb"),
		"Environment": "PRESTABLE",
		"Mongod":      map[string]interface{}{},
		"Databases":   []string{"testdb"},
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
		"ResourcesMongod": &mongodb.Resources{
			ResourcePresetId: s2Small26hdd.ResourcePresetId,
			DiskSize:         s2Small26hdd.DiskSize >> 30,
			DiskTypeId:       s2Small26hdd.DiskTypeId,
		},
		"ResourcesMongoCfg": &mongodb.Resources{
			ResourcePresetId: s2Micro16hdd.ResourcePresetId,
			DiskSize:         toBytes(11) >> 30,
			DiskTypeId:       s2Micro16hdd.DiskTypeId,
		},
		"ResourcesMongos": &mongodb.Resources{
			ResourcePresetId: s2Micro16hdd.ResourcePresetId,
			DiskSize:         toBytes(12) >> 30,
			DiskTypeId:       s2Micro16hdd.DiskTypeId,
		},
		"ResourcesMongoInfra": &mongodb.Resources{
			ResourcePresetId: s2Micro16hdd.ResourcePresetId,
			DiskSize:         toBytes(13) >> 30,
			DiskTypeId:       s2Micro16hdd.DiskTypeId,
		},
		"DiskSizeAutoscalingMongod": &mongodb.DiskSizeAutoscaling{
			DiskSizeLimit:           &wrapperspb.Int64Value{Value: (s2Small26hdd.DiskSize * 2) >> 30},
			EmergencyUsageThreshold: &wrapperspb.Int64Value{Value: 90},
		},
		"SecurityGroupIds": []string{"${yandex_vpc_security_group.sg-x.id}"},
		"MaintenanceWindow": map[string]interface{}{
			"Type": "WEEKLY",
			"Day":  "FRI",
			"Hour": 20,
		},
	}
}

func create6_0V0ConfigData() map[string]interface{} {
	return map[string]interface{}{
		"Version":     "6.0",
		"ClusterName": acctest.RandomWithPrefix("test-acc-tf-mongodb"),
		"Environment": "PRESTABLE",
		"Mongod":      map[string]interface{}{},
		"Mongos":      map[string]interface{}{},
		"MongoCfg":    map[string]interface{}{},
		"Databases":   []string{"testdb"},
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
			ResourcePresetId: s2Small26hdd.ResourcePresetId,
			DiskSize:         s2Small26hdd.DiskSize >> 30,
			DiskTypeId:       s2Small26hdd.DiskTypeId,
		},
		"SecurityGroupIds": []string{"${yandex_vpc_security_group.sg-x.id}"},
		"MaintenanceWindow": map[string]interface{}{
			"Type": "WEEKLY",
			"Day":  "FRI",
			"Hour": 20,
		},
	}
}

// 3 tests for check backward compatibility and upgrade to new resources
func TestAccMDBMongoDBCluster_6_0NotShardedV0(t *testing.T) {
	t.Parallel()

	configData := create6_0V0ConfigData()
	configData["Hosts"] = mongoHosts
	clusterName := configData["ClusterName"].(string)
	version := configData["Version"].(string)

	var testCluster mongodb.Cluster
	folderID := getExampleFolderID()

	newResourcesV0 := mongodb.Resources{
		ResourcePresetId: s2Small26hdd.ResourcePresetId,
		DiskSize:         toBytes(28),
		DiskTypeId:       s2Small26hdd.DiskTypeId,
	}

	newResourcesV1 := mongodb.Resources{
		ResourcePresetId: s2Small26hdd.ResourcePresetId,
		DiskSize:         toBytes(28),
		DiskTypeId:       s2Small26hdd.DiskTypeId,
	}
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
						"Resources": &s2Small26hdd,
					}),
				),
			},
			{ // Update resources
				Config: makeConfig(t, &configData, &map[string]interface{}{
					"Resources": &mongodb.Resources{
						ResourcePresetId: newResourcesV0.ResourcePresetId,
						DiskSize:         newResourcesV0.DiskSize >> 30,
						DiskTypeId:       newResourcesV0.DiskTypeId,
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMongoDBClusterExists(mongodbResource, &testCluster, 2),
					testAccCheckMDBMongoDBClusterHasMongodSpec(&testCluster, map[string]interface{}{
						"Resources": &newResourcesV0,
					}),
				),
			},
			{ // Migrate resources v0 to v1
				Config: func() string {
					delete(configData, "Resources")
					return makeConfig(t, &configData, &map[string]interface{}{
						"ResourcesMongod": &mongodb.Resources{
							ResourcePresetId: newResourcesV1.ResourcePresetId,
							DiskSize:         newResourcesV1.DiskSize >> 30,
							DiskTypeId:       newResourcesV1.DiskTypeId,
						},
					})
				}(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMongoDBClusterExists(mongodbResource, &testCluster, 2),
					testAccCheckMDBMongoDBClusterHasMongodSpec(&testCluster, map[string]interface{}{
						"ResourcesMongod": &newResourcesV1,
					}),
				),
			},
			mdbMongoDBClusterImportStep(),
		},
	})
}

func TestAccMDBMongoDBCluster_6_0ShardedCfgV0(t *testing.T) {
	t.Parallel()

	configData := create6_0V0ConfigData()
	configData["Hosts"] = shardedMongoCfgHosts
	clusterName := configData["ClusterName"].(string)
	version := configData["Version"].(string)

	var testCluster mongodb.Cluster
	folderID := getExampleFolderID()

	newResources := mongodb.Resources{
		ResourcePresetId: s2Small26hdd.ResourcePresetId,
		DiskSize:         toBytes(27),
		DiskTypeId:       s2Small26hdd.DiskTypeId,
	}
	newResourcesMongodV1 := mongodb.Resources{
		ResourcePresetId: s2Small26hdd.ResourcePresetId,
		DiskSize:         toBytes(27),
		DiskTypeId:       s2Small26hdd.DiskTypeId,
	}
	newResourcesMongosV1 := mongodb.Resources{
		ResourcePresetId: s2Small26hdd.ResourcePresetId,
		DiskSize:         toBytes(29),
		DiskTypeId:       s2Small26hdd.DiskTypeId,
	}
	newResourcesMongoCfgV1 := mongodb.Resources{
		ResourcePresetId: s2Small26hdd.ResourcePresetId,
		DiskSize:         toBytes(30),
		DiskTypeId:       s2Small26hdd.DiskTypeId,
	}

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
					testAccCheckMDBMongoDBClusterExists(mongodbResource, &testCluster, 7),
					resource.TestCheckResourceAttr(mongodbResource, "name", clusterName),
					resource.TestCheckResourceAttr(mongodbResource, "folder_id", folderID),
					testAccCheckMDBMongoDBClusterHasRightVersion(&testCluster, version),
					testAccCheckMDBMongoDBClusterHasMongodSpec(&testCluster, map[string]interface{}{
						"Resources": &s2Small26hdd,
					}),
				),
			},
			{ // Update resources
				Config: makeConfig(t, &configData, &map[string]interface{}{
					"Resources": &mongodb.Resources{
						ResourcePresetId: newResources.ResourcePresetId,
						DiskSize:         newResources.DiskSize >> 30,
						DiskTypeId:       newResources.DiskTypeId,
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMongoDBClusterExists(mongodbResource, &testCluster, 7),
					testAccCheckMDBMongoDBClusterHasMongodSpec(&testCluster, map[string]interface{}{
						"ResourcesMongod":   &newResources,
						"ResourcesMongoCfg": &s2Small26hdd,
						"ResourcesMongos":   &s2Small26hdd,
					}),
				),
			},
			{ // Migrate to resources V1
				Config: func() string {
					delete(configData, "Resources")
					return makeConfig(t, &configData, &map[string]interface{}{
						"ResourcesMongod": &mongodb.Resources{
							ResourcePresetId: newResourcesMongodV1.ResourcePresetId,
							DiskSize:         newResourcesMongodV1.DiskSize >> 30,
							DiskTypeId:       newResourcesMongodV1.DiskTypeId,
						},
						"ResourcesMongos": &mongodb.Resources{
							ResourcePresetId: newResourcesMongosV1.ResourcePresetId,
							DiskSize:         newResourcesMongosV1.DiskSize >> 30,
							DiskTypeId:       newResourcesMongosV1.DiskTypeId,
						},
						"ResourcesMongoCfg": &mongodb.Resources{
							ResourcePresetId: newResourcesMongoCfgV1.ResourcePresetId,
							DiskSize:         newResourcesMongoCfgV1.DiskSize >> 30,
							DiskTypeId:       newResourcesMongoCfgV1.DiskTypeId,
						},
					})
				}(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMongoDBClusterExists(mongodbResource, &testCluster, 7),
					testAccCheckMDBMongoDBClusterHasMongodSpec(&testCluster, map[string]interface{}{
						"ResourcesMongod":   &newResourcesMongodV1,
						"ResourcesMongos":   &newResourcesMongosV1,
						"ResourcesMongoCfg": &newResourcesMongoCfgV1,
					}),
				),
			},
			// Update disk size autoscaling
			{
				Config: makeConfig(t, &configData, &map[string]interface{}{
					"DiskSizeAutoscalingMongod": &mongodb.DiskSizeAutoscaling{
						DiskSizeLimit:           &wrapperspb.Int64Value{Value: 31},
						PlannedUsageThreshold:   &wrapperspb.Int64Value{Value: 80},
						EmergencyUsageThreshold: &wrapperspb.Int64Value{Value: 90},
					},
					"DiskSizeAutoscalingMongos": &mongodb.DiskSizeAutoscaling{
						DiskSizeLimit:           &wrapperspb.Int64Value{Value: 32},
						PlannedUsageThreshold:   &wrapperspb.Int64Value{Value: 81},
						EmergencyUsageThreshold: &wrapperspb.Int64Value{Value: 91},
					},
					"DiskSizeAutoscalingMongoCfg": &mongodb.DiskSizeAutoscaling{
						DiskSizeLimit:           &wrapperspb.Int64Value{Value: 33},
						PlannedUsageThreshold:   &wrapperspb.Int64Value{Value: 82},
						EmergencyUsageThreshold: &wrapperspb.Int64Value{Value: 92},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMongoDBClusterExists(mongodbResource, &testCluster, 7),
					testAccCheckMDBMongoDBClusterHasMongodSpec(&testCluster, map[string]interface{}{
						"DiskSizeAutoscalingMongod": &mongodb.DiskSizeAutoscaling{
							DiskSizeLimit:           &wrapperspb.Int64Value{Value: toBytes(31)},
							PlannedUsageThreshold:   &wrapperspb.Int64Value{Value: 80},
							EmergencyUsageThreshold: &wrapperspb.Int64Value{Value: 90},
						},
						"DiskSizeAutoscalingMongos": &mongodb.DiskSizeAutoscaling{
							DiskSizeLimit:           &wrapperspb.Int64Value{Value: toBytes(32)},
							PlannedUsageThreshold:   &wrapperspb.Int64Value{Value: 81},
							EmergencyUsageThreshold: &wrapperspb.Int64Value{Value: 91},
						},
						"DiskSizeAutoscalingMongoCfg": &mongodb.DiskSizeAutoscaling{
							DiskSizeLimit:           &wrapperspb.Int64Value{Value: toBytes(33)},
							PlannedUsageThreshold:   &wrapperspb.Int64Value{Value: 82},
							EmergencyUsageThreshold: &wrapperspb.Int64Value{Value: 92},
						},
					}),
				),
			},
			mdbMongoDBClusterImportStep(),
		},
	})
}

func TestAccMDBMongoDBCluster_6_0ShardedInfraV0(t *testing.T) {
	t.Parallel()

	configData := create6_0V0ConfigData()
	configData["Hosts"] = shardedMongoInfraHosts
	clusterName := configData["ClusterName"].(string)
	version := configData["Version"].(string)

	var testCluster mongodb.Cluster
	folderID := getExampleFolderID()

	newResources := mongodb.Resources{
		ResourcePresetId: s2Small26hdd.ResourcePresetId,
		DiskSize:         toBytes(27),
		DiskTypeId:       s2Small26hdd.DiskTypeId,
	}

	newResourcesMongodV1 := mongodb.Resources{
		ResourcePresetId: s2Small26hdd.ResourcePresetId,
		DiskSize:         toBytes(29),
		DiskTypeId:       s2Small26hdd.DiskTypeId,
	}

	newResourcesMongoInfraV1 := mongodb.Resources{
		ResourcePresetId: s2Small26hdd.ResourcePresetId,
		DiskSize:         toBytes(28),
		DiskTypeId:       s2Small26hdd.DiskTypeId,
	}

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
					testAccCheckMDBMongoDBClusterExists(mongodbResource, &testCluster, 5),
					resource.TestCheckResourceAttr(mongodbResource, "name", clusterName),
					resource.TestCheckResourceAttr(mongodbResource, "folder_id", folderID),
					testAccCheckMDBMongoDBClusterHasRightVersion(&testCluster, version),
					testAccCheckMDBMongoDBClusterHasMongodSpec(&testCluster, map[string]interface{}{
						"Resources": &s2Small26hdd,
					}),
				),
			},
			{ // Update resources
				Config: makeConfig(t, &configData, &map[string]interface{}{
					"Resources": &mongodb.Resources{
						ResourcePresetId: newResources.ResourcePresetId,
						DiskSize:         newResources.DiskSize >> 30,
						DiskTypeId:       newResources.DiskTypeId,
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMongoDBClusterExists(mongodbResource, &testCluster, 5),
					testAccCheckMDBMongoDBClusterHasMongodSpec(&testCluster, map[string]interface{}{
						"ResourcesMongod":     &newResources,
						"ResourcesMongoInfra": &s2Small26hdd,
					}),
				),
			},
			{ // Migrate to resources V1
				Config: func() string {
					delete(configData, "Resources")
					return makeConfig(t, &configData, &map[string]interface{}{
						"ResourcesMongod": &mongodb.Resources{
							ResourcePresetId: newResourcesMongodV1.ResourcePresetId,
							DiskSize:         newResourcesMongodV1.DiskSize >> 30,
							DiskTypeId:       newResourcesMongodV1.DiskTypeId,
						},
						"ResourcesMongoInfra": &mongodb.Resources{
							ResourcePresetId: newResourcesMongoInfraV1.ResourcePresetId,
							DiskSize:         newResourcesMongoInfraV1.DiskSize >> 30,
							DiskTypeId:       newResourcesMongoInfraV1.DiskTypeId,
						},
					})
				}(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMongoDBClusterExists(mongodbResource, &testCluster, 5),
					testAccCheckMDBMongoDBClusterHasMongodSpec(&testCluster, map[string]interface{}{
						"ResourcesMongod":     &newResourcesMongodV1,
						"ResourcesMongoInfra": &newResourcesMongoInfraV1,
					}),
				),
			},
			mdbMongoDBClusterImportStep(),
		},
	})
}

func TestAccMDBMongoDBCluster_6_0NotShardedV1(t *testing.T) {
	t.Parallel()

	configData := create6_0V1ConfigData()
	delete(configData, "ResourcesMongos")
	delete(configData, "ResourcesMongoCfg")
	delete(configData, "ResourcesMongoInfra")
	configData["Hosts"] = mongoHosts
	clusterName := configData["ClusterName"].(string)
	version := configData["Version"].(string)

	var testCluster mongodb.Cluster
	folderID := getExampleFolderID()

	newResources := mongodb.Resources{
		ResourcePresetId: s2Small26hdd.ResourcePresetId,
		DiskSize:         toBytes(30),
		DiskTypeId:       s2Small26hdd.DiskTypeId,
	}
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
						"ResourcesMongod": &s2Small26hdd,
					}),
					testAccCheckMDBMongoDBClusterHasDatabases(mongodbResource, []string{"testdb"}),
					testAccCheckMDBMongoDBClusterHasUsers(mongodbResource, map[string][]string{"john": {"testdb"}}),
					testAccCheckCreatedAtAttr(mongodbResource),
					resource.TestCheckResourceAttr(mongodbResource, "security_group_ids.#", "1"),
					resource.TestCheckNoResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.net.0.max_incoming_connections"),
					resource.TestCheckNoResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.net.0.compressors"),
					resource.TestCheckNoResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.operation_profiling.0.mode"),
					resource.TestCheckNoResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.operation_profiling.0.slow_op_threshold"),
					resource.TestCheckNoResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.operation_profiling.0.slow_op_sample_rate"),
					resource.TestCheckNoResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.storage.0.wired_tiger.0.block_compressor"),
					resource.TestCheckNoResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.storage.0.wired_tiger.0.prefix_compression"),
					resource.TestCheckNoResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.storage.0.journal.0.commit_interval"),
					resource.TestCheckNoResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.set_parameter.0.enable_flow_control"),
					resource.TestCheckNoResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.set_parameter.0.min_snapshot_history_window_in_seconds"),
				),
			},
			mdbMongoDBClusterImportStep(),
			{ // Update resources
				Config: makeConfig(t, &configData, &map[string]interface{}{
					"ResourcesMongod": &mongodb.Resources{
						ResourcePresetId: newResources.ResourcePresetId,
						DiskSize:         newResources.DiskSize >> 30,
						DiskTypeId:       newResources.DiskTypeId,
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMongoDBClusterExists(mongodbResource, &testCluster, 2),
					testAccCheckMDBMongoDBClusterHasMongodSpec(&testCluster, map[string]interface{}{
						"ResourcesMongod": &newResources,
					}),
				),
			},
			mdbMongoDBClusterImportStep(),
			{ // Update resources
				Config: makeConfig(t, &configData, &map[string]interface{}{
					"Mongod": map[string]interface{}{
						"Net": map[string]interface{}{
							"MaxConnections": 1100,
							"Compressors":    []string{"\"ZLIB\""},
						},
						"SetParameter": map[string]interface{}{
							"EnableFlowControl":                 true,
							"MinSnapshotHistoryWindowInSeconds": 300,
						},
						"OperationProfiling": map[string]interface{}{
							"Mode":         "ALL",
							"OpThreshold":  1000,
							"OpSampleRate": 0.5,
						},
						"Storage": map[string]interface{}{
							"WiredTiger": map[string]interface{}{
								"Compressor":        "ZLIB",
								"PrefixCompression": false,
							},
							"Journal": map[string]interface{}{
								"CommitInterval": 404,
							},
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMongoDBClusterExists(mongodbResource, &testCluster, 2),
					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.net.0.max_incoming_connections", "1100"),
					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.net.0.compressors.0", "ZLIB"),
					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.net.0.compressors.#", "1"),
					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.operation_profiling.0.mode", "ALL"),
					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.operation_profiling.0.slow_op_threshold", "1000"),
					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.operation_profiling.0.slow_op_sample_rate", "0.5"),
					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.storage.0.wired_tiger.0.block_compressor", "ZLIB"),
					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.storage.0.wired_tiger.0.prefix_compression", "false"),
					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.storage.0.journal.0.commit_interval", "404"),
					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.set_parameter.0.enable_flow_control", "true"),
					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.set_parameter.0.min_snapshot_history_window_in_seconds", "300"),
				),
			},
			mdbMongoDBClusterImportStep(),
		},
	})
}
func TestAccMDBMongoDBCluster_6_0ShardedCfgV1(t *testing.T) {
	t.Parallel()

	configData := create6_0V1ConfigData()
	delete(configData, "ResourcesMongoInfra")
	configData["Hosts"] = shardedMongoCfgHosts
	clusterName := configData["ClusterName"].(string)
	version := configData["Version"].(string)

	var testCluster mongodb.Cluster
	folderID := getExampleFolderID()

	oldResourcesMongos := mongodb.Resources{
		ResourcePresetId: s2Micro16hdd.ResourcePresetId,
		DiskSize:         toBytes(12),
		DiskTypeId:       s2Micro16hdd.DiskTypeId,
	}
	oldResourcesMongoCfg := mongodb.Resources{
		ResourcePresetId: s2Micro16hdd.ResourcePresetId,
		DiskSize:         toBytes(11),
		DiskTypeId:       s2Micro16hdd.DiskTypeId,
	}

	newResourcesMongod := mongodb.Resources{
		ResourcePresetId: s2Small26hdd.ResourcePresetId,
		DiskSize:         toBytes(28),
		DiskTypeId:       s2Small26hdd.DiskTypeId,
	}
	newResourcesMongos := mongodb.Resources{
		ResourcePresetId: s2Small26hdd.ResourcePresetId,
		DiskSize:         toBytes(29),
		DiskTypeId:       s2Small26hdd.DiskTypeId,
	}
	newResourcesMongoCfg := mongodb.Resources{
		ResourcePresetId: s2Small26hdd.ResourcePresetId,
		DiskSize:         toBytes(30),
		DiskTypeId:       s2Small26hdd.DiskTypeId,
	}

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
					testAccCheckMDBMongoDBClusterExists(mongodbResource, &testCluster, 7),
					resource.TestCheckResourceAttr(mongodbResource, "name", clusterName),
					resource.TestCheckResourceAttr(mongodbResource, "folder_id", folderID),
					testAccCheckMDBMongoDBClusterHasRightVersion(&testCluster, version),
					testAccCheckMDBMongoDBClusterHasMongodSpec(&testCluster, map[string]interface{}{
						"ResourcesMongod":   &s2Small26hdd,
						"ResourcesMongos":   &oldResourcesMongos,
						"ResourcesMongoCfg": &oldResourcesMongoCfg,
					}),
					testAccCheckMDBMongoDBClusterHasDatabases(mongodbResource, []string{"testdb"}),
					testAccCheckMDBMongoDBClusterHasUsers(mongodbResource, map[string][]string{"john": {"testdb"}}),
					testAccCheckCreatedAtAttr(mongodbResource),
					resource.TestCheckResourceAttr(mongodbResource, "security_group_ids.#", "1"),
					resource.TestCheckNoResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.net.0.max_incoming_connections"),
					resource.TestCheckNoResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.net.0.compressors"),
					resource.TestCheckNoResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.operation_profiling.0.mode"),
					resource.TestCheckNoResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.operation_profiling.0.slow_op_threshold"),
					resource.TestCheckNoResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.operation_profiling.0.slow_op_sample_rate"),
					resource.TestCheckNoResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.storage.0.wired_tiger.0.block_compressor"),
					resource.TestCheckNoResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.storage.0.wired_tiger.0.prefix_compression"),
					resource.TestCheckNoResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.storage.0.journal.0.commit_interval"),
					resource.TestCheckNoResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.set_parameter.0.enable_flow_control"),
					resource.TestCheckNoResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.set_parameter.0.min_snapshot_history_window_in_seconds"),
					resource.TestCheckNoResourceAttr(mongodbResource,
						"cluster_config.0.mongos.0.net.0.max_incoming_connections"),
					resource.TestCheckNoResourceAttr(mongodbResource,
						"cluster_config.0.mongos.0.net.0.compressors"),
					resource.TestCheckNoResourceAttr(mongodbResource,
						"cluster_config.0.mongocfg.0.net.0.max_incoming_connections"),
					resource.TestCheckNoResourceAttr(mongodbResource,
						"cluster_config.0.mongocfg.0.operation_profiling.0.mode"),
					resource.TestCheckNoResourceAttr(mongodbResource,
						"cluster_config.0.mongocfg.0.operation_profiling.0.slow_op_threshold"),
				),
			},
			mdbMongoDBClusterImportStep(),
			{ // Update resources
				Config: makeConfig(t, &configData, &map[string]interface{}{
					"ResourcesMongod": &mongodb.Resources{
						ResourcePresetId: newResourcesMongod.ResourcePresetId,
						DiskSize:         newResourcesMongod.DiskSize >> 30,
						DiskTypeId:       newResourcesMongod.DiskTypeId,
					},
					"ResourcesMongos": &mongodb.Resources{
						ResourcePresetId: newResourcesMongos.ResourcePresetId,
						DiskSize:         newResourcesMongos.DiskSize >> 30,
						DiskTypeId:       newResourcesMongos.DiskTypeId,
					},
					"ResourcesMongoCfg": &mongodb.Resources{
						ResourcePresetId: newResourcesMongoCfg.ResourcePresetId,
						DiskSize:         newResourcesMongoCfg.DiskSize >> 30,
						DiskTypeId:       newResourcesMongoCfg.DiskTypeId,
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMongoDBClusterExists(mongodbResource, &testCluster, 7),
					testAccCheckMDBMongoDBClusterHasMongodSpec(&testCluster, map[string]interface{}{
						"ResourcesMongod":   &newResourcesMongod,
						"ResourcesMongos":   &newResourcesMongos,
						"ResourcesMongoCfg": &newResourcesMongoCfg,
					}),
				),
			},
			mdbMongoDBClusterImportStep(),
			{ // Update mongod, mongos, mongocfg configs
				Config: makeConfig(t, &configData, &map[string]interface{}{
					"Mongod": map[string]interface{}{
						"Net": map[string]interface{}{
							"MaxConnections": 1100,
							"Compressors":    []string{"\"ZLIB\""},
						},
						"SetParameter": map[string]interface{}{
							"EnableFlowControl":                 true,
							"MinSnapshotHistoryWindowInSeconds": 300,
						},
						"OperationProfiling": map[string]interface{}{
							"Mode":         "ALL",
							"OpThreshold":  1000,
							"OpSampleRate": 0.5,
						},
						"Storage": map[string]interface{}{
							"WiredTiger": map[string]interface{}{
								"Compressor":        "ZLIB",
								"PrefixCompression": false,
							},
							"Journal": map[string]interface{}{
								"CommitInterval": 404,
							},
						},
					},
					"Mongos": map[string]interface{}{
						"Net": map[string]interface{}{
							"MaxConnections": 1101,
							"Compressors":    []string{"\"ZLIB\""},
						},
					},
					"MongoCfg": map[string]interface{}{
						"OperationProfiling": map[string]interface{}{
							"Mode":        "SLOW_OP",
							"OpThreshold": 1024,
						},
						"Net": map[string]interface{}{
							"MaxConnections": 1102,
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMongoDBClusterExists(mongodbResource, &testCluster, 7),
					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.net.0.max_incoming_connections", "1100"),
					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.net.0.compressors.0", "ZLIB"),
					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.net.0.compressors.#", "1"),
					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.operation_profiling.0.mode", "ALL"),
					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.operation_profiling.0.slow_op_threshold", "1000"),
					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.operation_profiling.0.slow_op_sample_rate", "0.5"),
					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.storage.0.wired_tiger.0.block_compressor", "ZLIB"),
					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.storage.0.wired_tiger.0.prefix_compression", "false"),
					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.storage.0.journal.0.commit_interval", "404"),
					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.set_parameter.0.enable_flow_control", "true"),
					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.set_parameter.0.min_snapshot_history_window_in_seconds", "300"),

					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongos.0.net.0.max_incoming_connections", "1101"),
					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongos.0.net.0.compressors.0", "ZLIB"),
					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongos.0.net.0.compressors.#", "1"),

					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongocfg.0.net.0.max_incoming_connections", "1102"),
					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongocfg.0.operation_profiling.0.mode", "SLOW_OP"),
					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongocfg.0.operation_profiling.0.slow_op_threshold", "1024"),
				),
			},
			mdbMongoDBClusterImportStep(),
		},
	})
}
func TestAccMDBMongoDBCluster_6_0ShardedInfraV1(t *testing.T) {
	t.Parallel()

	configData := create6_0V1ConfigData()
	delete(configData, "ResourcesMongos")
	delete(configData, "ResourcesMongoCfg")
	configData["Hosts"] = shardedMongoInfraHosts
	clusterName := configData["ClusterName"].(string)
	version := configData["Version"].(string)

	var testCluster mongodb.Cluster
	folderID := getExampleFolderID()

	oldResourcesMongoInfra := mongodb.Resources{
		ResourcePresetId: s2Micro16hdd.ResourcePresetId,
		DiskSize:         toBytes(13),
		DiskTypeId:       s2Micro16hdd.DiskTypeId,
	}

	newResourcesMongod := mongodb.Resources{
		ResourcePresetId: s2Small26hdd.ResourcePresetId,
		DiskSize:         toBytes(29),
		DiskTypeId:       s2Small26hdd.DiskTypeId,
	}
	newResourcesMongoInfra := mongodb.Resources{
		ResourcePresetId: s2Small26hdd.ResourcePresetId,
		DiskSize:         toBytes(28),
		DiskTypeId:       s2Small26hdd.DiskTypeId,
	}

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
					testAccCheckMDBMongoDBClusterExists(mongodbResource, &testCluster, 5),
					resource.TestCheckResourceAttr(mongodbResource, "name", clusterName),
					resource.TestCheckResourceAttr(mongodbResource, "folder_id", folderID),
					testAccCheckMDBMongoDBClusterHasRightVersion(&testCluster, version),
					testAccCheckMDBMongoDBClusterHasMongodSpec(&testCluster, map[string]interface{}{
						"ResourcesMongod":     &s2Small26hdd,
						"ResourcesMongoInfra": &oldResourcesMongoInfra,
					}),
					testAccCheckMDBMongoDBClusterHasDatabases(mongodbResource, []string{"testdb"}),
					testAccCheckMDBMongoDBClusterHasUsers(mongodbResource, map[string][]string{"john": {"testdb"}}),
					testAccCheckCreatedAtAttr(mongodbResource),
					resource.TestCheckResourceAttr(mongodbResource, "security_group_ids.#", "1"),
					resource.TestCheckNoResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.net.0.max_incoming_connections"),
					resource.TestCheckNoResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.net.0.compressors"),
					resource.TestCheckNoResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.operation_profiling.0.mode"),
					resource.TestCheckNoResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.operation_profiling.0.slow_op_threshold"),
					resource.TestCheckNoResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.operation_profiling.0.slow_op_sample_rate"),
					resource.TestCheckNoResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.storage.0.wired_tiger.0.block_compressor"),
					resource.TestCheckNoResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.storage.0.wired_tiger.0.prefix_compression"),
					resource.TestCheckNoResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.storage.0.journal.0.commit_interval"),
					resource.TestCheckNoResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.set_parameter.0.enable_flow_control"),
					resource.TestCheckNoResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.set_parameter.0.min_snapshot_history_window_in_seconds"),
				),
			},
			{ // Update resources
				Config: makeConfig(t, &configData, &map[string]interface{}{
					"ResourcesMongod": &mongodb.Resources{
						ResourcePresetId: newResourcesMongod.ResourcePresetId,
						DiskSize:         newResourcesMongod.DiskSize >> 30,
						DiskTypeId:       newResourcesMongod.DiskTypeId,
					},
					"ResourcesMongoInfra": &mongodb.Resources{
						ResourcePresetId: newResourcesMongoInfra.ResourcePresetId,
						DiskSize:         newResourcesMongoInfra.DiskSize >> 30,
						DiskTypeId:       newResourcesMongoInfra.DiskTypeId,
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMongoDBClusterExists(mongodbResource, &testCluster, 5),
					testAccCheckMDBMongoDBClusterHasMongodSpec(&testCluster, map[string]interface{}{
						"ResourcesMongod":     &newResourcesMongod,
						"ResourcesMongoInfra": &newResourcesMongoInfra,
					}),
				),
			},
			mdbMongoDBClusterImportStep(),
			// todo: add test on mongos and mongocfg config after add functionality to public api
			{ // Update resources
				Config: makeConfig(t, &configData, &map[string]interface{}{
					"Mongod": map[string]interface{}{
						"Net": map[string]interface{}{
							"MaxConnections": 1100,
							"Compressors":    []string{"\"DISABLED\""},
						},
						"SetParameter": map[string]interface{}{
							"EnableFlowControl":                 true,
							"MinSnapshotHistoryWindowInSeconds": 300,
						},
						"OperationProfiling": map[string]interface{}{
							"Mode":         "ALL",
							"OpThreshold":  1000,
							"OpSampleRate": 0.5,
						},
						"Storage": map[string]interface{}{
							"WiredTiger": map[string]interface{}{
								"Compressor":        "ZLIB",
								"PrefixCompression": false,
							},
							"Journal": map[string]interface{}{
								"CommitInterval": 404,
							},
						},
					},
					"Mongos": map[string]interface{}{
						"Net": map[string]interface{}{
							"MaxConnections": 1101,
						},
					},
					"MongoCfg": map[string]interface{}{
						"OperationProfiling": map[string]interface{}{
							"OpThreshold": 1000,
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMongoDBClusterExists(mongodbResource, &testCluster, 5),
					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.net.0.max_incoming_connections", "1100"),
					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.net.0.compressors.0", "DISABLED"),
					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.net.0.compressors.#", "1"),
					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.operation_profiling.0.mode", "ALL"),
					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.operation_profiling.0.slow_op_threshold", "1000"),
					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.operation_profiling.0.slow_op_sample_rate", "0.5"),
					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.storage.0.wired_tiger.0.block_compressor", "ZLIB"),
					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.storage.0.wired_tiger.0.prefix_compression", "false"),
					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.storage.0.journal.0.commit_interval", "404"),
					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.set_parameter.0.enable_flow_control", "true"),
					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongod.0.set_parameter.0.min_snapshot_history_window_in_seconds", "300"),
					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongos.0.net.0.max_incoming_connections", "1101"),
					resource.TestCheckResourceAttr(mongodbResource,
						"cluster_config.0.mongocfg.0.operation_profiling.0.slow_op_threshold", "1000"),
				),
			},
			mdbMongoDBClusterImportStep(),
		},
	})
}

func TestAccMDBMongoDBCluster_enableSharding(t *testing.T) {
	t.Parallel()

	configData := create6_0V1ConfigData()
	delete(configData, "ResourcesMongos")
	delete(configData, "ResourcesMongoCfg")
	delete(configData, "ResourcesMongoInfra")
	configData["Hosts"] = mongoHosts
	clusterName := configData["ClusterName"].(string)
	version := configData["Version"].(string)

	resourcesMongoInfra := mongodb.Resources{
		ResourcePresetId: s2Micro16hdd.ResourcePresetId,
		DiskSize:         toBytes(13),
		DiskTypeId:       s2Micro16hdd.DiskTypeId,
	}

	var testCluster mongodb.Cluster
	folderID := getExampleFolderID()

	updateHosts := []mongodb.Host{
		{
			ZoneId:    "ru-central1-a",
			SubnetId:  "${yandex_vpc_subnet.foo.id}",
			Type:      mongodb.Host_MONGOD,
			ShardName: "rs02",
		},
		{
			ZoneId:   "ru-central1-a",
			SubnetId: "${yandex_vpc_subnet.foo.id}",
			Type:     mongodb.Host_MONGOINFRA,
		},
		{
			ZoneId:   "ru-central1-b",
			SubnetId: "${yandex_vpc_subnet.bar.id}",
			Type:     mongodb.Host_MONGOINFRA,
		},
		{
			ZoneId:   "ru-central1-a",
			SubnetId: "${yandex_vpc_subnet.foo.id}",
			Type:     mongodb.Host_MONGOINFRA,
		},
	}

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
				Config: makeConfig(t, &configData, &map[string]interface{}{
					"ClusterDescription": "enableSharding",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMongoDBClusterExists(mongodbResource, &testCluster, 2),
					resource.TestCheckResourceAttr(mongodbResource, "name", clusterName),
					resource.TestCheckResourceAttr(mongodbResource, "folder_id", folderID),
					testAccCheckMDBMongoDBClusterHasRightVersion(&testCluster, version),
					testAccCheckMDBMongoDBClusterHasMongodSpec(&testCluster, map[string]interface{}{
						"ResourcesMongod": &s2Small26hdd,
					}),
					resource.TestCheckResourceAttr(mongodbResource, "sharded", "false"),
				),
			},
			mdbMongoDBClusterImportStep(),
			// EnableSharding
			{
				Config: makeConfig(t, &configData, &map[string]interface{}{
					"ResourcesMongoInfra": &mongodb.Resources{
						ResourcePresetId: resourcesMongoInfra.ResourcePresetId,
						DiskSize:         resourcesMongoInfra.DiskSize >> 30,
						DiskTypeId:       resourcesMongoInfra.DiskTypeId,
					},
					"Hosts": shardedMongoInfraHosts,
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMongoDBClusterExists(mongodbResource, &testCluster, 5),
					resource.TestCheckResourceAttr(mongodbResource, "name", clusterName),
					resource.TestCheckResourceAttr(mongodbResource, "folder_id", folderID),
					testAccCheckMDBMongoDBClusterHasRightVersion(&testCluster, version),
					testAccCheckMDBMongoDBClusterHasMongodSpec(&testCluster, map[string]interface{}{
						"ResourcesMongod": &s2Small26hdd,
						"ResourcesMongoInfra": &mongodb.Resources{
							ResourcePresetId: resourcesMongoInfra.ResourcePresetId,
							DiskSize:         resourcesMongoInfra.DiskSize,
							DiskTypeId:       resourcesMongoInfra.DiskTypeId,
						},
					}),
					testAccCheckMDBMongoDBClusterHasShards(mongodbResource, []string{"rs01"}),
					resource.TestCheckResourceAttr(mongodbResource, "sharded", "true"),
				),
			},
			mdbMongoDBClusterImportStep(),
			// delete and add shard
			{
				Config: makeConfig(t, &configData, &map[string]interface{}{
					"ResourcesMongoInfra": &mongodb.Resources{
						ResourcePresetId: resourcesMongoInfra.ResourcePresetId,
						DiskSize:         resourcesMongoInfra.DiskSize >> 30,
						DiskTypeId:       resourcesMongoInfra.DiskTypeId,
					},
					"Hosts": updateHosts,
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMongoDBClusterExists(mongodbResource, &testCluster, 4),
					resource.TestCheckResourceAttr(mongodbResource, "name", clusterName),
					resource.TestCheckResourceAttr(mongodbResource, "folder_id", folderID),
					testAccCheckMDBMongoDBClusterHasRightVersion(&testCluster, version),
					testAccCheckMDBMongoDBClusterHasMongodSpec(&testCluster, map[string]interface{}{
						"ResourcesMongod": &s2Small26hdd,
						"ResourcesMongoInfra": &mongodb.Resources{
							ResourcePresetId: resourcesMongoInfra.ResourcePresetId,
							DiskSize:         resourcesMongoInfra.DiskSize,
							DiskTypeId:       resourcesMongoInfra.DiskTypeId,
						},
					}),
					testAccCheckMDBMongoDBClusterHasShards(mongodbResource, []string{"rs02"}),
					resource.TestCheckResourceAttr(mongodbResource, "sharded", "true"),
				),
			},
			mdbMongoDBClusterImportStep(),
		},
	})
}

func TestAccMDBMongoDBCluster_restore(t *testing.T) {
	t.Parallel()

	configData := createRestoreConfigData()
	clusterName := configData["ClusterName"].(string)

	var r mongodb.Cluster
	folderID := getExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckMDBMongoDBClusterDestroy,
			testAccCheckVPCNetworkDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: makeConfig(t, &configData, nil),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMongoDBClusterExists(mongodbResource, &r, 2),
					resource.TestCheckResourceAttr(mongodbResource, "name", clusterName),
					resource.TestCheckResourceAttr(mongodbResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(mongodbResource, "cluster_config.0.access.0.data_lens", "true"),
					resource.TestCheckResourceAttr(mongodbResource, "cluster_config.0.access.0.data_transfer", "true"),
					resource.TestCheckResourceAttr(mongodbResource, "cluster_config.0.performance_diagnostics.0.enabled", "true"),
					testAccCheckMDBMongoDBClusterHasRightVersion(&r, configData["Version"].(string)),
					testAccCheckMDBMongoDBClusterHasMongodSpec(&r, map[string]interface{}{"Resources": &s2Micro16hdd}),
					testAccCheckMDBMongoDBClusterHasDatabases(mongodbResource, []string{"db1"}),
					testAccCheckMDBMongoDBClusterHasUsers(mongodbResource, map[string][]string{"user1": {"db1", "admin"}}),
					testAccCheckMDBMongoDBClusterContainsLabel(&r, "test_key", "test_value"),
					testAccCheckCreatedAtAttr(mongodbResource),
					resource.TestCheckResourceAttr(mongodbResource, "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(mongodbResource, "maintenance_window.0.type", "ANYTIME"),
					resource.TestCheckResourceAttr(mongodbResource, "deletion_protection", "true"),
				),
			},
			{
				Config: makeConfig(t, &configData, &map[string]interface{}{
					"DeletionProtection": "false",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMongoDBClusterExists(mongodbResource, &r, 2),
					resource.TestCheckResourceAttr(mongodbResource, "deletion_protection", "false"),
				),
			},
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
	if expected == nil {
		return nil
	}
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

func supportTestDiskSizeAutoscaling(actual *mongodb.DiskSizeAutoscaling, expected *mongodb.DiskSizeAutoscaling) error {
	if expected == nil {
		return nil
	}
	if actual.DiskSizeLimit.GetValue() != expected.DiskSizeLimit.GetValue() {
		return fmt.Errorf("Expected disk size limit '%+v', got '%+v'",
			expected.DiskSizeLimit, actual.DiskSizeLimit)
	}
	if expected.PlannedUsageThreshold != nil && actual.PlannedUsageThreshold == nil {
		return fmt.Errorf("Expected non nil planned usage threshold")
	}
	if expected.PlannedUsageThreshold != nil && actual.PlannedUsageThreshold.GetValue() != expected.PlannedUsageThreshold.GetValue() {
		return fmt.Errorf("Expected planned usage threshold '%+v', got '%+v'", expected.PlannedUsageThreshold, actual.PlannedUsageThreshold)
	}
	if expected.EmergencyUsageThreshold != nil && actual.EmergencyUsageThreshold == nil {
		return fmt.Errorf("Expected non nil emergency usage threshold")
	}
	if expected.EmergencyUsageThreshold != nil && actual.EmergencyUsageThreshold.GetValue() != expected.EmergencyUsageThreshold.GetValue() {
		return fmt.Errorf("Expected emergency usage threshold '%+v', got '%+v'", expected.EmergencyUsageThreshold, actual.EmergencyUsageThreshold)
	}
	return nil
}

func testAccCheckMDBMongoDBClusterHasHostParameters(n string, expected map[string]interface{}, count int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		resp, err := config.sdk.MDB().MongoDB().Cluster().ListHosts(context.Background(), &mongodb.ListClusterHostsRequest{
			ClusterId: rs.Primary.ID,
			PageSize:  defaultMDBPageSize,
		})
		if err != nil {
			return err
		}
		hosts := resp.Hosts

		matchCount := 0

		for _, host := range hosts {
			if expected == nil && host.HostParameters == nil {
				return nil
			}

			if host.HostParameters == nil {
				continue
			}
			if host.HostParameters.Hidden != expected["hidden"].(bool) {
				continue
			}
			if host.HostParameters.Priority != expected["priority"].(float64) {
				continue
			}
			if host.HostParameters.SecondaryDelaySecs != expected["secondary_delay_secs"].(int64) {
				continue
			}
			if maps.Equal(host.HostParameters.Tags, expected["tags"].(map[string]string)) {
				matchCount += 1
			}
		}
		if matchCount != count {
			return fmt.Errorf("found %v matching host parameters, expected %v", matchCount, count)
		}
		return nil
	}
}

func testAccCheckMDBMongoDBClusterHasMongodSpec(r *mongodb.Cluster, expected map[string]interface{}) resource.TestCheckFunc {
	var expectedResourcesMongod, expectedResourcesMongocfg, expectedResourcesMongos, expectedResourcesMongoinfra *mongodb.Resources
	if expectedResources, ok := expected["Resources"]; ok {
		expectedResourcesMongod = expectedResources.(*mongodb.Resources)
		expectedResourcesMongocfg = expectedResources.(*mongodb.Resources)
		expectedResourcesMongos = expectedResources.(*mongodb.Resources)
		expectedResourcesMongoinfra = expectedResources.(*mongodb.Resources)
	} else {
		if expectedResources, ok := expected["ResourcesMongod"]; ok {
			expectedResourcesMongod = expectedResources.(*mongodb.Resources)
		}
		if expectedResources, ok := expected["ResourcesMongos"]; ok {
			expectedResourcesMongos = expectedResources.(*mongodb.Resources)
		}
		if expectedResources, ok := expected["ResourcesMongoCfg"]; ok {
			expectedResourcesMongocfg = expectedResources.(*mongodb.Resources)
		}
		if expectedResources, ok := expected["ResourcesMongoInfra"]; ok {
			expectedResourcesMongoinfra = expectedResources.(*mongodb.Resources)
		}
	}

	var expectedDsaMongod, expectedDsaMongocfg, expectedDsaMongos, expectedDsaMongoinfra *mongodb.DiskSizeAutoscaling
	if expectedResources, ok := expected["DiskSizeAutoscalingMongod"]; ok {
		expectedDsaMongod = expectedResources.(*mongodb.DiskSizeAutoscaling)
	}
	if expectedResources, ok := expected["DiskSizeAutoscalingMongos"]; ok {
		expectedDsaMongos = expectedResources.(*mongodb.DiskSizeAutoscaling)
	}
	if expectedResources, ok := expected["DiskSizeAutoscalingMongoCfg"]; ok {
		expectedDsaMongocfg = expectedResources.(*mongodb.DiskSizeAutoscaling)
	}
	if expectedResources, ok := expected["DiskSizeAutoscalingMongoInfra"]; ok {
		expectedDsaMongoinfra = expectedResources.(*mongodb.DiskSizeAutoscaling)
	}

	return func(s *terraform.State) error {
		switch r.Config.Version {
		case "6.0-enterprise":
			{
				mongo := r.Config.Mongodb.(*mongodb.ClusterConfig_Mongodb_6_0Enterprise).Mongodb_6_0Enterprise
				d := mongo.Mongod
				if d != nil {
					err := supportTestResources(d.Resources, expectedResourcesMongod)
					if err != nil {
						return err
					}
					err = supportTestDiskSizeAutoscaling(d.DiskSizeAutoscaling, expectedDsaMongod)
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
					err := supportTestResources(s.Resources, expectedResourcesMongos)
					if err != nil {
						return err
					}
					err = supportTestDiskSizeAutoscaling(s.DiskSizeAutoscaling, expectedDsaMongos)
					if err != nil {
						return err
					}
				}

				cfg := mongo.Mongocfg
				if cfg != nil {
					err := supportTestResources(cfg.Resources, expectedResourcesMongocfg)
					if err != nil {
						return err
					}
					err = supportTestDiskSizeAutoscaling(cfg.DiskSizeAutoscaling, expectedDsaMongocfg)
					if err != nil {
						return err
					}
				}

				infra := mongo.Mongoinfra
				if infra != nil {
					err := supportTestResources(infra.Resources, expectedResourcesMongoinfra)
					if err != nil {
						return err
					}
					err = supportTestDiskSizeAutoscaling(infra.DiskSizeAutoscaling, expectedDsaMongoinfra)
					if err != nil {
						return err
					}
				}
			}
		case "5.0-enterprise":
			{
				mongo := r.Config.Mongodb.(*mongodb.ClusterConfig_Mongodb_5_0Enterprise).Mongodb_5_0Enterprise
				d := mongo.Mongod
				if d != nil {
					err := supportTestResources(d.Resources, expectedResourcesMongod)
					if err != nil {
						return err
					}
					err = supportTestDiskSizeAutoscaling(d.DiskSizeAutoscaling, expectedDsaMongod)
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
					err := supportTestResources(s.Resources, expectedResourcesMongos)
					if err != nil {
						return err
					}
					err = supportTestDiskSizeAutoscaling(s.DiskSizeAutoscaling, expectedDsaMongos)
					if err != nil {
						return err
					}
				}

				cfg := mongo.Mongocfg
				if cfg != nil {
					err := supportTestResources(cfg.Resources, expectedResourcesMongocfg)
					if err != nil {
						return err
					}
					err = supportTestDiskSizeAutoscaling(cfg.DiskSizeAutoscaling, expectedDsaMongocfg)
					if err != nil {
						return err
					}
				}

				infra := mongo.Mongoinfra
				if infra != nil {
					err := supportTestResources(infra.Resources, expectedResourcesMongoinfra)
					if err != nil {
						return err
					}
					err = supportTestDiskSizeAutoscaling(infra.DiskSizeAutoscaling, expectedDsaMongoinfra)
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
					err := supportTestResources(d.Resources, expectedResourcesMongod)
					if err != nil {
						return err
					}
					err = supportTestDiskSizeAutoscaling(d.DiskSizeAutoscaling, expectedDsaMongod)
					if err != nil {
						return err
					}
				}

				s := mongo.Mongos
				if s != nil {
					err := supportTestResources(s.Resources, expectedResourcesMongos)
					if err != nil {
						return err
					}
					err = supportTestDiskSizeAutoscaling(s.DiskSizeAutoscaling, expectedDsaMongos)
					if err != nil {
						return err
					}
				}

				cfg := mongo.Mongocfg
				if cfg != nil {
					err := supportTestResources(cfg.Resources, expectedResourcesMongocfg)
					if err != nil {
						return err
					}
					err = supportTestDiskSizeAutoscaling(cfg.DiskSizeAutoscaling, expectedDsaMongocfg)
					if err != nil {
						return err
					}
				}

				infra := mongo.Mongoinfra
				if infra != nil {
					err := supportTestResources(infra.Resources, expectedResourcesMongoinfra)
					if err != nil {
						return err
					}
					err = supportTestDiskSizeAutoscaling(infra.DiskSizeAutoscaling, expectedDsaMongoinfra)
					if err != nil {
						return err
					}
				}
			}
		case "6.0":
			{
				mongo := r.Config.Mongodb.(*mongodb.ClusterConfig_Mongodb_6_0).Mongodb_6_0
				d := mongo.Mongod
				if d != nil {
					err := supportTestResources(d.Resources, expectedResourcesMongod)
					if err != nil {
						return err
					}
					err = supportTestDiskSizeAutoscaling(d.DiskSizeAutoscaling, expectedDsaMongod)
					if err != nil {
						return err
					}
				}

				s := mongo.Mongos
				if s != nil {
					err := supportTestResources(s.Resources, expectedResourcesMongos)
					if err != nil {
						return err
					}
					err = supportTestDiskSizeAutoscaling(s.DiskSizeAutoscaling, expectedDsaMongos)
					if err != nil {
						return err
					}
				}

				cfg := mongo.Mongocfg
				if cfg != nil {
					err := supportTestResources(cfg.Resources, expectedResourcesMongocfg)
					if err != nil {
						return err
					}
					err = supportTestDiskSizeAutoscaling(cfg.DiskSizeAutoscaling, expectedDsaMongocfg)
					if err != nil {
						return err
					}
				}

				infra := mongo.Mongoinfra
				if infra != nil {
					err := supportTestResources(infra.Resources, expectedResourcesMongoinfra)
					if err != nil {
						return err
					}
					err = supportTestDiskSizeAutoscaling(infra.DiskSizeAutoscaling, expectedDsaMongoinfra)
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
					err := supportTestResources(d.Resources, expectedResourcesMongod)
					if err != nil {
						return err
					}
					err = supportTestDiskSizeAutoscaling(d.DiskSizeAutoscaling, expectedDsaMongod)
					if err != nil {
						return err
					}
				}

				s := mongo.Mongos
				if s != nil {
					err := supportTestResources(s.Resources, expectedResourcesMongos)
					if err != nil {
						return err
					}
					err = supportTestDiskSizeAutoscaling(s.DiskSizeAutoscaling, expectedDsaMongos)
					if err != nil {
						return err
					}
				}

				cfg := mongo.Mongocfg
				if cfg != nil {
					err := supportTestResources(cfg.Resources, expectedResourcesMongocfg)
					if err != nil {
						return err
					}
					err = supportTestDiskSizeAutoscaling(cfg.DiskSizeAutoscaling, expectedDsaMongocfg)
					if err != nil {
						return err
					}
				}

				infra := mongo.Mongoinfra
				if infra != nil {
					err := supportTestResources(infra.Resources, expectedResourcesMongoinfra)
					if err != nil {
						return err
					}
					err = supportTestDiskSizeAutoscaling(infra.DiskSizeAutoscaling, expectedDsaMongoinfra)
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
					err := supportTestResources(d.Resources, expectedResourcesMongod)
					if err != nil {
						return err
					}
					err = supportTestDiskSizeAutoscaling(d.DiskSizeAutoscaling, expectedDsaMongod)
					if err != nil {
						return err
					}
				}

				s := mongo.Mongos
				if s != nil {
					err := supportTestResources(s.Resources, expectedResourcesMongos)
					if err != nil {
						return err
					}
					err = supportTestDiskSizeAutoscaling(s.DiskSizeAutoscaling, expectedDsaMongos)
					if err != nil {
						return err
					}
				}

				cfg := mongo.Mongocfg
				if cfg != nil {
					err := supportTestResources(cfg.Resources, expectedResourcesMongocfg)
					if err != nil {
						return err
					}
					err = supportTestDiskSizeAutoscaling(cfg.DiskSizeAutoscaling, expectedDsaMongocfg)
					if err != nil {
						return err
					}
				}

				infra := mongo.Mongoinfra
				if infra != nil {
					err := supportTestResources(infra.Resources, expectedResourcesMongoinfra)
					if err != nil {
						return err
					}
					err = supportTestDiskSizeAutoscaling(infra.DiskSizeAutoscaling, expectedDsaMongoinfra)
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
					err := supportTestResources(d.Resources, expectedResourcesMongod)
					if err != nil {
						return err
					}
					err = supportTestDiskSizeAutoscaling(d.DiskSizeAutoscaling, expectedDsaMongod)
					if err != nil {
						return err
					}
				}

				s := mongo.Mongos
				if s != nil {
					err := supportTestResources(s.Resources, expectedResourcesMongos)
					if err != nil {
						return err
					}
					err = supportTestDiskSizeAutoscaling(s.DiskSizeAutoscaling, expectedDsaMongos)
					if err != nil {
						return err
					}
				}

				cfg := mongo.Mongocfg
				if cfg != nil {
					err := supportTestResources(cfg.Resources, expectedResourcesMongocfg)
					if err != nil {
						return err
					}
					err = supportTestDiskSizeAutoscaling(cfg.DiskSizeAutoscaling, expectedDsaMongocfg)
					if err != nil {
						return err
					}
				}

				infra := mongo.Mongoinfra
				if infra != nil {
					err := supportTestResources(infra.Resources, expectedResourcesMongoinfra)
					if err != nil {
						return err
					}
					err = supportTestDiskSizeAutoscaling(infra.DiskSizeAutoscaling, expectedDsaMongoinfra)
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
					err := supportTestResources(d.Resources, expectedResourcesMongod)
					if err != nil {
						return err
					}
					err = supportTestDiskSizeAutoscaling(d.DiskSizeAutoscaling, expectedDsaMongod)
					if err != nil {
						return err
					}
				}

				s := mongo.Mongos
				if s != nil {
					err := supportTestResources(s.Resources, expectedResourcesMongos)
					if err != nil {
						return err
					}
					err = supportTestDiskSizeAutoscaling(s.DiskSizeAutoscaling, expectedDsaMongos)
					if err != nil {
						return err
					}
				}

				cfg := mongo.Mongocfg
				if cfg != nil {
					err := supportTestResources(cfg.Resources, expectedResourcesMongocfg)
					if err != nil {
						return err
					}
					err = supportTestDiskSizeAutoscaling(cfg.DiskSizeAutoscaling, expectedDsaMongocfg)
					if err != nil {
						return err
					}
				}

				infra := mongo.Mongoinfra
				if infra != nil {
					err := supportTestResources(infra.Resources, expectedResourcesMongoinfra)
					if err != nil {
						return err
					}
					err = supportTestDiskSizeAutoscaling(infra.DiskSizeAutoscaling, expectedDsaMongoinfra)
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
					err := supportTestResources(d.Resources, expectedResourcesMongod)
					if err != nil {
						return err
					}
					err = supportTestDiskSizeAutoscaling(d.DiskSizeAutoscaling, expectedDsaMongod)
					if err != nil {
						return err
					}
				}

				s := mongo.Mongos
				if s != nil {
					err := supportTestResources(s.Resources, expectedResourcesMongos)
					if err != nil {
						return err
					}
					err = supportTestDiskSizeAutoscaling(s.DiskSizeAutoscaling, expectedDsaMongos)
					if err != nil {
						return err
					}
				}

				cfg := mongo.Mongocfg
				if cfg != nil {
					err := supportTestResources(cfg.Resources, expectedResourcesMongocfg)
					if err != nil {
						return err
					}
					err = supportTestDiskSizeAutoscaling(cfg.DiskSizeAutoscaling, expectedDsaMongocfg)
					if err != nil {
						return err
					}
				}

				infra := mongo.Mongoinfra
				if infra != nil {
					err := supportTestResources(infra.Resources, expectedResourcesMongoinfra)
					if err != nil {
						return err
					}
					err = supportTestDiskSizeAutoscaling(infra.DiskSizeAutoscaling, expectedDsaMongoinfra)
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

func testAccCheckMDBMongoDBClusterHasShards(r string, shards []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("Not found: %s", r)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		resp, err := config.sdk.MDB().MongoDB().Cluster().ListShards(context.Background(), &mongodb.ListClusterShardsRequest{
			ClusterId: rs.Primary.ID,
			PageSize:  defaultMDBPageSize,
		})
		if err != nil {
			return err
		}
		var shrds []string
		for _, d := range resp.Shards {
			shrds = append(shrds, d.Name)
		}

		if len(shrds) != len(shards) {
			return fmt.Errorf("Expected %d shards, found %d", len(shards), len(shrds))
		}

		sort.Strings(shrds)
		sort.Strings(shards)
		if fmt.Sprintf("%v", shrds) != fmt.Sprintf("%v", shards) {
			return fmt.Errorf("Cluster has wrong shards, %v. Expected %v", shrds, shards)
		}

		return nil
	}
}
