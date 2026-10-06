package yandex_cloudregistry_scan_policy

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cloudregistry "github.com/yandex-cloud/go-genproto/yandex/cloud/cloudregistry/v1"
	operation "github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	sdk "github.com/yandex-cloud/go-sdk/v2"
	"github.com/yandex-cloud/go-sdk/v2/credentials"
	"github.com/yandex-cloud/go-sdk/v2/pkg/endpoints"
	"github.com/yandex-cloud/go-sdk/v2/pkg/options"
	config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type scanPolicyTestListener struct {
	incoming chan net.Conn
	closed   chan struct{}
	once     sync.Once
}

func (l *scanPolicyTestListener) Accept() (net.Conn, error) {
	select {
	case c := <-l.incoming:
		return c, nil
	case <-l.closed:
		return nil, net.ErrClosed
	}
}
func (l *scanPolicyTestListener) Close() error   { l.once.Do(func() { close(l.closed) }); return nil }
func (l *scanPolicyTestListener) Addr() net.Addr { return &net.TCPAddr{} }
func (l *scanPolicyTestListener) DialContext(ctx context.Context) (net.Conn, error) {
	client, server := net.Pipe()
	select {
	case l.incoming <- server:
		return client, nil
	case <-ctx.Done():
		_ = client.Close()
		_ = server.Close()
		return nil, ctx.Err()
	case <-l.closed:
		_ = client.Close()
		_ = server.Close()
		return nil, net.ErrClosed
	}
}

type scanPolicyTestAPI struct {
	cloudregistry.UnimplementedScanPolicyServiceServer
	getRules      *cloudregistry.ScanRules
	updateRequest *cloudregistry.UpdateScanPolicyRequest
	getCalls      int
}

func scanPolicyTestPolicy(rules *cloudregistry.ScanRules) *cloudregistry.ScanPolicy {
	return &cloudregistry.ScanPolicy{Id: "test-policy", RegistryId: "test-registry", Name: "scan-policy-test", Description: "scan policy test", Rules: rules, CreatedAt: timestamppb.New(time.Unix(1700000000, 0))}
}

func scanPolicyTestRules(paths []string) *cloudregistry.ScanRules {
	return &cloudregistry.ScanRules{PushRule: &cloudregistry.PushRule{Paths: paths}}
}

func (s *scanPolicyTestAPI) Get(_ context.Context, req *cloudregistry.GetScanPolicyRequest) (*cloudregistry.ScanPolicy, error) {
	if req.ScanPolicyId != "test-policy" {
		return nil, fmt.Errorf("unexpected ID: %q", req.ScanPolicyId)
	}
	s.getCalls++
	return scanPolicyTestPolicy(s.getRules), nil
}

func (s *scanPolicyTestAPI) completed(metadata proto.Message) (*operation.Operation, error) {
	md, err := anypb.New(metadata)
	if err != nil {
		return nil, err
	}
	value, err := anypb.New(scanPolicyTestPolicy(nil))
	if err != nil {
		return nil, err
	}
	return &operation.Operation{Id: "test-operation", Done: true, Metadata: md, Result: &operation.Operation_Response{Response: value}}, nil
}

func (s *scanPolicyTestAPI) Update(_ context.Context, req *cloudregistry.UpdateScanPolicyRequest) (*operation.Operation, error) {
	s.updateRequest = proto.Clone(req).(*cloudregistry.UpdateScanPolicyRequest)
	for _, field := range req.GetUpdateMask().GetPaths() {
		switch field {
		case "rules":
			s.getRules = proto.Clone(req.GetRules()).(*cloudregistry.ScanRules)
		case "rules.push_rule":
			s.getRules.PushRule = proto.Clone(req.GetRules().GetPushRule()).(*cloudregistry.PushRule)
		case "rules.schedule_rules":
			s.getRules.ScheduleRules = proto.Clone(req.GetRules()).(*cloudregistry.ScanRules).ScheduleRules
		}
	}
	return s.completed(&cloudregistry.UpdateScanPolicyMetadata{ScanPolicyId: "test-policy"})
}

func scanPolicyTestResource(t *testing.T, ctx context.Context, api *scanPolicyTestAPI) *yandexCloudregistryScanPolicyResource {
	t.Helper()
	listener := &scanPolicyTestListener{incoming: make(chan net.Conn), closed: make(chan struct{})}
	server := grpc.NewServer()
	cloudregistry.RegisterScanPolicyServiceServer(server, api)
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(func() { server.Stop(); _ = listener.Close() })
	client, err := sdk.Build(ctx,
		options.WithCredentials(credentials.NoAuthentication()),
		options.WithEndpointsResolver(endpoints.NewSingleEndpointResolver("passthrough:///scan-policy-test")),
		options.WithPlaintext(),
		options.WithCustomDialOptions(grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) })),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Shutdown(context.Background()) })
	return &yandexCloudregistryScanPolicyResource{providerConfig: &config.Config{SDKv2: client}}
}

