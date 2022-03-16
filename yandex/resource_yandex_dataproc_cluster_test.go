package yandex

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
	"text/template"

	"github.com/golang/protobuf/ptypes/duration"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/golang/protobuf/ptypes"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/dataproc/v1"
)

var testDataprocZone = "ru-central1-b"

func init() {
	zone, ok := os.LookupEnv("YC_ZONE")
	if ok {
		testDataprocZone = zone
	}
	resource.AddTestSweepers("yandex_dataproc_cluster", &resource.Sweeper{
		Name: "yandex_dataproc_cluster",
		F:    testSweepDataprocCluster,
	})
}

func testSweepDataprocCluster(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := conf.sdk.Dataproc().Cluster().List(conf.Context(), &dataproc.ListClustersRequest{
		FolderId: conf.FolderID,
		PageSize: defaultMDBPageSize,
	})
	if err != nil {
		return fmt.Errorf("error getting Data Proc clusters: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Clusters {
		if !sweepDataprocCluster(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Data Proc cluster %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepDataprocCluster(conf *Config, id string) bool {
	return sweepWithRetry(sweepDataprocClusterOnce, conf, "Data Proc cluster", id)
}

func sweepDataprocClusterOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexDataprocClusterDeleteTimeout)
	defer cancel()

	mask := field_mask.FieldMask{Paths: []string{"deletion_protection"}}
	op, err := conf.sdk.Dataproc().Cluster().Update(ctx, &dataproc.UpdateClusterRequest{
		ClusterId:          id,
		DeletionProtection: false,
		UpdateMask:         &mask,
	})
	err = handleSweepOperation(ctx, conf, op, err)
	if err != nil && !strings.EqualFold(errorMessage(err), "no changes detected") {
		return err
	}

	op, err = conf.sdk.Dataproc().Cluster().Delete(ctx, &dataproc.DeleteClusterRequest{
		ClusterId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

func TestIsVersionPrefix(t *testing.T) {
	assert.True(t, isVersionPrefix("1", "1.2.3"))
	assert.True(t, isVersionPrefix("1", "1.2"))
	assert.True(t, isVersionPrefix("1.2", "1.2"))
	assert.True(t, isVersionPrefix("1.2.3", "1.2.3"))
	assert.False(t, isVersionPrefix("1.2.3", "1.2.4"))
	assert.False(t, isVersionPrefix("1.3.3", "1.2.3"))
	assert.False(t, isVersionPrefix("1.3", "1.2.3"))
	assert.False(t, isVersionPrefix("2", "1.2.3"))
	assert.False(t, isVersionPrefix("2.4.4.5", "1.2.3"))
}

func TestExpandDataprocClusterConfig(t *testing.T) {
	raw := map[string]interface{}{
		"folder_id":   "",
		"name":        "dataproc_cluster_777",
		"description": "dataproc cluster 777",
		"labels":      map[string]interface{}{"label1": "val1", "label2": "val2"},
		"cluster_config": []interface{}{
			map[string]interface{}{
				"version_id": "1.1",
				"hadoop": []interface{}{
					map[string]interface{}{
						"services":        []interface{}{"HDFS", "YARN"},
						"properties":      map[string]interface{}{"prop1": "val1", "prop2": "val2"},
						"ssh_public_keys": []interface{}{"id_rsa.pub", "id_dsa.pub"},
					},
				},
				"subcluster_spec": []interface{}{
					map[string]interface{}{
						"name": "main",
						"role": "MASTERNODE",
						"resources": []interface{}{
							map[string]interface{}{
								"resource_preset_id": "s2.large",
								"disk_type_id":       "network-ssd",
								"disk_size":          200,
							},
						},
						"subnet_id":        "subnet-777",
						"assign_public_ip": true,
						"hosts_count":      1,
					},
					map[string]interface{}{
						"name": "data_001",
						"role": "DATANODE",
						"resources": []interface{}{
							map[string]interface{}{
								"resource_preset_id": "s2.small",
								"disk_type_id":       "network-hdd",
								"disk_size":          2000,
							},
						},
						"subnet_id":   "subnet-777",
						"hosts_count": 10,
					},
					map[string]interface{}{
						"name": "compute_001",
						"role": "COMPUTENODE",
						"resources": []interface{}{
							map[string]interface{}{
								"resource_preset_id": "s2.small",
								"disk_type_id":       "network-hdd",
								"disk_size":          2000,
							},
						},
						"subnet_id":   "subnet-777",
						"hosts_count": 10,
					},
					map[string]interface{}{
						"name": "compute_002",
						"role": "COMPUTENODE",
						"resources": []interface{}{
							map[string]interface{}{
								"resource_preset_id": "s2.small",
								"disk_type_id":       "network-hdd",
								"disk_size":          2000,
							},
						},
						"autoscaling_config": []interface{}{
							map[string]interface{}{
								"max_hosts_count":        20,
								"preemptible":            true,
								"warmup_duration":        121,
								"stabilization_duration": 122,
								"measurement_duration":   123,
								"cpu_utilization_target": 82.0,
								"decommission_timeout":   65,
							},
						},
						"subnet_id":   "subnet-777",
						"hosts_count": 10,
					},
				},
			},
		},
		"zone_id":             "ru-central1-b",
		"service_account_id":  "sa-777",
		"bucket":              "bucket-777",
		"ui_proxy":            "true",
		"security_group_ids":  []interface{}{"security_group_id1"},
		"host_group_ids":      []interface{}{"hg1", "hg2"},
		"deletion_protection": "false",
	}
	resourceData := schema.TestResourceDataRaw(t, resourceYandexDataprocCluster().Schema, raw)

	config := &Config{FolderID: "folder-777"}
	req, err := prepareDataprocCreateClusterRequest(resourceData, config)
	if err != nil {
		require.NoError(t, err)
	}

	expected := &dataproc.CreateClusterRequest{
		FolderId:    "folder-777",
		Name:        "dataproc_cluster_777",
		Description: "dataproc cluster 777",
		Labels:      map[string]string{"label1": "val1", "label2": "val2"},
		ConfigSpec: &dataproc.CreateClusterConfigSpec{
			VersionId: "1.1",
			Hadoop: &dataproc.HadoopConfig{
				Services:      []dataproc.HadoopConfig_Service{dataproc.HadoopConfig_HDFS, dataproc.HadoopConfig_YARN},
				Properties:    map[string]string{"prop1": "val1", "prop2": "val2"},
				SshPublicKeys: []string{"id_rsa.pub", "id_dsa.pub"},
			},
			SubclustersSpec: []*dataproc.CreateSubclusterConfigSpec{
				{
					Name: "main",
					Role: dataproc.Role_MASTERNODE,
					Resources: &dataproc.Resources{
						ResourcePresetId: "s2.large",
						DiskTypeId:       "network-ssd",
						DiskSize:         200 * (1 << 30),
					},
					SubnetId:       "subnet-777",
					HostsCount:     1,
					AssignPublicIp: true,
				},
				{
					Name: "data_001",
					Role: dataproc.Role_DATANODE,
					Resources: &dataproc.Resources{
						ResourcePresetId: "s2.small",
						DiskTypeId:       "network-hdd",
						DiskSize:         2000 * (1 << 30),
					},
					SubnetId:   "subnet-777",
					HostsCount: 10,
				},
				{
					Name: "compute_001",
					Role: dataproc.Role_COMPUTENODE,
					Resources: &dataproc.Resources{
						ResourcePresetId: "s2.small",
						DiskTypeId:       "network-hdd",
						DiskSize:         2000 * (1 << 30),
					},
					SubnetId:   "subnet-777",
					HostsCount: 10,
				},
				{
					Name: "compute_002",
					Role: dataproc.Role_COMPUTENODE,
					Resources: &dataproc.Resources{
						ResourcePresetId: "s2.small",
						DiskTypeId:       "network-hdd",
						DiskSize:         2000 * (1 << 30),
					},
					AutoscalingConfig: &dataproc.AutoscalingConfig{
						MaxHostsCount:         20,
						Preemptible:           true,
						WarmupDuration:        &duration.Duration{Seconds: 121},
						StabilizationDuration: &duration.Duration{Seconds: 122},
						MeasurementDuration:   &duration.Duration{Seconds: 123},
						CpuUtilizationTarget:  82,
						DecommissionTimeout:   65,
					},
					SubnetId:   "subnet-777",
					HostsCount: 10,
				},
			},
		},
		ZoneId:             "ru-central1-b",
		ServiceAccountId:   "sa-777",
		Bucket:             "bucket-777",
		UiProxy:            true,
		SecurityGroupIds:   []string{"security_group_id1"},
		HostGroupIds:       []string{"hg2", "hg1"},
		DeletionProtection: false,
	}

	assert.Equal(t, expected, req)
}

func TestFlattenDataprocClusterConfig(t *testing.T) {
	cluster := &dataproc.Cluster{
		Config: &dataproc.ClusterConfig{
			VersionId: "1.4",
			Hadoop: &dataproc.HadoopConfig{
				Services:      []dataproc.HadoopConfig_Service{dataproc.HadoopConfig_HDFS, dataproc.HadoopConfig_YARN},
				Properties:    map[string]string{"prop1": "val1", "prop2": "val2"},
				SshPublicKeys: []string{"id_rsa.pub", "id_dsa.pub"},
			},
		},
	}

	subclusters := []*dataproc.Subcluster{
		{
			Id:   "subcluster-001",
			Name: "main",
			Role: dataproc.Role_MASTERNODE,
			Resources: &dataproc.Resources{
				ResourcePresetId: "s2.large",
				DiskTypeId:       "network-ssd",
				DiskSize:         200 * (1 << 30),
			},
			SubnetId:       "subnet-777",
			HostsCount:     1,
			AssignPublicIp: true,
			CreatedAt:      ptypes.TimestampNow(),
		},
		{
			Id:   "subcluster-002",
			Name: "data_001",
			Role: dataproc.Role_DATANODE,
			Resources: &dataproc.Resources{
				ResourcePresetId: "s2.small",
				DiskTypeId:       "network-hdd",
				DiskSize:         2000 * (1 << 30),
			},
			SubnetId:   "subnet-777",
			HostsCount: 10,
			CreatedAt:  ptypes.TimestampNow(),
		},
		{
			Id:   "subcluster-003",
			Name: "compute_001",
			Role: dataproc.Role_COMPUTENODE,
			Resources: &dataproc.Resources{
				ResourcePresetId: "s2.small",
				DiskTypeId:       "network-hdd",
				DiskSize:         2000 * (1 << 30),
			},
			SubnetId:   "subnet-777",
			HostsCount: 10,
			CreatedAt:  ptypes.TimestampNow(),
		},
		{
			Id:   "subcluster-004",
			Name: "compute_002",
			Role: dataproc.Role_COMPUTENODE,
			Resources: &dataproc.Resources{
				ResourcePresetId: "s2.small",
				DiskTypeId:       "network-hdd",
				DiskSize:         2000 * (1 << 30),
			},
			AutoscalingConfig: &dataproc.AutoscalingConfig{
				MaxHostsCount:         20,
				Preemptible:           true,
				WarmupDuration:        &duration.Duration{Seconds: 121},
				StabilizationDuration: &duration.Duration{Seconds: 122},
				MeasurementDuration:   &duration.Duration{Seconds: 123},
				CpuUtilizationTarget:  82,
				DecommissionTimeout:   65,
			},
			SubnetId:   "subnet-777",
			HostsCount: 10,
			CreatedAt:  ptypes.TimestampNow(),
		},
	}

	config := flattenDataprocClusterConfig(cluster, subclusters)

	expected := []map[string]interface{}{
		{
			"version_id": "1.4",
			"hadoop": []map[string]interface{}{
				{
					"services":        []string{"HDFS", "YARN"},
					"properties":      map[string]string{"prop1": "val1", "prop2": "val2"},
					"ssh_public_keys": []string{"id_rsa.pub", "id_dsa.pub"},
				},
			},
			"subcluster_spec": []interface{}{
				map[string]interface{}{
					"id":   "subcluster-001",
					"name": "main",
					"role": "MASTERNODE",
					"resources": []map[string]interface{}{
						{
							"disk_size":          200,
							"disk_type_id":       "network-ssd",
							"resource_preset_id": "s2.large",
						},
					},
					"subnet_id":        "subnet-777",
					"assign_public_ip": true,
					"hosts_count":      int64(1),
				},
				map[string]interface{}{
					"id":   "subcluster-002",
					"name": "data_001",
					"role": "DATANODE",
					"resources": []map[string]interface{}{
						{
							"disk_size":          2000,
							"disk_type_id":       "network-hdd",
							"resource_preset_id": "s2.small",
						},
					},
					"subnet_id":        "subnet-777",
					"assign_public_ip": false,
					"hosts_count":      int64(10),
				},
				map[string]interface{}{
					"id":   "subcluster-003",
					"name": "compute_001",
					"role": "COMPUTENODE",
					"resources": []map[string]interface{}{
						{
							"disk_size":          2000,
							"disk_type_id":       "network-hdd",
							"resource_preset_id": "s2.small",
						},
					},
					"subnet_id":        "subnet-777",
					"assign_public_ip": false,
					"hosts_count":      int64(10),
				},
				map[string]interface{}{
					"id":   "subcluster-004",
					"name": "compute_002",
					"role": "COMPUTENODE",
					"resources": []map[string]interface{}{
						{
							"disk_size":          2000,
							"disk_type_id":       "network-hdd",
							"resource_preset_id": "s2.small",
						},
					},
					"autoscaling_config": []map[string]interface{}{
						{
							"max_hosts_count":        20,
							"preemptible":            true,
							"warmup_duration":        121,
							"stabilization_duration": 122,
							"measurement_duration":   123,
							"cpu_utilization_target": 82.0,
							"decommission_timeout":   65,
						},
					},
					"subnet_id":        "subnet-777",
					"assign_public_ip": false,
					"hosts_count":      int64(10),
				},
			},
		},
	}

	assert.Equal(t, expected, config)
}

type dataprocTFConfigParams struct {
	Bucket1            string
	Bucket2            string
	CurrentBucket      string
	Description        string
	FolderID           string
	Labels             string
	Name               string
	NetworkName        string
	Properties         string
	SA1Name            string
	SA2Name            string
	SAId               string
	SSHKey             string
	Subcluster1        string
	Subcluster2        string
	Subcluster3        string
	SubnetName         string
	Zone               string
	DeletionProtection bool
}

func (cfg *dataprocTFConfigParams) update(updater func(*dataprocTFConfigParams)) dataprocTFConfigParams {
	updater(cfg)
	return *cfg
}

func defaultDataprocConfigParams(t *testing.T) dataprocTFConfigParams {
	clusterName := acctest.RandomWithPrefix("tf-dataproc")
	description := "Dataproc Cluster created by Terraform"
	folderID := getExampleFolderID()
	sshKey, err := ioutil.ReadFile("test-fixtures/id_rsa.pub")
	if err != nil {
		t.Fatal(err)
	}

	return dataprocTFConfigParams{
		Bucket1:       acctest.RandomWithPrefix("tf-dataproc"),
		Bucket2:       acctest.RandomWithPrefix("tf-dataproc"),
		CurrentBucket: "yandex_storage_bucket.tf-dataproc-1.bucket",
		Description:   description,
		FolderID:      folderID,
		Name:          clusterName,
		NetworkName:   acctest.RandomWithPrefix("tf-dataproc"),
		SA1Name:       acctest.RandomWithPrefix("tf-dataproc"),
		SA2Name:       acctest.RandomWithPrefix("tf-dataproc"),
		SAId:          "yandex_iam_service_account.tf-dataproc-sa.id",
		SSHKey:        string(sshKey),
		SubnetName:    acctest.RandomWithPrefix("tf-dataproc"),
		Zone:          testDataprocZone,
		Labels: `{
				created_by = "terraform"
			}`,
		Subcluster1: `
			subcluster_spec {
				name = "main"
				role = "MASTERNODE"
				resources {
					resource_preset_id = "s2.small"
					disk_type_id       = "network-hdd"
					disk_size          = 24
				}
				subnet_id = yandex_vpc_subnet.tf-dataproc-subnet.id
				hosts_count = 1
			}`,
		Subcluster2: `
			subcluster_spec {
				name = "data"
				role = "DATANODE"
				resources {
					resource_preset_id = "s2.micro"
					disk_type_id       = "network-hdd"
					disk_size          = 24
				}
				subnet_id = yandex_vpc_subnet.tf-dataproc-subnet.id
				hosts_count = 1
			}`,
		Subcluster3: "",
		Properties: `{
			    "yarn:yarn.resourcemanager.am.max-attempts" = 5
			}`,
		DeletionProtection: false,
	}
}

func TestAccDataprocCluster(t *testing.T) {
	var cluster dataproc.Cluster
	templateParams := defaultDataprocConfigParams(t)
	clusterName := templateParams.Name
	services := []string{"HDFS", "YARN", "SPARK", "TEZ", "MAPREDUCE", "HIVE"}
	properties := map[string]string{
		"yarn:yarn.resourcemanager.am.max-attempts": "5",
	}
	updatedProperties := map[string]string{
		"yarn:yarn.resourcemanager.am.max-attempts": "6",
		"hdfs:dfs.webhdfs.enabled":                  "true",
	}
	resourceName := "yandex_dataproc_cluster.tf-dataproc-cluster"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDataprocClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataprocClusterConfig(t, templateParams),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataprocClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "folder_id", getExampleFolderID()),
					resource.TestCheckResourceAttr(resourceName, "zone_id", testDataprocZone),
					resource.TestCheckResourceAttr(resourceName, "name", clusterName),
					resource.TestCheckResourceAttr(resourceName, "description",
						"Dataproc Cluster created by Terraform"),
					resource.TestCheckResourceAttr(resourceName, "bucket", templateParams.Bucket1),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.version_id", "1.4"),
					testAccCheckCreatedAtAttr(resourceName),
					resource.TestCheckResourceAttr(resourceName, "labels.created_by", "terraform"),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
					testAccCheckDataprocClusterServices(&cluster, services),
					testAccCheckDataprocClusterProperties(&cluster, properties),
					testAccCheckDataprocSubclusters(resourceName, map[string]*dataproc.Subcluster{
						"main": {
							Name: "main",
							Role: dataproc.Role_MASTERNODE,
							Resources: &dataproc.Resources{
								ResourcePresetId: "s2.small",
								DiskTypeId:       "network-hdd",
								DiskSize:         24 * (1 << 30),
							},
							HostsCount: 1,
						},
						"data": {
							Name: "data",
							Role: dataproc.Role_DATANODE,
							Resources: &dataproc.Resources{
								ResourcePresetId: "s2.micro",
								DiskTypeId:       "network-hdd",
								DiskSize:         24 * (1 << 30),
							},
							HostsCount: 1,
						},
					}),
				),
			},
			dataprocClusterImportStep(resourceName),
			{
				Config: testAccDataprocClusterConfig(t, templateParams.update(func(cfg *dataprocTFConfigParams) {
					cfg.Name += "-updated"
					cfg.Description += " updated"
					cfg.CurrentBucket = "yandex_storage_bucket.tf-dataproc-2.bucket"
					cfg.SAId = "yandex_iam_service_account.tf-dataproc-sa-2.id"
					cfg.Labels = `{
							created_by = "terraform"
							updated_by = "terraform"
						}`
					cfg.Properties = `{
							"yarn:yarn.resourcemanager.am.max-attempts" = 6
							"hdfs:dfs.webhdfs.enabled" = "true"
						}`
					// modify existing subcluster
					cfg.Subcluster2 = `
						subcluster_spec {
							name = "data-renamed"
							role = "DATANODE"
							resources {
								resource_preset_id = "s2.small"
								disk_type_id       = "network-hdd"
								disk_size          = 32
							}
							subnet_id = yandex_vpc_subnet.tf-dataproc-subnet.id
							hosts_count = 2
						}`
					// add new subcluster
					cfg.Subcluster3 = `
						subcluster_spec {
							name = "compute-1"
							role = "COMPUTENODE"
							resources {
								resource_preset_id = "s2.small"
								disk_type_id       = "network-hdd"
								disk_size          = 24
							}
							subnet_id = yandex_vpc_subnet.tf-dataproc-subnet.id
							hosts_count = 1
						}`
				})),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataprocClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "name", clusterName+"-updated"),
					resource.TestCheckResourceAttr(resourceName, "description",
						"Dataproc Cluster created by Terraform updated"),
					resource.TestCheckResourceAttr(resourceName, "bucket", templateParams.Bucket2),
					resource.TestCheckResourceAttr(resourceName, "labels.created_by", "terraform"),
					resource.TestCheckResourceAttr(resourceName, "labels.updated_by", "terraform"),
					testAccCheckDataprocClusterProperties(&cluster, updatedProperties),
					testAccCheckDataprocSubclusters(resourceName, map[string]*dataproc.Subcluster{
						"main": {
							Name: "main",
							Role: dataproc.Role_MASTERNODE,
							Resources: &dataproc.Resources{
								ResourcePresetId: "s2.small",
								DiskTypeId:       "network-hdd",
								DiskSize:         24 * (1 << 30),
							},
							HostsCount: 1,
						},
						"data-renamed": {
							Name: "data-renamed",
							Role: dataproc.Role_DATANODE,
							Resources: &dataproc.Resources{
								ResourcePresetId: "s2.small",
								DiskTypeId:       "network-hdd",
								DiskSize:         32 * (1 << 30),
							},
							HostsCount: 2,
						},
						"compute-1": {
							Name: "compute-1",
							Role: dataproc.Role_COMPUTENODE,
							Resources: &dataproc.Resources{
								ResourcePresetId: "s2.small",
								DiskTypeId:       "network-hdd",
								DiskSize:         24 * (1 << 30),
							},
							HostsCount: 1,
						},
					}),
				),
			},
			dataprocClusterImportStep(resourceName),
			{
				Config: testAccDataprocClusterConfig(t, templateParams.update(func(cfg *dataprocTFConfigParams) {
					// delete subcluster
					cfg.Subcluster3 = ""
				})),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataprocClusterExists(resourceName, &cluster),
					testAccCheckDataprocSubclusters(resourceName, map[string]*dataproc.Subcluster{
						"main": {
							Name: "main",
							Role: dataproc.Role_MASTERNODE,
							Resources: &dataproc.Resources{
								ResourcePresetId: "s2.small",
								DiskTypeId:       "network-hdd",
								DiskSize:         24 * (1 << 30),
							},
							HostsCount: 1,
						},
						"data-renamed": {
							Name: "data-renamed",
							Role: dataproc.Role_DATANODE,
							Resources: &dataproc.Resources{
								ResourcePresetId: "s2.small",
								DiskTypeId:       "network-hdd",
								DiskSize:         32 * (1 << 30),
							},
							HostsCount: 2,
						},
					}),
				),
			},
			dataprocClusterImportStep(resourceName),
		},
	})
}

