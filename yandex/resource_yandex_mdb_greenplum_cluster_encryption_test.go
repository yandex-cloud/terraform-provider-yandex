package yandex

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestMDBGreenplumClusterDiskEncryptionRequests(t *testing.T) {
	for _, tt := range []struct{ name, key string }{
		{name: "without encryption"},
		{name: "with encryption", key: "kms-key-id"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			raw := map[string]interface{}{
				"name": "test-cluster", "folder_id": "folder-id", "network_id": "network-id",
				"environment": "PRESTABLE", "zone": "ru-central1-a", "subnet_id": "subnet-id",
				"user_name": "user", "user_password": "test-password", "version": "6.25",
				"restore": []interface{}{map[string]interface{}{"backup_id": "backup-id", "restore_pxf": true}},
			}
			if tt.key != "" {
				raw["disk_encryption_key_id"] = tt.key
			}
			r := resourceYandexMDBGreenplumCluster()
			if err := r.InternalValidate(nil, true); err != nil {
				t.Fatal(err)
			}
			field := r.Schema["disk_encryption_key_id"]
			if !field.Optional || !field.ForceNew || field.Computed {
				t.Fatal("encryption key must be optional and require replacement")
			}
			d := schema.TestResourceDataRaw(t, r.Schema, raw)
			create, err := prepareCreateGreenplumClusterRequest(d, &Config{})
			if err != nil {
				t.Fatal(err)
			}
			if (create.DiskEncryptionKeyId == nil) != (tt.key == "") || create.GetDiskEncryptionKeyId().GetValue() != tt.key {
				t.Fatalf("unexpected create encryption key: %v", create.DiskEncryptionKeyId)
			}
			restore, err := prepareRestoreGreenplumClusterRequest(d, create, "backup-id")
			if err != nil {
				t.Fatal(err)
			}
			if restore.DiskEncryptionKeyId == nil || restore.DiskEncryptionKeyId.GetValue() != tt.key {
				t.Fatalf("restore must explicitly set encryption key to %q, got %v", tt.key, restore.DiskEncryptionKeyId)
			}
			if restore.BackupId != "backup-id" || !restore.RestorePxf || restore.NetworkId != create.NetworkId {
				t.Fatal("restore options were not preserved")
			}
			if tt.key == "" && create.DiskEncryptionKeyId != nil {
				t.Fatal("restore mutated create request")
			}
		})
	}
}
