package sdkcommon_test

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	operationpb "github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	sdkoperationv2 "github.com/yandex-cloud/go-sdk/v2/pkg/operation"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestJoinedStrings(t *testing.T) {
	testKeys := []string{"key1", "key2", "key3"}
	joinedKeys := sdkcommon.GetJoinedKeys(testKeys)
	assert.Equal(t, "`key1`, `key2`, `key3`", joinedKeys)

	testKey := []string{"key1"}
	joinedKey := sdkcommon.GetJoinedKeys(testKey)
	assert.Equal(t, "`key1`", joinedKey)
}

func TestParseDuration(t *testing.T) {
	d := (&schema.Resource{
		//lintignore:XS001
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}).TestResourceData()
	i := d.Get("name")
	r, err := sdkcommon.ParseDuration(i.(string))
	require.NoError(t, err)
	require.Nil(t, r)
}

func TestRemoveWriteOnlyFields(t *testing.T) {
	resourceSchema := map[string]*schema.Schema{
		"password": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Password.",
		},
		"password_wo": {
			Type:         schema.TypeString,
			Optional:     true,
			WriteOnly:    true,
			RequiredWith: []string{"password_wo_version"},
			Description:  "Write-only password.",
		},
		"password_wo_version": {
			Type:         schema.TypeInt,
			Optional:     true,
			RequiredWith: []string{"password_wo"},
			Description:  "Version of the write-only password.",
		},
		"api_key_wo": {
			Type:        schema.TypeString,
			Optional:    true,
			WriteOnly:   true,
			Description: "Write-only API key without declarative pairing.",
		},
		"api_key_wo_version": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Version of the write-only API key.",
		},
		"config": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Configuration.",
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"token_wo": {
					Type:         schema.TypeString,
					Optional:     true,
					WriteOnly:    true,
					RequiredWith: []string{"config.0.token_wo_version"},
					Description:  "Write-only token.",
				},
				"token_wo_version": {
					Type:         schema.TypeInt,
					Optional:     true,
					RequiredWith: []string{"config.0.token_wo"},
					Description:  "Version of the write-only token.",
				},
			}},
		},
	}

	sdkcommon.RemoveWriteOnlyFields(resourceSchema)

	require.Contains(t, resourceSchema, "password")
	require.NotContains(t, resourceSchema, "password_wo")
	require.NotContains(t, resourceSchema, "password_wo_version")
	require.NotContains(t, resourceSchema, "api_key_wo")
	require.NotContains(t, resourceSchema, "api_key_wo_version")

	nestedSchema := resourceSchema["config"].Elem.(*schema.Resource).Schema
	require.NotContains(t, nestedSchema, "token_wo")
	require.NotContains(t, nestedSchema, "token_wo_version")
}

func TestWaitConflictingOperationV2PreservesSDKV1ErrorContract(t *testing.T) {
	t.Run("initial get error is returned", func(t *testing.T) {
		wantErr := status.Error(codes.PermissionDenied, "cannot get operation")
		err := sdkcommon.WaitConflictingOperationV2(context.Background(), "test-operation", func(context.Context, string, ...grpc.CallOption) (sdkoperationv2.YCOperation, error) {
			return nil, wantErr
		}, func(int) time.Duration { return 0 })
		require.ErrorIs(t, err, wantErr)
	})

	t.Run("wait errors are ignored after initial get", func(t *testing.T) {
		polls := 0
		err := sdkcommon.WaitConflictingOperationV2(context.Background(), "test-operation", func(context.Context, string, ...grpc.CallOption) (sdkoperationv2.YCOperation, error) {
			polls++
			switch {
			case polls == 1:
				return &operationpb.Operation{Id: "test-operation"}, nil
			case polls <= 4:
				return nil, status.Error(codes.NotFound, "operation is not visible yet")
			default:
				return &operationpb.Operation{
					Id:   "test-operation",
					Done: true,
					Result: &operationpb.Operation_Error{
						Error: status.New(codes.Internal, "conflicting operation failed").Proto(),
					},
				}, nil
			}
		}, func(int) time.Duration { return 0 })
		require.NoError(t, err)
		require.Equal(t, 5, polls)
	})
}
