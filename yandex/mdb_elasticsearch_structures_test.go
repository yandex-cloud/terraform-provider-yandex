package yandex

import (
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/elasticsearch/v1"
)

func TestParseElasticsearchEnv(t *testing.T) {
	_, err := parseElasticsearchEnv("SOME")
	require.Error(t, err, "parsed unexpected environment")

	env, err := parseElasticsearchEnv("PRODUCTION")
	require.NoError(t, err, "environment not parsed")
	require.Equal(t, env, elasticsearch.Cluster_PRODUCTION)

	env, err = parseElasticsearchEnv("PRESTABLE")
	require.NoError(t, err, "environment not parsed")
	require.Equal(t, env, elasticsearch.Cluster_PRESTABLE)
}

func TestParseElasticsearchHostType(t *testing.T) {
	_, err := parseElasticsearchHostType("SOME")
	require.Error(t, err, "parsed unexpected host type")

	host, err := parseElasticsearchHostType("DATA_NODE")
	require.NoError(t, err, "host type not parsed")
	require.Equal(t, host, elasticsearch.Host_DATA_NODE)

	host, err = parseElasticsearchHostType("MASTER_NODE")
	require.NoError(t, err, "host type not parsed")
	require.Equal(t, host, elasticsearch.Host_MASTER_NODE)
}

func TestExpandElasticsearcHosts(t *testing.T) {
	raw := []interface{}{map[string]interface{}{
		"name":             "nodename",
		"fqdn":             "somecluster.at.yandex",
		"zone":             "sas",
		"type":             "DATA_NODE",
		"subnet_id":        "subnet",
		"assign_public_ip": true,
	}, map[string]interface{}{
		"zone": "man",
		"type": "MASTER_NODE",
	}}

	expected := ElasticsearchHostList{
		{Zone: "man", Type: elasticsearch.Host_MASTER_NODE},
		{Name: "nodename", Fqdn: "somecluster.at.yandex", Zone: "sas", Type: elasticsearch.Host_DATA_NODE, Subnet: "subnet", PublicIp: true},
	}

	hosts, err := expandElasticsearchHosts(schema.NewSet(schema.HashResource(elasticsearchHostResource), raw))
	require.NoError(t, err, "failed expand elasticsearch host specs")
	sort.Slice(hosts, func(i, j int) bool { return hosts[i].Zone < hosts[j].Zone })

	require.Equal(t, expected, hosts)
}

func TestExpandElasticsearchConfigSpec(t *testing.T) {
	raw := map[string]interface{}{
		"config": []interface{}{map[string]interface{}{
			"version":        "7.10",
			"edition":        "basic",
			"admin_password": "password",
			"data_node": []interface{}{map[string]interface{}{
				"resources": []interface{}{map[string]interface{}{
					"resource_preset_id": "s2.micro",
					"disk_type_id":       "network-ssd",
					"disk_size":          10,
				}},
			}},
			"master_node": []interface{}{map[string]interface{}{
				"resources": []interface{}{map[string]interface{}{
					"resource_preset_id": "s2.micro",
					"disk_type_id":       "network-ssd",
					"disk_size":          10,
				}},
			}},
			"plugins": []interface{}{"analysis-icu"},
		}},
		"host": []interface{}{map[string]interface{}{
			"name": "data",
			"zone": "sas",
			"type": "DATA_NODE",
		}, map[string]interface{}{
			"name": "master",
			"zone": "man",
			"type": "MASTER_NODE",
		}},
	}

	expected := &elasticsearch.ConfigSpec{
		Version:       "7.10",
		Edition:       "basic",
		AdminPassword: "password",
		ElasticsearchSpec: &elasticsearch.ElasticsearchSpec{
			DataNode: &elasticsearch.ElasticsearchSpec_DataNode{
				Resources: &elasticsearch.Resources{
					ResourcePresetId: "s2.micro",
					DiskTypeId:       "network-ssd",
					DiskSize:         10 * 1024 * 1024 * 1024,
				},
			},
			MasterNode: &elasticsearch.ElasticsearchSpec_MasterNode{
				Resources: &elasticsearch.Resources{
					ResourcePresetId: "s2.micro",
					DiskTypeId:       "network-ssd",
					DiskSize:         10 * 1024 * 1024 * 1024,
				},
			},
			Plugins: []string{"analysis-icu"},
		},
	}

	resourceData := schema.TestResourceDataRaw(t, resourceYandexMDBElasticsearchCluster().Schema, raw)

	spec := expandElasticsearchConfigSpec(resourceData)

	require.Equal(t, expected, spec)
}

