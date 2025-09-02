package mdbcommon

import (
	"context"
	"math/rand"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func buildTestGreaterValidatorRequest(value int64) validator.Int64Request {
	reqConf := tfsdk.Config{
		Raw: tftypes.NewValue(
			tftypes.Object{}, map[string]tftypes.Value{
				"attr1": tftypes.NewValue(tftypes.Number, 5),
				"block1": tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{
					"attr3": tftypes.NewValue(tftypes.Number, 10),
					"attr4": tftypes.NewValue(tftypes.Number, 15),
				}),
			},
		),
		Schema: schema.Schema{
			Attributes: map[string]schema.Attribute{
				"attr1": schema.Int64Attribute{
					Required: true,
					Validators: []validator.Int64{
						Int64GreaterValidator(),
					},
				},
				"attr2": schema.Int64Attribute{
					Optional: true,
				},
			},
			Blocks: map[string]schema.Block{
				"block1": schema.SingleNestedBlock{
					Attributes: map[string]schema.Attribute{
						"attr3": schema.Int64Attribute{
							Required: true,
						},
						"attr4": schema.Int64Attribute{
							Required: true,
						},
					},
				},
			},
		},
	}

	return validator.Int64Request{
		Config:      reqConf,
		ConfigValue: types.Int64Value(value),
		Path:        path.Root("attr1"),
	}
}

func TestYandexProvider_Int64GreaterValidator(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname      string
		validator     *int64GreaterValidator
		req           validator.Int64Request
		expectedError bool
	}{
		{
			testname:      "CheckWithNullComparing",
			validator:     Int64GreaterValidator(path.MatchRoot("attr2")),
			req:           buildTestGreaterValidatorRequest(rand.Int63()),
			expectedError: false,
		},
		{
			testname:      "CheckWithInt64ComparingSuccess",
			validator:     Int64GreaterValidator(path.MatchRoot("block1").AtName("attr3")),
			req:           buildTestGreaterValidatorRequest(11),
			expectedError: false,
		},
		{
			testname:      "CheckWithInt64ComparingFailed",
			validator:     Int64GreaterValidator(path.MatchRoot("block1").AtName("attr3")),
			req:           buildTestGreaterValidatorRequest(5),
			expectedError: true,
		},
		{
			testname:      "CheckWithInt64SeveralComparingSuccess",
			validator:     Int64GreaterValidator(path.MatchRoot("block1").AtName("attr3"), path.MatchRoot("block1").AtName("attr4")),
			req:           buildTestGreaterValidatorRequest(20),
			expectedError: false,
		},
		{
			testname:      "CheckWithInt64SeveralComparingFailed",
			validator:     Int64GreaterValidator(path.MatchRoot("block1").AtName("attr3"), path.MatchRoot("block1").AtName("attr4")),
			req:           buildTestGreaterValidatorRequest(12),
			expectedError: true,
		},
	}

	for _, c := range cases {
		var resp validator.Int64Response
		c.validator.ValidateInt64(ctx, c.req, &resp)
		if resp.Diagnostics.HasError() != c.expectedError {
			t.Errorf(
				"Unexpected validation status %s test: expected %t, actual %t with errors: %v",
				c.testname,
				c.expectedError,
				resp.Diagnostics.HasError(),
				resp.Diagnostics.Errors(),
			)
		}
	}
}
