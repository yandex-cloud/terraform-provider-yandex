package yandex_resource_manager_folder_iam_policy_binding

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	sdk "github.com/yandex-cloud/go-sdk/v2"
	"github.com/yandex-cloud/go-sdk/v2/credentials"
	"github.com/yandex-cloud/go-sdk/v2/pkg/endpoints"
	"github.com/yandex-cloud/go-sdk/v2/pkg/options"
	"github.com/yandex-cloud/terraform-provider-yandex/internal/provider/framework/config"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

const testFolderID = "folder"

type folderPolicyAPI struct {
	mu       sync.Mutex
	bindings map[string]map[string]string
	writes   []string
}

func (api *folderPolicyAPI) invoke(_ context.Context, method string, req, reply any, _ *grpc.ClientConn, _ grpc.UnaryInvoker, _ ...grpc.CallOption) error {
	api.mu.Lock()
	defer api.mu.Unlock()

	if r, ok := req.(interface{ GetResourceId() string }); ok && r.GetResourceId() != testFolderID {
		return fmt.Errorf("%s: resource_id = %q, want %q", method, r.GetResourceId(), testFolderID)
	}

	var result proto.Message
	switch req := req.(type) {
	case *access.ListAccessPolicyBindingsRequest:
		page := &access.ListAccessPolicyBindingsResponse{}
		for id, params := range api.bindings {
			page.AccessPolicyBindings = append(page.AccessPolicyBindings, &access.AccessPolicyBinding{
				AccessPolicyTemplateId: id, Parameters: params,
			})
		}
		proto.Merge(reply.(*access.ListAccessPolicyBindingsResponse), page)
		return nil
	case *access.BindAccessPolicyRequest:
		binding := req.GetAccessPolicyBinding()
		api.writes = append(api.writes, "bind "+binding.GetAccessPolicyTemplateId())
		api.bindings[binding.GetAccessPolicyTemplateId()] = binding.GetParameters()
		result = &access.BindAccessPolicyResponse{}
	case *access.UpdateAccessPolicyBindingParametersRequest:
		binding := req.GetAccessPolicyBinding()
		api.writes = append(api.writes, "update "+binding.GetAccessPolicyTemplateId())
		api.bindings[binding.GetAccessPolicyTemplateId()] = binding.GetParameters()
		result = &access.UpdateAccessPolicyBindingParametersResponse{}
	case *access.UnbindAccessPolicyRequest:
		api.writes = append(api.writes, "unbind "+req.GetAccessPolicyTemplateId())
		delete(api.bindings, req.GetAccessPolicyTemplateId())
		result = &access.UnbindAccessPolicyResponse{}
	default:
		return fmt.Errorf("unexpected RPC %s", method)
	}

	response, err := anypb.New(result)
	if err != nil {
		return err
	}
	op := reply.(*operation.Operation)
	op.Done = true
	op.Result = &operation.Operation_Response{Response: response}
	return nil
}

func TestPolicyBindingPreservesOtherTemplates(t *testing.T) {
	for _, tc := range []struct {
		method    string
		wantWrite string
	}{
		{"create", "bind A"},
		{"update", "update A"},
		{"delete", "unbind A"},
	} {
		t.Run(tc.method, func(t *testing.T) {
			ctx := t.Context()
			external := map[string]string{"owner": "external"}
			api := &folderPolicyAPI{bindings: map[string]map[string]string{"B": external}}
			client, err := sdk.Build(ctx,
				options.WithCredentials(credentials.NoAuthentication()),
				options.WithEndpointsResolver(endpoints.NewSingleEndpointResolver("passthrough:///folder-test")),
				options.WithPlaintext(),
				options.WithCustomDialOptions(grpc.WithUnaryInterceptor(api.invoke)),
			)
			require.NoError(t, err)
			t.Cleanup(func() { _ = client.Shutdown(context.Background()) })
			cfg := &config.Config{SDKv2: client}

			var schema resource.SchemaResponse
			NewResource().Schema(ctx, resource.SchemaRequest{}, &schema)
			plan := tfsdk.Plan{Schema: schema.Schema, Raw: tftypes.NewValue(schema.Schema.Type().TerraformType(ctx), nil)}
			require.Empty(t, plan.SetAttribute(ctx, path.Root("folder_id"), testFolderID).Errors())
			require.Empty(t, plan.SetAttribute(ctx, path.Root("access_policy_template_id"), "A").Errors())
			params := map[string]string{"owner": "new"}
			require.Empty(t, plan.SetAttribute(ctx, path.Root("parameters"), params).Errors())
			prior := tfsdk.State{Schema: plan.Schema, Raw: plan.Raw}
			if tc.method != "create" {
				old := map[string]string{"owner": "old"}
				api.bindings["A"] = old
				require.Empty(t, prior.SetAttribute(ctx, path.Root("parameters"), old).Errors())
			}
			want := tfsdk.State{Schema: plan.Schema, Raw: plan.Raw}
			require.Empty(t, want.SetAttribute(ctx, path.Root("id"), testFolderID+"/A").Errors())

			apply := func(t *testing.T) {
				t.Helper()
				r := &IAMPolicyBindingUpdater{providerConfig: cfg}
				var state tfsdk.State
				var diags diag.Diagnostics
				switch tc.method {
				case "create":
					resp := resource.CreateResponse{State: tfsdk.State{
						Schema: plan.Schema, Raw: tftypes.NewValue(plan.Schema.Type().TerraformType(ctx), nil),
					}}
					r.Create(ctx, resource.CreateRequest{Plan: plan}, &resp)
					state, diags = resp.State, resp.Diagnostics
				case "update":
					resp := resource.UpdateResponse{State: prior}
					r.Update(ctx, resource.UpdateRequest{Plan: plan, State: prior}, &resp)
					state, diags = resp.State, resp.Diagnostics
				case "delete":
					resp := resource.DeleteResponse{State: prior}
					r.Delete(ctx, resource.DeleteRequest{State: prior}, &resp)
					state, diags = resp.State, resp.Diagnostics
				default:
					t.Fatalf("unknown method %q", tc.method)
				}
				require.Empty(t, diags.Errors())
				if tc.method == "delete" {
					require.True(t, state.Raw.IsNull())
				} else {
					require.True(t, want.Raw.Equal(state.Raw), "got %s, want %s", state.Raw, want.Raw)
				}
			}
			t.Run("concurrent requests", func(t *testing.T) {
				for range 2 {
					t.Run("apply", func(t *testing.T) {
						t.Parallel()
						apply(t)
					})
				}
			})
			apply(t)

			require.Equal(t, []string{tc.wantWrite}, api.writes)
			require.Equal(t, external, api.bindings["B"])
			if tc.method == "delete" {
				require.NotContains(t, api.bindings, "A")
			} else {
				require.Equal(t, params, api.bindings["A"])
			}
		})
	}
}
