package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	sdkSchema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkTerraform "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/kafka/v1"
)

func TestMDBKafkaUserPasswordWoSchema(t *testing.T) {
	resourceSchema := sdkSchema.InternalMap(resourceYandexMDBKafkaUser().Schema)

	tests := []struct {
		name      string
		config    map[string]any
		wantError bool
	}{
		{
			name: "write-only password with version",
			config: map[string]any{
				"cluster_id":          "cluster-id",
				"name":                "alice",
				"password_wo":         "write-only-password",
				"password_wo_version": 1,
			},
		},
		{
			name: "write-only password without version",
			config: map[string]any{
				"cluster_id":  "cluster-id",
				"name":        "alice",
				"password_wo": "write-only-password",
			},
			wantError: true,
		},
		{
			name: "version without write-only password",
			config: map[string]any{
				"cluster_id":          "cluster-id",
				"name":                "alice",
				"password_wo_version": 1,
			},
			wantError: true,
		},
		{
			name: "legacy password",
			config: map[string]any{
				"cluster_id": "cluster-id",
				"name":       "alice",
				"password":   "legacy-password",
			},
		},
		{
			name: "password is required",
			config: map[string]any{
				"cluster_id": "cluster-id",
				"name":       "alice",
			},
			wantError: true,
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

type testKafkaUserRawConfig struct {
	value  cty.Value
	values map[string]any
}

func (c testKafkaUserRawConfig) GetRawConfig() cty.Value {
	return c.value
}

func (c testKafkaUserRawConfig) GetOk(key string) (any, bool) {
	value, ok := c.values[key]
	return value, ok
}

func TestValidateKafkaUserPasswordConflict(t *testing.T) {
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
			config := testKafkaUserRawConfig{value: cty.ObjectVal(tt.config)}
			err := validateKafkaUserPasswordConflict(config)
			if gotError := err != nil; gotError != tt.wantError {
				t.Fatalf("validateKafkaUserPasswordConflict() error = %v, wantError %t", err, tt.wantError)
			}
		})
	}
}

func TestValidateKafkaUserPasswordPair(t *testing.T) {
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
			config := testKafkaUserRawConfig{value: tt.config}
			err := validateKafkaUserPasswordPair(config)
			if gotError := err != nil; gotError != tt.wantError {
				t.Fatalf("validateKafkaUserPasswordPair() error = %v, wantError %t", err, tt.wantError)
			}
		})
	}
}

func TestKafkaUserPassword(t *testing.T) {
	legacy := testKafkaUserRawConfig{
		value:  cty.EmptyObjectVal,
		values: map[string]any{"password": "legacy-password"},
	}
	assert.Equal(t, "legacy-password", kafkaUserPassword(legacy))

	writeOnly := testKafkaUserRawConfig{
		value: cty.ObjectVal(map[string]cty.Value{
			"password_wo": cty.StringVal("write-only-password"),
		}),
		values: map[string]any{"password": "legacy-password"},
	}
	assert.Equal(t, "write-only-password", kafkaUserPassword(writeOnly))
}

func TestBuildKafkaUserSpecPasswordWo(t *testing.T) {
	resourceData := testResourceDataRawWithWriteOnly(t, resourceYandexMDBKafkaUser(), map[string]any{
		"cluster_id":          "cluster-id",
		"name":                "alice",
		"password_wo":         "write-only-password",
		"password_wo_version": 1,
	})

	userSpec, err := buildKafkaUserSpec(resourceData)

	assert.NoError(t, err)
	assert.Equal(t, "write-only-password", userSpec.Password)
}

func testResourceDataRawWithWriteOnly(t *testing.T, resource *sdkSchema.Resource, raw map[string]any) *sdkSchema.ResourceData {
	t.Helper()

	// TestResourceDataRaw does not populate InstanceDiff.RawConfig, where the SDK
	// exposes write-only values to resource callbacks during a real apply.
	resourceConfig := sdkTerraform.NewResourceConfigRaw(raw)
	resourceSchema := sdkSchema.InternalMap(resource.Schema)
	diff, err := resourceSchema.Diff(context.Background(), nil, resourceConfig, nil, nil, true)
	if err != nil {
		t.Fatalf("building resource diff: %v", err)
	}
	rawConfig, err := sdkSchema.JSONMapToStateValue(raw, resource.CoreConfigSchema())
	if err != nil {
		t.Fatalf("building raw config value: %v", err)
	}
	diff.RawConfig = rawConfig

	resourceData, err := resourceSchema.Data(nil, diff)
	if err != nil {
		t.Fatalf("building resource data: %v", err)
	}
	return resourceData
}

func TestKafkaUserPasswordWoUpdateMask(t *testing.T) {
	assert.Equal(t, "password", mdbKafkaUserUpdateFieldsMap["password_wo_version"])
}

func TestRedactKafkaUserRequests(t *testing.T) {
	const password = "user-secret"
	createRequest := &kafka.CreateUserRequest{
		ClusterId: "cluster-id",
		UserSpec: &kafka.UserSpec{
			Name:     "alice",
			Password: password,
		},
	}
	updateRequest := &kafka.UpdateUserRequest{
		ClusterId: "cluster-id",
		UserName:  "alice",
		Password:  password,
	}

	redactedCreate := redactKafkaUserCreateRequest(createRequest)
	redactedUpdate := redactKafkaUserUpdateRequest(updateRequest)

	assert.Equal(t, password, createRequest.UserSpec.Password)
	assert.Equal(t, password, updateRequest.Password)
	assert.Equal(t, redactedKafkaSecret, redactedCreate.UserSpec.Password)
	assert.Equal(t, redactedKafkaSecret, redactedUpdate.Password)
	assert.NotContains(t, fmt.Sprintf("%+v", redactedCreate), password)
	assert.NotContains(t, fmt.Sprintf("%+v", redactedUpdate), password)
	assert.NotSame(t, createRequest, redactedCreate)
	assert.NotSame(t, updateRequest, redactedUpdate)
}

func TestRedactKafkaClusterCreateRequest(t *testing.T) {
	const password = "inline-user-secret"
	request := &kafka.CreateClusterRequest{
		Name: "cluster",
		UserSpecs: []*kafka.UserSpec{
			{Name: "alice", Password: password},
		},
	}

	redacted := redactKafkaClusterCreateRequest(request)

	assert.Equal(t, password, request.UserSpecs[0].Password)
	assert.Equal(t, redactedKafkaSecret, redacted.UserSpecs[0].Password)
	assert.NotContains(t, fmt.Sprintf("%+v", redacted), password)
	assert.NotSame(t, request, redacted)
}

func TestKafkaUserDataSourceExcludesPasswordWo(t *testing.T) {
	dataSource := dataSourceYandexMDBKafkaUser()

	assert.NotContains(t, dataSource.Schema, "password_wo")
	assert.NotContains(t, dataSource.Schema, "password_wo_version")
}
