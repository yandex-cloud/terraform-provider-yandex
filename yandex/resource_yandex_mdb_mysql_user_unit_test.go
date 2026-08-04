package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	sdkSchema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkTerraform "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
)

func TestRedactMySQLUserCreateRequest(t *testing.T) {
	const password = "create-secret"
	request := &mysql.CreateUserRequest{
		ClusterId: "cluster-id",
		UserSpec: &mysql.UserSpec{
			Name:     "john",
			Password: password,
		},
	}

	redactedRequest := redactMySQLUserCreateRequest(request)
	loggedRequest := fmt.Sprintf("%+v", redactedRequest)

	assert.NotContains(t, loggedRequest, password)
	assert.Contains(t, loggedRequest, redactedMySQLUserPassword)
	assert.Contains(t, loggedRequest, "john")
	assert.Equal(t, password, request.UserSpec.Password)
	assert.NotSame(t, request, redactedRequest)
	assert.NotSame(t, request.UserSpec, redactedRequest.UserSpec)
}

func TestRedactMySQLUserUpdateRequest(t *testing.T) {
	const password = "update-secret"
	request := &mysql.UpdateUserRequest{
		ClusterId:         "cluster-id",
		UserName:          "john",
		Password:          password,
		GlobalPermissions: []mysql.GlobalPermission{mysql.GlobalPermission_PROCESS},
	}

	redactedRequest := redactMySQLUserUpdateRequest(request)
	loggedRequest := fmt.Sprintf("%+v", redactedRequest)

	assert.NotContains(t, loggedRequest, password)
	assert.Contains(t, loggedRequest, redactedMySQLUserPassword)
	assert.Contains(t, loggedRequest, "john")
	assert.Equal(t, password, request.Password)
	assert.NotSame(t, request, redactedRequest)
}

func TestMDBMySQLUserPasswordWoRequiredWith(t *testing.T) {
	resourceSchema := sdkSchema.InternalMap(resourceYandexMDBMySQLUser().Schema)

	tests := []struct {
		name      string
		config    map[string]interface{}
		wantError bool
	}{
		{
			name: "write-only password with version",
			config: map[string]interface{}{
				"cluster_id":          "cluster-id",
				"name":                "john",
				"password_wo":         "mysecureP@ssw0rd",
				"password_wo_version": 1,
			},
		},
		{
			name: "write-only password without version",
			config: map[string]interface{}{
				"cluster_id":  "cluster-id",
				"name":        "john",
				"password_wo": "mysecureP@ssw0rd",
			},
			wantError: true,
		},
		{
			name: "version without write-only password",
			config: map[string]interface{}{
				"cluster_id":          "cluster-id",
				"name":                "john",
				"password_wo_version": 1,
			},
			wantError: true,
		},
		{
			name: "legacy password",
			config: map[string]interface{}{
				"cluster_id": "cluster-id",
				"name":       "john",
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

type testMySQLUserRawConfig struct {
	value cty.Value
}

func (c testMySQLUserRawConfig) GetRawConfig() cty.Value {
	return c.value
}

func TestValidateMySQLUserPasswordConflict(t *testing.T) {
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
			config := testMySQLUserRawConfig{value: cty.ObjectVal(tt.config)}
			err := validateMySQLUserPasswordConflict(config)
			if gotError := err != nil; gotError != tt.wantError {
				t.Fatalf("validateMySQLUserPasswordConflict() error = %v, wantError %t", err, tt.wantError)
			}
		})
	}
}

func TestValidateMySQLUserPasswordPair(t *testing.T) {
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
			config := testMySQLUserRawConfig{value: tt.config}
			err := validateMySQLUserPasswordPair(config)
			if gotError := err != nil; gotError != tt.wantError {
				t.Fatalf("validateMySQLUserPasswordPair() error = %v, wantError %t", err, tt.wantError)
			}
		})
	}
}
