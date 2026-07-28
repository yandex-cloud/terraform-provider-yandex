//go:build tf1_12

package yandex

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	sdkSchema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkTerraform "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestMDBPostgreSQLUserPasswordWoRequiredWith(t *testing.T) {
	resourceSchema := sdkSchema.InternalMap(resourceYandexMDBPostgreSQLUser().Schema)

	tests := []struct {
		name      string
		config    map[string]interface{}
		wantError bool
	}{
		{
			name: "write-only password with version",
			config: map[string]interface{}{
				"cluster_id":          "cluster-id",
				"name":                "alice",
				"password_wo":         "mysecureP@ssw0rd",
				"password_wo_version": 1,
			},
		},
		{
			name: "write-only password without version",
			config: map[string]interface{}{
				"cluster_id":  "cluster-id",
				"name":        "alice",
				"password_wo": "mysecureP@ssw0rd",
			},
			wantError: true,
		},
		{
			name: "version without write-only password",
			config: map[string]interface{}{
				"cluster_id":          "cluster-id",
				"name":                "alice",
				"password_wo_version": 1,
			},
			wantError: true,
		},
		{
			name: "legacy password",
			config: map[string]interface{}{
				"cluster_id": "cluster-id",
				"name":       "alice",
				"password":   "mysecureP@ssw0rd",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := resourceSchema.Validate(sdkTerraform.NewResourceConfigRaw(tt.config))
			if gotError := diags.HasError(); gotError != tt.wantError {
				t.Fatalf("schema validation error = %t, want %t; diagnostics: %#v", gotError, tt.wantError, diags)
			}
		})
	}
}

type testPgUserRawConfig struct {
	value cty.Value
}

func (c testPgUserRawConfig) GetRawConfig() cty.Value {
	return c.value
}

func TestValidatePgUserPasswordConflict(t *testing.T) {
	tests := []struct {
		name      string
		config    map[string]cty.Value
		wantError bool
	}{
		{
			name: "legacy and write-only passwords conflict",
			config: map[string]cty.Value{
				"password":    cty.StringVal("legacy-password"),
				"password_wo": cty.StringVal("write-only-password"),
			},
			wantError: true,
		},
		{
			name: "legacy password only",
			config: map[string]cty.Value{
				"password": cty.StringVal("legacy-password"),
			},
		},
		{
			name: "write-only password only",
			config: map[string]cty.Value{
				"password_wo": cty.StringVal("write-only-password"),
			},
		},
		{
			name: "unknown legacy password defers conflict",
			config: map[string]cty.Value{
				"password":    cty.UnknownVal(cty.String),
				"password_wo": cty.StringVal("write-only-password"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := testPgUserRawConfig{value: cty.ObjectVal(tt.config)}
			err := validatePgUserPasswordConflict(config)
			if gotError := err != nil; gotError != tt.wantError {
				t.Fatalf("validatePgUserPasswordConflict() error = %v, wantError %t", err, tt.wantError)
			}
		})
	}
}

func TestValidatePgUserPasswordPair(t *testing.T) {
	tests := []struct {
		name      string
		config    cty.Value
		wantError bool
	}{
		{
			name: "write-only password with version",
			config: cty.ObjectVal(map[string]cty.Value{
				"password_wo":         cty.StringVal("write-only-password"),
				"password_wo_version": cty.NumberIntVal(1),
			}),
		},
		{
			name:   "neither write-only field",
			config: cty.EmptyObjectVal,
		},
		{
			name: "write-only password with null version at apply",
			config: cty.ObjectVal(map[string]cty.Value{
				"password_wo":         cty.StringVal("write-only-password"),
				"password_wo_version": cty.NullVal(cty.Number),
			}),
			wantError: true,
		},
		{
			name: "version with null write-only password at apply",
			config: cty.ObjectVal(map[string]cty.Value{
				"password_wo":         cty.NullVal(cty.String),
				"password_wo_version": cty.NumberIntVal(1),
			}),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := testPgUserRawConfig{value: tt.config}
			err := validatePgUserPasswordPair(config)
			if gotError := err != nil; gotError != tt.wantError {
				t.Fatalf("validatePgUserPasswordPair() error = %v, wantError %t", err, tt.wantError)
			}
		})
	}
}

