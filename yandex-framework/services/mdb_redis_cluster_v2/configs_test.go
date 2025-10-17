package mdb_redis_cluster_v2_test

import (
	"bytes"
	"reflect"
	"testing"
	"text/template"
)

func merge(a, b interface{}) {
	ra := reflect.ValueOf(a).Elem()
	rb := reflect.ValueOf(b).Elem()

	numFields := ra.NumField()

	for i := 0; i < numFields; i++ {
		field_a := ra.Field(i)
		field_b := rb.Field(i)

		switch field_a.Kind() {
		case reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
			if !field_b.IsNil() {
				field_a.Set(field_b)
			}
		}
	}
}

func makeConfig(t *testing.T, data *redisConfigTest, patch ...*redisConfigTest) string {
	for _, p := range patch {
		if p != nil {
			merge(data, p)
		}
	}

	tmpl, err := template.New("redis_config").Parse(templateRedis)
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
	return buf.String()
}

func testAccBaseConfig(name, description string) *redisConfigTest {
	baseDiskSize := 16
	diskTypeId := "network-ssd"
	baseFlavor := "hm3-c2-m8"
	return &redisConfigTest{
		Name:        newPtr(name),
		Description: newPtr(description),
		Environment: newPtr("PRESTABLE"),
		Resources: &hostResource{
			ResourcePresetId: newPtr(baseFlavor),
			DiskSize:         newPtr(baseDiskSize),
			DiskTypeId:       newPtr(diskTypeId),
		},
	}
}

func testAccAllSettingsConfig(name, description, version string, baseDiskSize int, diskTypeId, baseFlavor string, hosts map[string]host) *redisConfigTest {
	return &redisConfigTest{
		Name:               newPtr(name),
		Environment:        newPtr("PRESTABLE"),
		Description:        newPtr(description),
		TlsEnabled:         newPtr(true),
		PersistenceMode:    newPtr("OFF"),
		AnnounceHostnames:  newPtr(true),
		DeletionProtection: newPtr(true),
		AuthSentinel:       newPtr(true),
		Resources: &hostResource{
			ResourcePresetId: newPtr(baseFlavor),
			DiskSize:         newPtr(baseDiskSize),
			DiskTypeId:       newPtr(diskTypeId),
		},
		Labels:           map[string]string{"foo": "bar", "foo2": "bar2"},
		SecurityGroupIds: []string{"${yandex_vpc_security_group.sg-x.id}"},
		Access: &access{
			DataLens: newPtr(true),
			WebSql:   newPtr(true),
		},
		Hosts: hosts,
		DiskSizeAutoscaling: &diskSizeAutoscaling{
			DiskSizeLimit:           newPtr(baseDiskSize * 2),
			EmergencyUsageThreshold: newPtr(83),
		},
		Modules: &valkeyModules{
			ValkeySearch: &valkeySearch{
				Enabled: newPtr(false),
			},
			ValkeyJson: &valkeyJson{
				Enabled: newPtr(true),
			},
			ValkeyBloom: &valkeyBloom{
				Enabled: newPtr(false),
			},
		},
		MaintenanceWindow: &maintenanceWindow{
			Type: newPtr("WEEKLY"),
			Hour: newPtr(1),
			Day:  newPtr("MON"),
		},
		Config: &config{
			Password:                        newPtr("12345678P"),
			Timeout:                         newPtr(100),
			MaxmemoryPolicy:                 newPtr("ALLKEYS_LRU"),
			NotifyKeyspaceEvents:            newPtr("Elg"),
			SlowlogLogSlowerThan:            newPtr(5000),
			SlowlogMaxLen:                   newPtr(19),
			Databases:                       newPtr(18),
			MaxmemoryPercent:                newPtr(70),
			ClientOutputBufferLimitNormal:   newPtr("16777215 8388607 61"),
			ClientOutputBufferLimitPubsub:   newPtr("16777214 8388606 62"),
			UseLuajit:                       newPtr(true),
			IoThreadsAllowed:                newPtr(true),
			Version:                         &version,
			LuaTimeLimit:                    newPtr(4444),
			ReplBacklogSizePercent:          newPtr(15),
			ClusterRequireFullCoverage:      newPtr(true),
			ClusterAllowReadsWhenDown:       newPtr(true),
			ClusterAllowPubsubshardWhenDown: newPtr(true),
			LfuDecayTime:                    newPtr(14),
			LfuLogFactor:                    newPtr(13),
			TurnBeforeSwitchover:            newPtr(true),
			AllowDataLoss:                   newPtr(true),
			BackupRetainPeriodDays:          newPtr(12),
			ZsetMaxListpackEntries:          newPtr(256),
			BackupWindowStart: &backupWindowStart{
				Hours:   newPtr(10),
				Minutes: newPtr(11),
			},
		},
	}
}