func scanPolicyTestModel(t *testing.T, ctx context.Context, paths []string) yandexCloudregistryScanPolicyModel {
	t.Helper()
	model := NewYandexCloudregistryScanPolicyModel()
	model.ID = types.StringValue("test-policy")
	model.ScanPolicyId = model.ID
	model.RegistryId = types.StringValue("test-registry")
	model.Name = types.StringValue("scan-policy-test")
	model.Description = types.StringValue("scan policy test")
	model.Disabled = types.BoolValue(false)
	model.ScanLangPackages = types.BoolValue(false)
	model.Timeouts = timeouts.Value{Object: types.ObjectNull(map[string]attr.Type{
		"create": types.StringType, "read": types.StringType, "update": types.StringType, "delete": types.StringType,
	})}
	var diagnostics diag.Diagnostics
	model.Rules = flattenYandexCloudregistryScanPolicyRules(ctx, scanPolicyTestRules(paths), NewYandexCloudregistryScanPolicyRulesModel(), &diagnostics)
	if diagnostics.HasError() {
		t.Fatal(diagnostics)
	}
	return model
}

func TestScanPolicyUpdateRules(t *testing.T) {
	for _, tc := range []struct {
		name                         string
		before                       []string
		oldDisabled, newDisabled     bool
		descriptionOnly              bool
		withSchedule, changeSchedule bool
	}{
		{name: "null_to_star", before: nil},
		{name: "paths_change_preserves_schedule", before: []string{"tf-path-probe"}, withSchedule: true},
		{name: "schedule_change_preserves_paths", before: []string{"*"}, withSchedule: true, changeSchedule: true},
		{name: "old_path_to_star", before: []string{"tf-path-probe"}},
		{name: "path_change_preserves_disabled", before: []string{"tf-path-probe"}, oldDisabled: true, newDisabled: true},
		{name: "disabled_change_preserves_paths", before: []string{"*"}, newDisabled: true},
		{name: "description_does_not_update_rules", before: []string{"*"}, descriptionOnly: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			oldRules := scanPolicyTestRules(tc.before)
			oldRules.PushRule.Disabled = tc.oldDisabled
			newRules := scanPolicyTestRules([]string{"*"})
			newRules.PushRule.Disabled = tc.newDisabled
			if tc.withSchedule {
				oldRules.ScheduleRules = []*cloudregistry.ScheduledRule{{Amount: 7, IntervalUnit: cloudregistry.ScheduledRule_DAYS, Paths: []string{"team"}, Disabled: true}}
				newRules.ScheduleRules = []*cloudregistry.ScheduledRule{proto.Clone(oldRules.ScheduleRules[0]).(*cloudregistry.ScheduledRule)}
				if tc.changeSchedule {
					newRules.ScheduleRules[0].Amount = 14
				}
			}
			api := &scanPolicyTestAPI{getRules: oldRules}
			r := scanPolicyTestResource(t, ctx, api)
			schema := YandexCloudregistryScanPolicyResourceSchema(ctx)
			oldModel := scanPolicyTestModel(t, ctx, tc.before)
			planModel := scanPolicyTestModel(t, ctx, []string{"*"})
			var diagnostics diag.Diagnostics
			oldModel.Rules = flattenYandexCloudregistryScanPolicyRules(ctx, oldRules, NewYandexCloudregistryScanPolicyRulesModel(), &diagnostics)
			planModel.Rules = flattenYandexCloudregistryScanPolicyRules(ctx, newRules, NewYandexCloudregistryScanPolicyRulesModel(), &diagnostics)
			if diagnostics.HasError() {
				t.Fatal(diagnostics)
			}
			expectedMask := []string{"rules"}
			if tc.descriptionOnly {
				oldModel.Description = types.StringValue("before")
				expectedMask = []string{"description"}
			}
			state := tfsdk.State{Schema: schema}
			if d := state.Set(ctx, oldModel); d.HasError() {
				t.Fatal(d)
			}
			plan := tfsdk.Plan{Schema: schema}
			if d := plan.Set(ctx, planModel); d.HasError() {
				t.Fatal(d)
			}
			response := resource.UpdateResponse{State: tfsdk.State{Schema: schema, Raw: plan.Raw}}
			r.Update(ctx, resource.UpdateRequest{Plan: plan, State: state}, &response)
			if response.Diagnostics.HasError() {
				t.Fatal(response.Diagnostics)
			}
			if api.updateRequest == nil {
				t.Fatal("Update was not called")
			}
			if !reflect.DeepEqual(api.updateRequest.GetUpdateMask().GetPaths(), expectedMask) {
				t.Errorf("update mask = %v, want %v", api.updateRequest.GetUpdateMask().GetPaths(), expectedMask)
			}
			if !proto.Equal(api.updateRequest.GetRules(), newRules) {
				t.Errorf("request rules = %v, want %v", api.updateRequest.GetRules(), newRules)
			}
			if api.getCalls != 1 {
				t.Fatalf("Get calls = %d, want 1", api.getCalls)
			}
			var actual types.Object
			if d := response.State.GetAttribute(ctx, path.Root("rules"), &actual); d.HasError() {
				t.Fatal(d)
			}
			if !actual.Equal(planModel.Rules) {
				t.Errorf("state rules = %s, plan rules = %s", actual.String(), planModel.Rules.String())
			}
			t.Logf("mask = %v; state matches planned rules = %t", api.updateRequest.GetUpdateMask().GetPaths(), actual.Equal(planModel.Rules))
		})
	}
}
