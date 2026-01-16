package cdn_resource

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/cdn/v1"
)

func TestFlattenHostOptions(t *testing.T) {
	tests := []struct {
		name                string
		input               *cdn.ResourceOptions_HostOptions
		planOptions         *CDNOptionsModel
		expectedForwardHost bool
		expectedCustomHost  string
		expectedForwardNull bool
		expectedCustomNull  bool
	}{
		{
			name:                "nil input",
			input:               nil,
			planOptions:         nil,
			expectedForwardNull: true,
			expectedCustomNull:  true,
		},
		{
			name: "forward host header true",
			input: &cdn.ResourceOptions_HostOptions{
				HostVariant: &cdn.ResourceOptions_HostOptions_ForwardHostHeader{
					ForwardHostHeader: &cdn.ResourceOptions_BoolOption{
						Enabled: true,
						Value:   true,
					},
				},
			},
			planOptions:         nil,
			expectedForwardHost: true,
			expectedForwardNull: false,
			expectedCustomNull:  true,
		},
		{
			name: "forward host header false",
			input: &cdn.ResourceOptions_HostOptions{
				HostVariant: &cdn.ResourceOptions_HostOptions_ForwardHostHeader{
					ForwardHostHeader: &cdn.ResourceOptions_BoolOption{
						Enabled: true,
						Value:   false,
					},
				},
			},
			planOptions:         nil,
			expectedForwardHost: false,
			expectedForwardNull: false,
			expectedCustomNull:  true,
		},
		{
			name: "custom host header set",
			input: &cdn.ResourceOptions_HostOptions{
				HostVariant: &cdn.ResourceOptions_HostOptions_Host{
					Host: &cdn.ResourceOptions_StringOption{
						Enabled: true,
						Value:   "example.com",
					},
				},
			},
			planOptions:         nil,
			expectedCustomHost:  "example.com",
			expectedCustomNull:  false,
			expectedForwardNull: true,
		},
		// New test cases for plan preservation
		{
			name: "custom host header active, preserve forward=false from plan",
			input: &cdn.ResourceOptions_HostOptions{
				HostVariant: &cdn.ResourceOptions_HostOptions_Host{
					Host: &cdn.ResourceOptions_StringOption{
						Enabled: true,
						Value:   "video.avs.io",
					},
				},
			},
			planOptions: &CDNOptionsModel{
				ForwardHostHeader: types.BoolValue(false),
			},
			expectedCustomHost:  "video.avs.io",
			expectedCustomNull:  false,
			expectedForwardHost: false, // Should be false, NOT null
			expectedForwardNull: false,
		},
		{
			name: "forward host header active, preserve custom='' from plan",
			input: &cdn.ResourceOptions_HostOptions{
				HostVariant: &cdn.ResourceOptions_HostOptions_ForwardHostHeader{
					ForwardHostHeader: &cdn.ResourceOptions_BoolOption{
						Enabled: true,
						Value:   true,
					},
				},
			},
			planOptions: &CDNOptionsModel{
				CustomHostHeader: types.StringValue(""),
			},
			expectedForwardHost: true,
			expectedForwardNull: false,
			expectedCustomHost:  "",
			expectedCustomNull:  false, // Should be empty string, NOT null
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			opt := &CDNOptionsModel{}
			flattenHostOptions(tt.input, opt, tt.planOptions, &diags)

			if tt.expectedForwardNull {
				assert.True(t, opt.ForwardHostHeader.IsNull(), "ForwardHostHeader should be null")
			} else {
				assert.False(t, opt.ForwardHostHeader.IsNull(), "ForwardHostHeader should not be null")
				assert.Equal(t, tt.expectedForwardHost, opt.ForwardHostHeader.ValueBool(), "ForwardHostHeader value mismatch")
			}

			if tt.expectedCustomNull {
				assert.True(t, opt.CustomHostHeader.IsNull(), "CustomHostHeader should be null")
			} else {
				assert.False(t, opt.CustomHostHeader.IsNull(), "CustomHostHeader should not be null")
				assert.Equal(t, tt.expectedCustomHost, opt.CustomHostHeader.ValueString(), "CustomHostHeader value mismatch")
			}
		})
	}
}