func testAccAllSettingsConfigChanged(name, description, version string, baseDiskSize int, diskTypeId, baseFlavor string, hosts map[string]host) *redisConfigTest {
	return &redisConfigTest{
		Name:               newPtr(name),
		Environment:        newPtr("PRESTABLE"),
		Description:        newPtr(description),
		TlsEnabled:         newPtr(true),
		PersistenceMode:    newPtr("ON"),
		AnnounceHostnames:  newPtr(false),
		DeletionProtection: newPtr(false),
		AuthSentinel:       newPtr(false),
		Resources: &hostResource{
			ResourcePresetId: newPtr(baseFlavor),
			DiskSize:         newPtr(baseDiskSize),
			DiskTypeId:       newPtr(diskTypeId),
		},
		Labels:           map[string]string{"qwe": "rty", "foo2": "bar2"},
		SecurityGroupIds: []string{"${yandex_vpc_security_group.sg-x.id}", "${yandex_vpc_security_group.sg-y.id}"},
		Access: &access{
			DataLens: newPtr(false),
			WebSql:   newPtr(false),
		},
		Hosts: hosts,
		DiskSizeAutoscaling: &diskSizeAutoscaling{
			DiskSizeLimit:           newPtr(baseDiskSize * 3),
			EmergencyUsageThreshold: newPtr(84),
		},
		Modules: &valkeyModules{
			ValkeySearch: &valkeySearch{
				Enabled:       newPtr(true),
				ReaderThreads: newPtr(3),
				WriterThreads: newPtr(3),
			},
			ValkeyJson: &valkeyJson{
				Enabled: newPtr(true),
			},
			ValkeyBloom: &valkeyBloom{
				Enabled: newPtr(true),
			},
		},
		MaintenanceWindow: &maintenanceWindow{
			Type: newPtr("WEEKLY"),
			Hour: newPtr(2),
			Day:  newPtr("FRI"),
		},
		Config: &config{
			Password:                        newPtr("12345678PQ"),
			Timeout:                         newPtr(101),
			MaxmemoryPolicy:                 newPtr("VOLATILE_LFU"),
			NotifyKeyspaceEvents:            newPtr("Ex"),
			SlowlogLogSlowerThan:            newPtr(5001),
			SlowlogMaxLen:                   newPtr(20),
			Databases:                       newPtr(21),
			MaxmemoryPercent:                newPtr(71),
			ClientOutputBufferLimitNormal:   newPtr("16777212 8388605 63"),
			ClientOutputBufferLimitPubsub:   newPtr("33554432 16777216 60"),
			UseLuajit:                       newPtr(false),
			IoThreadsAllowed:                newPtr(false),
			Version:                         &version,
			LuaTimeLimit:                    newPtr(4440),
			ReplBacklogSizePercent:          newPtr(16),
			ClusterRequireFullCoverage:      newPtr(false),
			ClusterAllowReadsWhenDown:       newPtr(false),
			ClusterAllowPubsubshardWhenDown: newPtr(false),
			LfuDecayTime:                    newPtr(22),
			LfuLogFactor:                    newPtr(23),
			TurnBeforeSwitchover:            newPtr(false),
			AllowDataLoss:                   newPtr(false),
			BackupRetainPeriodDays:          newPtr(31),
			ZsetMaxListpackEntries:          newPtr(128),
			BackupWindowStart: &backupWindowStart{
				Hours:   newPtr(20),
				Minutes: newPtr(15),
			},
		},
	}
}

var defaultZone = "ru-central1-d"
var defaultSubnet = "${yandex_vpc_subnet.foo.id}"
var secondSubnet = "${yandex_vpc_subnet.bar.id}"

type hostResource struct {
	ResourcePresetId *string
	DiskSize         *int
	DiskTypeId       *string
}

type access struct {
	DataLens *bool
	WebSql   *bool
}
type maintenanceWindow struct {
	Type *string
	Day  *string
	Hour *int
}

type diskSizeAutoscaling struct {
	DiskSizeLimit           *int
	PlannedUsageThreshold   *int
	EmergencyUsageThreshold *int
}

type valkeyModules struct {
	ValkeySearch *valkeySearch
	ValkeyJson   *valkeyJson
	ValkeyBloom  *valkeyBloom
}