func testAccDataprocClusterConfig(t *testing.T, templateParams dataprocTFConfigParams) string {
	tfConfigTemplate := `
resource "yandex_vpc_network" "tf-dataproc-net" {
  name = "{{.NetworkName}}"
}

resource "yandex_vpc_subnet" "tf-dataproc-subnet" {
  name           = "{{.SubnetName}}"
  zone           = "{{.Zone}}"
  network_id     = yandex_vpc_network.tf-dataproc-net.id
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_iam_service_account" "tf-dataproc-sa" {
  name        = "{{.SA1Name}}"
  description = "service account to manage Dataproc Cluster created by Terraform"
}

resource "yandex_iam_service_account" "tf-dataproc-sa-2" {
  name        = "{{.SA2Name}}"
  description = "service account to manage Dataproc Cluster created by Terraform"
}

resource "yandex_resourcemanager_folder_iam_member" "dataproc-manager" {
	folder_id   = "{{.FolderID}}"
	member      = "serviceAccount:${yandex_iam_service_account.tf-dataproc-sa.id}"
	role        = "mdb.dataproc.agent"
	sleep_after = 30
}

resource "yandex_resourcemanager_folder_iam_member" "dataproc-manager-2" {
	folder_id   = "{{.FolderID}}"
	member      = "serviceAccount:${yandex_iam_service_account.tf-dataproc-sa-2.id}"
	role        = "mdb.dataproc.agent"
	sleep_after = 30
}

// required in order to create bucket
resource "yandex_resourcemanager_folder_iam_member" "bucket-creator" {
	folder_id   = "{{.FolderID}}"
	member      = "serviceAccount:${yandex_iam_service_account.tf-dataproc-sa.id}"
	role        = "editor"
	sleep_after = 30
}

resource "yandex_iam_service_account_static_access_key" "tf-dataproc-sa-static-key" {
  service_account_id = yandex_iam_service_account.tf-dataproc-sa.id
  description        = "static access key for object storage"

  depends_on = [
    yandex_resourcemanager_folder_iam_member.bucket-creator
  ]
}

resource "yandex_storage_bucket" "tf-dataproc-1" {
  bucket     = "{{.Bucket1}}"
  access_key = yandex_iam_service_account_static_access_key.tf-dataproc-sa-static-key.access_key
  secret_key = yandex_iam_service_account_static_access_key.tf-dataproc-sa-static-key.secret_key
}

resource "yandex_storage_bucket" "tf-dataproc-2" {
  bucket     = "{{.Bucket2}}"
  access_key = yandex_iam_service_account_static_access_key.tf-dataproc-sa-static-key.access_key
  secret_key = yandex_iam_service_account_static_access_key.tf-dataproc-sa-static-key.secret_key
}

resource "yandex_dataproc_cluster" "tf-dataproc-cluster" {
  depends_on = [yandex_resourcemanager_folder_iam_member.dataproc-manager,
				yandex_resourcemanager_folder_iam_member.dataproc-manager-2]

  bucket             = {{.CurrentBucket}}
  description        = "{{.Description}}"
  labels             = {{.Labels}}
  name               = "{{.Name}}"
  service_account_id = {{.SAId}}
  zone_id            = "{{.Zone}}"
  deletion_protection = {{.DeletionProtection}}

  cluster_config {
    version_id = "1.4"

    hadoop {
      services = ["HDFS", "YARN", "SPARK", "TEZ", "MAPREDUCE", "HIVE"]
      properties = {{.Properties}}
      ssh_public_keys = ["{{.SSHKey}}"]
    }

	{{.Subcluster1}}
	{{.Subcluster2}}
	{{.Subcluster3}}
  }
}
`
	tfConfig := bytes.Buffer{}
	err := template.
		Must(template.New("main.tf").Parse(tfConfigTemplate)).
		Execute(&tfConfig, templateParams)
	require.NoError(t, err)
	return tfConfig.String()
}