func TestExpandElasticsearchConfigSpec_SuppressMasters(t *testing.T) {
	raw := map[string]interface{}{
		"config": []interface{}{map[string]interface{}{
			"version":        "7.10",
			"edition":        "basic",
			"admin_password": "password",
			"data_node": []interface{}{map[string]interface{}{
				"resources": []interface{}{map[string]interface{}{
					"resource_preset_id": "s2.micro",
					"disk_type_id":       "network-ssd",
					"disk_size":          10,
				}},
			}},
			"master_node": []interface{}{map[string]interface{}{
				"resources": []interface{}{map[string]interface{}{
					"resource_preset_id": "s2.micro",
					"disk_type_id":       "network-ssd",
					"disk_size":          10,
				}},
			}},
			"plugins": []interface{}{"analysis-icu"},
		}},
		"host": []interface{}{map[string]interface{}{
			"name": "data",
			"zone": "sas",
			"type": "DATA_NODE",
		}},
	}

	expected := &elasticsearch.ConfigSpec{
		Version:       "7.10",
		Edition:       "basic",
		AdminPassword: "password",
		ElasticsearchSpec: &elasticsearch.ElasticsearchSpec{
			DataNode: &elasticsearch.ElasticsearchSpec_DataNode{
				Resources: &elasticsearch.Resources{
					ResourcePresetId: "s2.micro",
					DiskTypeId:       "network-ssd",
					DiskSize:         10 * 1024 * 1024 * 1024,
				},
			},
			Plugins: []string{"analysis-icu"},
		},
	}

	resourceData := schema.TestResourceDataRaw(t, resourceYandexMDBElasticsearchCluster().Schema, raw)

	spec := expandElasticsearchConfigSpec(resourceData)

	require.Equal(t, expected, spec)
}

func TestExpandElasticsearchConfigSpec_MinRequired(t *testing.T) {
	raw := map[string]interface{}{
		"config": []interface{}{map[string]interface{}{
			"admin_password": "password",
			"data_node": []interface{}{map[string]interface{}{
				"resources": []interface{}{map[string]interface{}{
					"resource_preset_id": "s2.micro",
					"disk_type_id":       "network-ssd",
					"disk_size":          10,
				}},
			}},
		}},
	}

	expected := &elasticsearch.ConfigSpec{
		AdminPassword: "password",
		ElasticsearchSpec: &elasticsearch.ElasticsearchSpec{
			DataNode: &elasticsearch.ElasticsearchSpec_DataNode{
				Resources: &elasticsearch.Resources{
					ResourcePresetId: "s2.micro",
					DiskTypeId:       "network-ssd",
					DiskSize:         10 * 1024 * 1024 * 1024,
				},
			},
			Plugins: []string{},
		},
	}

	resourceData := schema.TestResourceDataRaw(t, resourceYandexMDBElasticsearchCluster().Schema, raw)

	spec := expandElasticsearchConfigSpec(resourceData)

	require.Equal(t, expected, spec)
}

func TestFlattenElasticsearchConfig(t *testing.T) {
	config := &elasticsearch.ClusterConfig{
		Version: "7.10",
		Edition: "basic",
		Elasticsearch: &elasticsearch.Elasticsearch{
			DataNode: &elasticsearch.Elasticsearch_DataNode{
				Resources: &elasticsearch.Resources{
					ResourcePresetId: "s2.micro",
					DiskTypeId:       "network-ssd",
					DiskSize:         10 * 1024 * 1024 * 1024,
				},
			},
			MasterNode: &elasticsearch.Elasticsearch_MasterNode{
				Resources: &elasticsearch.Resources{
					ResourcePresetId: "s2.micro",
					DiskTypeId:       "network-ssd",
					DiskSize:         10 * 1024 * 1024 * 1024,
				},
			},
			Plugins: []string{"analysis-icu"},
		},
	}

	expected := []interface{}{map[string]interface{}{
		"version":        "7.10",
		"edition":        "basic",
		"admin_password": "password",
		"data_node": []interface{}{map[string]interface{}{
			"resources": []interface{}{map[string]interface{}{
				"resource_preset_id": "s2.micro",
				"disk_type_id":       "network-ssd",
				"disk_size":          10,
			}},
		}},
		"master_node": []interface{}{map[string]interface{}{
			"resources": []interface{}{map[string]interface{}{
				"resource_preset_id": "s2.micro",
				"disk_type_id":       "network-ssd",
				"disk_size":          10,
			}},
		}},
		"plugins": []string{"analysis-icu"},
	}}

	raw := flattenElasticsearchClusterConfig(config, "password")

	require.Equal(t, expected, raw)
}