type valkeySearch struct {
	Enabled       *bool
	ReaderThreads *int
	WriterThreads *int
}

type valkeyJson struct {
	Enabled *bool
}

type valkeyBloom struct {
	Enabled *bool
}

type backupWindowStart struct {
	Hours   *int
	Minutes *int
}

type host struct {
	Zone            *string
	ShardName       *string
	SubnetId        *string
	FQDN            *string
	ReplicaPriority *int
	AssignPublicIp  *bool
}

type config struct {
	Password                        *string
	Timeout                         *int
	MaxmemoryPolicy                 *string
	NotifyKeyspaceEvents            *string
	SlowlogLogSlowerThan            *int
	SlowlogMaxLen                   *int
	Databases                       *int
	MaxmemoryPercent                *int
	ClientOutputBufferLimitNormal   *string
	ClientOutputBufferLimitPubsub   *string
	UseLuajit                       *bool
	IoThreadsAllowed                *bool
	Version                         *string
	LuaTimeLimit                    *int
	ReplBacklogSizePercent          *int
	ClusterRequireFullCoverage      *bool
	ClusterAllowReadsWhenDown       *bool
	ClusterAllowPubsubshardWhenDown *bool
	LfuDecayTime                    *int
	LfuLogFactor                    *int
	TurnBeforeSwitchover            *bool
	AllowDataLoss                   *bool
	BackupRetainPeriodDays          *int
	BackupWindowStart               *backupWindowStart
	ZsetMaxListpackEntries          *int
}

type redisConfigTest struct {
	Name                *string
	Environment         *string
	Description         *string
	Sharded             *bool
	TlsEnabled          *bool
	PersistenceMode     *string
	AnnounceHostnames   *bool
	FolderId            *string
	DeletionProtection  *bool
	AuthSentinel        *bool
	DiskEncryptionKeyId *string

	Resources           *hostResource
	Labels              map[string]string
	SecurityGroupIds    []string
	Hosts               map[string]host
	Access              *access
	DiskSizeAutoscaling *diskSizeAutoscaling
	Modules             *valkeyModules
	MaintenanceWindow   *maintenanceWindow
	Config              *config
}