func testAccCheckDataprocClusterExists(resourceName string, cluster *dataproc.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		clusterID, err := getResourceID(resourceName, s)
		if err != nil {
			return err
		}

		config := testAccProvider.Meta().(*Config)
		found, err := config.sdk.Dataproc().Cluster().Get(context.Background(), &dataproc.GetClusterRequest{
			ClusterId: clusterID,
		})
		if err != nil {
			return err
		}

		*cluster = *found
		return nil
	}
}

func testAccCheckDataprocSubclusters(resource string, expectedSubclusters map[string]*dataproc.Subcluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		subnetID, err := getResourceID("yandex_vpc_subnet.tf-dataproc-subnet", s)
		if err != nil {
			return err
		}

		clusterID, err := getResourceID(resource, s)
		if err != nil {
			return err
		}

		config := testAccProvider.Meta().(*Config)
		resp, err := config.sdk.Dataproc().Subcluster().List(context.Background(), &dataproc.ListSubclustersRequest{
			ClusterId: clusterID,
			PageSize:  defaultMDBPageSize,
		})
		if err != nil {
			return err
		}

		realSubclusters := resp.Subclusters
		realSubclusterByName := make(map[string]*dataproc.Subcluster)
		for _, subcluster := range realSubclusters {
			realSubclusterByName[subcluster.Name] = subcluster
		}

		for name, expectedSubcluster := range expectedSubclusters {
			realSubcluster, ok := realSubclusterByName[name]
			if !ok {
				return fmt.Errorf("subcluster not found '%s'", name)
			}

			if realSubcluster.Role != expectedSubcluster.Role {
				return fmt.Errorf("invalid role for subcluster '%s': expected '%s', got '%s'",
					name, expectedSubcluster.Role.String(), realSubcluster.Role.String())
			}

			if realSubcluster.HostsCount != expectedSubcluster.HostsCount {
				return fmt.Errorf("invalid hosts count for subcluster '%s': expected '%d', got '%d'",
					name, expectedSubcluster.HostsCount, realSubcluster.HostsCount)
			}

			if realSubcluster.Resources.ResourcePresetId != expectedSubcluster.Resources.ResourcePresetId {
				return fmt.Errorf("invalid resource preset id for subcluster '%s': expected '%s', got '%s'",
					name, expectedSubcluster.Resources.ResourcePresetId, realSubcluster.Resources.ResourcePresetId)
			}

			if realSubcluster.Resources.DiskTypeId != expectedSubcluster.Resources.DiskTypeId {
				return fmt.Errorf("invalid disk type for subcluster '%s': expected '%s', got '%s'",
					name, expectedSubcluster.Resources.DiskTypeId, realSubcluster.Resources.DiskTypeId)
			}

			if realSubcluster.Resources.DiskSize != expectedSubcluster.Resources.DiskSize {
				return fmt.Errorf("invalid disk size for subcluster '%s': expected  %d, got %d",
					name, expectedSubcluster.Resources.DiskSize, realSubcluster.Resources.DiskSize)
			}

			if realSubcluster.SubnetId != subnetID {
				return fmt.Errorf("invalid subnet id for subcluster '%s': expected  '%s', got '%s'",
					name, subnetID, realSubcluster.SubnetId)
			}
		}

		for _, realSubcluster := range realSubclusters {
			expected := false
			for _, expectedSubcluster := range expectedSubclusters {
				if expectedSubcluster.Name == realSubcluster.Name {
					expected = true
					break
				}
			}
			if !expected {
				return fmt.Errorf("got unexpected subcluster '%s'", realSubcluster.Name)
			}
		}

		return nil
	}
}