func TestExpandElasticsearchConfigSpecUpdate(t *testing.T) {
	raw := map[string]interface{}{
		"config": []interface{}{map[string]interface{}{
			"version":        "7.10",
			"edition":        "basic",
			"admin_password": "password",
			"data_node": []interface{}{map[string]interface{}{
				"resources": []interface{}{map[string]interface{}{
					"resource_preset_id": "s2.micro",
					"disk_type_id":       "network-ssd",
					"disk_size":          10,
				}},
			}},
			"master_node": []interface{}{map[string]interface{}{
				"resources": []interface{}{map[string]interface{}{
					"resource_preset_id": "s2.micro",
					"disk_type_id":       "network-ssd",
					"disk_size":          10,
				}},
			}},
			"plugins": []interface{}{"analysis-icu"},
		}},
		"host": []interface{}{map[string]interface{}{
			"name": "data",
			"zone": "sas",
			"type": "DATA_NODE",
		}, map[string]interface{}{
			"name": "master",
			"zone": "man",
			"type": "MASTER_NODE",
		}},
	}

	expected := &elasticsearch.ConfigSpecUpdate{
		Version:       "7.10",
		Edition:       "basic",
		AdminPassword: "password",
		ElasticsearchSpec: &elasticsearch.ElasticsearchSpec{
			DataNode: &elasticsearch.ElasticsearchSpec_DataNode{
				Resources: &elasticsearch.Resources{
					ResourcePresetId: "s2.micro",
					DiskTypeId:       "network-ssd",
					DiskSize:         10 * 1024 * 1024 * 1024,
				},
			},
			MasterNode: &elasticsearch.ElasticsearchSpec_MasterNode{
				Resources: &elasticsearch.Resources{
					ResourcePresetId: "s2.micro",
					DiskTypeId:       "network-ssd",
					DiskSize:         10 * 1024 * 1024 * 1024,
				},
			},
			Plugins: []string{"analysis-icu"},
		},
	}

	resourceData := schema.TestResourceDataRaw(t, resourceYandexMDBElasticsearchCluster().Schema, raw)

	spec := expandElasticsearchConfigSpecUpdate(resourceData)

	require.Equal(t, expected, spec)
}

func TestExpandElasticsearchConfigSpecUpdate_SuppressMasters(t *testing.T) {
	raw := map[string]interface{}{
		"config": []interface{}{map[string]interface{}{
			"version":        "7.10",
			"edition":        "basic",
			"admin_password": "password",
			"data_node": []interface{}{map[string]interface{}{
				"resources": []interface{}{map[string]interface{}{
					"resource_preset_id": "s2.micro",
					"disk_type_id":       "network-ssd",
					"disk_size":          10,
				}},
			}},
			"master_node": []interface{}{map[string]interface{}{
				"resources": []interface{}{map[string]interface{}{
					"resource_preset_id": "s2.micro",
					"disk_type_id":       "network-ssd",
					"disk_size":          10,
				}},
			}},
			"plugins": []interface{}{"analysis-icu"},
		}},
		"host": []interface{}{map[string]interface{}{
			"name": "data",
			"zone": "sas",
			"type": "DATA_NODE",
		}},
	}

	expected := &elasticsearch.ConfigSpecUpdate{
		Version:       "7.10",
		Edition:       "basic",
		AdminPassword: "password",
		ElasticsearchSpec: &elasticsearch.ElasticsearchSpec{
			DataNode: &elasticsearch.ElasticsearchSpec_DataNode{
				Resources: &elasticsearch.Resources{
					ResourcePresetId: "s2.micro",
					DiskTypeId:       "network-ssd",
					DiskSize:         10 * 1024 * 1024 * 1024,
				},
			},
			Plugins: []string{"analysis-icu"},
		},
	}

	resourceData := schema.TestResourceDataRaw(t, resourceYandexMDBElasticsearchCluster().Schema, raw)

	spec := expandElasticsearchConfigSpecUpdate(resourceData)

	require.Equal(t, expected, spec)
}

func TestExpandElasticsearchConfigSpecUpdate_MinRequired(t *testing.T) {
	raw := map[string]interface{}{
		"config": []interface{}{map[string]interface{}{
			"admin_password": "password",
			"data_node": []interface{}{map[string]interface{}{
				"resources": []interface{}{map[string]interface{}{
					"resource_preset_id": "s2.micro",
					"disk_type_id":       "network-ssd",
					"disk_size":          10,
				}},
			}},
		}},
	}

	expected := &elasticsearch.ConfigSpecUpdate{
		AdminPassword: "password",
		ElasticsearchSpec: &elasticsearch.ElasticsearchSpec{
			DataNode: &elasticsearch.ElasticsearchSpec_DataNode{
				Resources: &elasticsearch.Resources{
					ResourcePresetId: "s2.micro",
					DiskTypeId:       "network-ssd",
					DiskSize:         10 * 1024 * 1024 * 1024,
				},
			},
			Plugins: []string{},
		},
	}

	resourceData := schema.TestResourceDataRaw(t, resourceYandexMDBElasticsearchCluster().Schema, raw)

	spec := expandElasticsearchConfigSpecUpdate(resourceData)

	require.Equal(t, expected, spec)
}