const redisVPCDependencies = `
resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-d"
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["10.3.0.0/24"]
}

resource "yandex_vpc_subnet" "bar" {
  zone           = "ru-central1-b"
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["10.4.0.0/24"]
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

// {{- /*gotype: github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/test/mdb/redis/cluster.redisConfigTest*/ -}}
var templateRedis = redisVPCDependencies + `
resource "yandex_mdb_redis_cluster_v2" "bar" {
  name        = "{{.Name}}"
  {{with .Description}} description  = "{{.}}" {{end}}
  {{with .Environment}} environment  = "{{.}}" {{end}}
  network_id  = "${yandex_vpc_network.foo.id}"
  {{with .Sharded}} sharded  = {{.}} {{end}}
  {{with .TlsEnabled}} tls_enabled  = {{.}} {{end}}
  {{with .PersistenceMode}} persistence_mode  = "{{.}}" {{end}}
  {{with .AnnounceHostnames}} announce_hostnames  = {{.}} {{end}}
  {{with .FolderId}} folder_id  = "{{.}}" {{end}}
  {{with .DeletionProtection}} deletion_protection  = {{.}} {{end}}
  {{with .AuthSentinel}} auth_sentinel  = {{.}} {{end}}
  {{with .DiskEncryptionKeyId}} disk_encryption_key_id  = "{{.}}" {{end}}



  {{with .SecurityGroupIds}}
  security_group_ids = [{{range $i, $r := .}}{{if $i}}, {{end}}"{{.}}"{{end}}]
  {{end}}  

  {{with .Hosts}}
  hosts = {
    {{- range $key, $value := .}}
      "{{ $key }}" = {
	    {{with $value.AssignPublicIp}} assign_public_ip  = {{.}} {{end}}
	    {{with $value.ReplicaPriority}} replica_priority  = {{.}} {{end}}
	    {{with $value.ShardName}} shard_name  = "{{.}}" {{end}}
	    {{with $value.SubnetId}} subnet_id  = "{{.}}" {{end}}
	    {{with $value.Zone}} zone = "{{.}}" {{end}}
      }
    {{- end}}
  }
  {{end}}

  {{with .Labels}}
  labels = {
    {{- range $key, $value := .}}
      {{ $key }} = "{{ $value }}"
    {{- end}}
  }  
  {{end}}

  {{with .Access}}
  access = {
	  {{with .WebSql}} web_sql  = {{.}} {{end}}
	  {{with .DataLens}} data_lens  = {{.}} {{end}}
  }
  {{end}}

  {{with .DiskSizeAutoscaling}}
  disk_size_autoscaling = {
	  {{with .DiskSizeLimit}} disk_size_limit  = {{.}} {{end}}
	  {{with .PlannedUsageThreshold}} planned_usage_threshold  = {{.}} {{end}}
	  {{with .EmergencyUsageThreshold}} emergency_usage_threshold  = {{.}} {{end}}
  }
  {{end}}

  {{with .Modules}}
  modules = {
      {{with .ValkeySearch}} valkey_search = {
		  {{with .Enabled}} enabled = {{.}} {{end}}
		  {{with .ReaderThreads}} reader_threads = {{.}} {{end}}
		  {{with .WriterThreads}} writer_threads = {{.}} {{end}}
	  }
	  {{end}}
	  {{with .ValkeyBloom}} valkey_bloom = {
		  {{with .Enabled}} enabled = {{.}} {{end}}
	  }
	  {{end}}
	  {{with .ValkeyJson}} valkey_json = {
		  {{with .Enabled}} enabled = {{.}} {{end}}
	  }
	  {{end}}
  }
  {{end}}

  {{with .MaintenanceWindow}}
  maintenance_window = {
	  {{with .Type}} type  = "{{.}}" {{end}}
	  {{with .Day}} day  = "{{.}}" {{end}}
	  {{with .Hour}} hour  = {{.}} {{end}}
  }
  {{end}}

  {{with .Resources}}
  resources = {
	  {{with .DiskSize}} disk_size  = {{.}} {{end}}
	  {{with .DiskTypeId}} disk_type_id  = "{{.}}" {{end}}
	  {{with .ResourcePresetId}} resource_preset_id  = "{{.}}" {{end}}
  }
  {{end}}

  {{with .Config}}
  config = {
	  {{with .Password}} password  = "{{.}}" {{end}}
	  {{with .Timeout}} timeout  = {{.}} {{end}}
	  {{with .MaxmemoryPolicy}} maxmemory_policy  = "{{.}}"{{end}}
	  {{with .NotifyKeyspaceEvents}} notify_keyspace_events  = "{{.}}"{{end}}
	  {{with .SlowlogLogSlowerThan}} slowlog_log_slower_than  = {{.}} {{end}}
	  {{with .SlowlogMaxLen}} slowlog_max_len  = {{.}} {{end}}
	  {{with .Databases}} databases  = {{.}} {{end}}
	  {{with .MaxmemoryPercent}} maxmemory_percent  = {{.}} {{end}}
	  {{with .ClientOutputBufferLimitNormal}} client_output_buffer_limit_normal  = "{{.}}" {{end}}
	  {{with .ClientOutputBufferLimitPubsub}} client_output_buffer_limit_pubsub  = "{{.}}" {{end}}
	  {{with .UseLuajit}} use_luajit  = {{.}} {{end}}
	  {{with .IoThreadsAllowed}} io_threads_allowed  = {{.}} {{end}}
	  {{with .Version}} version  = "{{.}}" {{end}}
	  {{with .LuaTimeLimit}} lua_time_limit  = {{.}} {{end}}
	  {{with .ReplBacklogSizePercent}} repl_backlog_size_percent  = {{.}} {{end}}
	  {{with .ClusterRequireFullCoverage}} cluster_require_full_coverage  = {{.}} {{end}}
	  {{with .ClusterAllowReadsWhenDown}} cluster_allow_reads_when_down  = {{.}} {{end}}
	  {{with .ClusterAllowPubsubshardWhenDown}} cluster_allow_pubsubshard_when_down  = {{.}} {{end}}
	  {{with .LfuDecayTime}} lfu_decay_time  = {{.}} {{end}}
	  {{with .LfuLogFactor}} lfu_log_factor  = {{.}} {{end}}
      {{with .ZsetMaxListpackEntries}} zset_max_listpack_entries  = {{.}} {{end}}
	  {{with .TurnBeforeSwitchover}} turn_before_switchover  = {{.}} {{end}}
	  {{with .AllowDataLoss}} allow_data_loss  = {{.}} {{end}}
	  {{with .BackupRetainPeriodDays}} backup_retain_period_days  = {{.}} {{end}}
	  {{with .BackupWindowStart}} 
        backup_window_start = {
            	  {{with .Hours}} hours  = {{.}} {{end}}
	  			  {{with .Minutes}} minutes  = {{.}} {{end}}
      }
      {{end}}
  }
  {{end}}
}
`
