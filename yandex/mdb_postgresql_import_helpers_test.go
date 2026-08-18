package yandex

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"
)

func TestPGImportStepCompareHostsByFQDN(t *testing.T) {
	// state after apply: config order na(a), nd(b), nb(b), nc(d)
	applied := map[string]string{
		"host.#":                    "4",
		"host.0.fqdn":               "rc1a-aaaa.mdb.yandexcloud.net",
		"host.0.zone":               "ru-central1-a",
		"host.0.assign_public_ip":   "true",
		"host.0.priority":           "0",
		"host.1.fqdn":               "rc1b-ghlt.mdb.yandexcloud.net",
		"host.1.zone":               "ru-central1-b",
		"host.1.assign_public_ip":   "false",
		"host.2.fqdn":               "rc1b-c28n.mdb.yandexcloud.net",
		"host.2.zone":               "ru-central1-b",
		"host.2.assign_public_ip":   "false",
		"host.3.fqdn":               "rc1d-dddd.mdb.yandexcloud.net",
		"host.3.zone":               "ru-central1-d",
		"host.3.replication_source": "rc1b-c28n.mdb.yandexcloud.net",
	}

	// state after import: API order, hosts of ru-central1-b are swapped
	imported := map[string]string{
		"host.#":                    "4",
		"host.0.fqdn":               "rc1a-aaaa.mdb.yandexcloud.net",
		"host.0.zone":               "ru-central1-a",
		"host.0.assign_public_ip":   "true",
		"host.0.priority":           "0",
		"host.1.fqdn":               "rc1b-c28n.mdb.yandexcloud.net",
		"host.1.zone":               "ru-central1-b",
		"host.1.assign_public_ip":   "false",
		"host.2.fqdn":               "rc1b-ghlt.mdb.yandexcloud.net",
		"host.2.zone":               "ru-central1-b",
		"host.2.assign_public_ip":   "false",
		"host.3.fqdn":               "rc1d-dddd.mdb.yandexcloud.net",
		"host.3.zone":               "ru-central1-d",
		"host.3.replication_source": "rc1b-c28n.mdb.yandexcloud.net",
	}

	savedHosts := new(map[string]map[string]string)
	err := testAccPGSaveHostAttrs("yandex_mdb_postgresql_cluster.ha_cluster_with_names", savedHosts)(&terraform.State{
		Modules: []*terraform.ModuleState{
			{
				Path: []string{"root"},
				Resources: map[string]*terraform.ResourceState{
					"yandex_mdb_postgresql_cluster.ha_cluster_with_names": {
						Type:    pgResourceType,
						Primary: &terraform.InstanceState{Attributes: applied},
					},
				},
			},
		},
	})
	require.NoError(t, err)
	require.Len(t, *savedHosts, 4)

	step := mdbPGClusterImportStepCompareHostsByFQDN("yandex_mdb_postgresql_cluster.ha_cluster_with_names", savedHosts)
	require.Contains(t, step.ImportStateVerifyIgnore, "host.")

	importedState := &terraform.InstanceState{Attributes: imported}
	importedState.Ephemeral.Type = pgResourceType

	// swapped hosts of the same zone are fine
	require.NoError(t, step.ImportStateCheck([]*terraform.InstanceState{importedState}))

	// a real change of a host attribute is still detected
	imported["host.1.assign_public_ip"] = "true"
	require.ErrorContains(t, step.ImportStateCheck([]*terraform.InstanceState{importedState}), "rc1b-c28n")

	// a lost host is detected
	imported["host.1.assign_public_ip"] = "false"
	imported["host.1.fqdn"] = "rc1b-zzzz.mdb.yandexcloud.net"
	require.ErrorContains(t, step.ImportStateCheck([]*terraform.InstanceState{importedState}), "is missing after import")

	// a lost hosts count is detected
	imported["host.#"] = "3"
	require.ErrorContains(t, step.ImportStateCheck([]*terraform.InstanceState{importedState}), "got 3 hosts after import, expected 4")

	imported["host.#"] = "4"
	imported["host.1.fqdn"] = "rc1b-c28n.mdb.yandexcloud.net"
	require.NoError(t, step.ImportStateCheck([]*terraform.InstanceState{importedState}))

	// an attribute added to the host schema is compared without touching the helper
	imported["host.1.new_host_attr"] = "true"
	require.ErrorContains(t, step.ImportStateCheck([]*terraform.InstanceState{importedState}), `attribute "new_host_attr"`)

	// the same for an attribute lost after import
	delete(imported, "host.1.new_host_attr")
	delete(imported, "host.1.zone")
	require.ErrorContains(t, step.ImportStateCheck([]*terraform.InstanceState{importedState}), `attribute "zone"`)

	// attributes not returned by the API are still not compared
	imported["host.1.zone"] = "ru-central1-b"
	imported["host.1.name"] = "nb"
	imported["host.1.role"] = "PRIMARY"
	imported["host.1.replication_source_name"] = "na"
	require.NoError(t, step.ImportStateCheck([]*terraform.InstanceState{importedState}))
}