// Test that a PostgreSQL User can be created with password_wo and updated by incrementing password_wo_version
func TestAccMDBPostgreSQLUserPasswordWo_TF1_12(t *testing.T) {
	t.Parallel()
	clusterName := acctest.RandomWithPrefix("tf-postgresql-user-pw-wo")
	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBPostgreSQLUserConfigPasswordWo(clusterName, "initialP@ssw0rd", 1, "USER_PASSWORD_ENCRYPTION_MD5"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(pgUserResourceNameAlice, "name", "alice"),
					resource.TestCheckResourceAttr(pgUserResourceNameAlice, "password_wo_version", "1"),
					resource.TestCheckNoResourceAttr(pgUserResourceNameAlice, "password_wo"),
					resource.TestCheckResourceAttr(pgUserResourceNameAlice, "user_password_encryption", "USER_PASSWORD_ENCRYPTION_MD5"),
				),
			},
			{
				Config:             testAccMDBPostgreSQLUserConfigPasswordWo(clusterName, "initialP@ssw0rd", 1, "USER_PASSWORD_ENCRYPTION_MD5"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				Config:      testAccMDBPostgreSQLUserConfigPasswordConflict(clusterName),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`only one of .password. or .password_wo. can be specified`),
			},
			{
				Config:      testAccMDBPostgreSQLUserConfigPasswordWoWithoutVersion(clusterName),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`all of .password_wo,password_wo_version. must be\s+specified`),
			},
			{
				Config:      testAccMDBPostgreSQLUserConfigPasswordWoVersionWithoutPassword(clusterName),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`all of .password_wo,password_wo_version. must be\s+specified`),
			},
			{
				Config: testAccMDBPostgreSQLUserConfigPasswordWo(clusterName, "rotatedP@ssw0rd", 2, "USER_PASSWORD_ENCRYPTION_SCRAM_SHA_256"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(pgUserResourceNameAlice, "name", "alice"),
					resource.TestCheckResourceAttr(pgUserResourceNameAlice, "password_wo_version", "2"),
					resource.TestCheckNoResourceAttr(pgUserResourceNameAlice, "password_wo"),
					resource.TestCheckResourceAttr(pgUserResourceNameAlice, "user_password_encryption", "USER_PASSWORD_ENCRYPTION_SCRAM_SHA_256"),
				),
			},
		},
	})
}

func testAccMDBPostgreSQLUserConfigPasswordWo(name, passwordWo string, passwordWoVersion int, passwordEncryption string) string {
	return testAccMDBPostgreSQLUserConfigStep0(name) + fmt.Sprintf(`
resource "yandex_mdb_postgresql_user" "alice" {
	cluster_id               = yandex_mdb_postgresql_cluster.foo.id
	name                     = "alice"
	password_wo              = "%s"
	password_wo_version      = %d
	user_password_encryption = "%s"
	login                    = true
	conn_limit               = 50
	}`, passwordWo, passwordWoVersion, passwordEncryption)
}

func testAccMDBPostgreSQLUserConfigPasswordConflict(name string) string {
	return testAccMDBPostgreSQLUserConfigStep0(name) + `
resource "yandex_mdb_postgresql_user" "alice" {
	cluster_id          = yandex_mdb_postgresql_cluster.foo.id
	name                = "alice"
	password            = "mysecureP@ssw0rd"
	password_wo         = "mysecureP@ssw0rd"
	password_wo_version = 1
	login               = true
}`
}

func testAccMDBPostgreSQLUserConfigPasswordWoWithoutVersion(name string) string {
	return testAccMDBPostgreSQLUserConfigStep0(name) + `
resource "yandex_mdb_postgresql_user" "alice" {
	cluster_id  = yandex_mdb_postgresql_cluster.foo.id
	name        = "alice"
	password_wo = "mysecureP@ssw0rd"
	login       = true
}`
}

func testAccMDBPostgreSQLUserConfigPasswordWoVersionWithoutPassword(name string) string {
	return testAccMDBPostgreSQLUserConfigStep0(name) + `
resource "yandex_mdb_postgresql_user" "alice" {
	cluster_id          = yandex_mdb_postgresql_cluster.foo.id
	name                = "alice"
	password_wo_version = 1
	login               = true
}`
}