func testAccCheckDataprocClusterServices(r *dataproc.Cluster, services []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		realServices := make([]string, len(r.Config.Hadoop.Services))
		for idx, service := range r.Config.Hadoop.Services {
			realServices[idx] = service.String()
		}
		sort.Strings(realServices)
		sort.Strings(services)
		if !reflect.DeepEqual(realServices, services) {
			return fmt.Errorf("incorrect list of services: expected %#v, got %#v", services, realServices)
		}
		return nil
	}
}

func testAccCheckDataprocClusterProperties(r *dataproc.Cluster, properties map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		realProperties := r.Config.Hadoop.Properties
		if !reflect.DeepEqual(realProperties, properties) {
			return fmt.Errorf("incorrect list of properties: expected %#v, got %#v", properties, realProperties)
		}
		return nil
	}
}

func dataprocClusterImportStep(name string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      name,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateVerifyIgnore: []string{
			// this will not pass the check because we reorder subclusters returned by the cloud
			"cluster_config.0.subcluster_spec",
		},
	}
}

func testAccCheckDataprocClusterDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_dataproc_cluster" {
			continue
		}

		_, err := config.sdk.Dataproc().Cluster().Get(context.Background(), &dataproc.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})

		if err == nil {
			return fmt.Errorf("expected Data Proc to be deleted, but it still exists")
		}
	}

	return nil
}
