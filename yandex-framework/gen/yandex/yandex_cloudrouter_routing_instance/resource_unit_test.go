package yandex_cloudrouter_routing_instance_test

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	frameworkprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	cloudrouter "github.com/yandex-cloud/go-genproto/yandex/cloud/cloudrouter/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	ycsdk "github.com/yandex-cloud/go-sdk/v2"
	"github.com/yandex-cloud/go-sdk/v2/credentials"
	"github.com/yandex-cloud/go-sdk/v2/pkg/endpoints"
	"github.com/yandex-cloud/go-sdk/v2/pkg/options"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/gen/yandex/yandex_cloudrouter_routing_instance"
	yandexframework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const routingInstanceResourceName = "yandex_cloudrouter_routing_instance.test"

func TestCloudRouterRoutingInstance_lifecycle(t *testing.T) {
	api := &routingInstanceAPIStub{}
	initial := routingInstanceTestConfig("router-initial", `
  description = "initial description"
  labels = { stage = "initial" }
`+routingInstanceTestTopology)
	reordered := routingInstanceTestConfig("router-initial", `
  description = "initial description"
  labels = { stage = "initial" }
`+routingInstanceTestReorderedTopology)
	metadataUpdated := routingInstanceTestConfig("router-updated", `
  description = ""
  labels = {}
`+routingInstanceTestTopology)
	topologyUpdated := routingInstanceTestConfig("router-updated", routingInstanceTestUpdatedTopology)
	cleared := routingInstanceTestConfig("router-updated", `
  vpc_info = []
  cic_private_connection_info = []
`)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: routingInstanceTestFactories(t, api),
		CheckDestroy:             api.checkDestroyed,
		Steps: []resource.TestStep{
			{
				Config: initial,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(routingInstanceResourceName, "folder_id", "folder-id"),
					resource.TestCheckResourceAttr(routingInstanceResourceName, "region_id", "ru-central1"),
					resource.TestCheckResourceAttr(routingInstanceResourceName, "status", "ACTIVE"),
					resource.TestCheckResourceAttr(routingInstanceResourceName, "created_at", "2026-09-11T00:00:00Z"),
					resource.TestCheckResourceAttr(routingInstanceResourceName, "vpc_info.#", "2"),
					resource.TestCheckResourceAttr(routingInstanceResourceName, "cic_private_connection_info.#", "2"),
					api.checkMutations(1, 0),
				),
			},
			{ResourceName: routingInstanceResourceName, ImportState: true, ImportStateVerify: true},
			{Config: reordered, PlanOnly: true},
			{Config: routingInstanceTestConfig("router-initial", ""), PlanOnly: true},
			{
				Config: metadataUpdated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(routingInstanceResourceName, "description", ""),
					resource.TestCheckResourceAttr(routingInstanceResourceName, "labels.%", "0"),
					resource.TestCheckResourceAttr(routingInstanceResourceName, "vpc_info.#", "2"),
					api.checkUpdateMask("description", "labels", "name"),
					api.checkMutations(1, 1),
				),
			},
			{
				Config: topologyUpdated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(routingInstanceResourceName, "vpc_info.#", "1"),
					resource.TestCheckResourceAttr(routingInstanceResourceName, "cic_private_connection_info.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(routingInstanceResourceName, "vpc_info.*", map[string]string{
						"vpc_network_id": "network-a",
						"az_infos.#":     "1",
					}),
					api.checkUpdateMask("cic_private_connection_info", "vpc_info"),
					api.checkMutations(1, 2),
				),
			},
			{
				Config: cleared,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(routingInstanceResourceName, "vpc_info.#", "0"),
					resource.TestCheckResourceAttr(routingInstanceResourceName, "cic_private_connection_info.#", "0"),
					api.checkUpdateMask("cic_private_connection_info", "vpc_info"),
					api.checkMutations(1, 3),
				),
			},
			{
				Config: routingInstanceTestConfig("router-updated", `
  folder_id = "other-folder"
  vpc_info = []
  cic_private_connection_info = []
`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(routingInstanceResourceName, "folder_id", "other-folder"),
					resource.TestCheckResourceAttr(routingInstanceResourceName, "id", "routing-instance-1"),
					api.checkMutations(1, 3),
				),
			},
			{
				PreConfig: func() {
					api.mu.Lock()
					defer api.mu.Unlock()
					api.instance = nil
				},
				Config: cleared,
				Check:  api.checkMutations(2, 3),
			},
		},
	})
}

func TestCloudRouterRoutingInstance_apiOrder(t *testing.T) {
	api := &routingInstanceAPIStub{}
	config := routingInstanceTestConfig("router-order", `
  vpc_info = [
    {
      vpc_network_id = "network-a"
      az_infos = [
        { manual_info = { az_id = "ru-central1-a", prefixes = [] } },
        { manual_info = { az_id = "ru-central1-b" } }
      ]
    },
    {
      vpc_network_id = "network-b"
      az_infos = [
        { manual_info = { az_id = "ru-central1-a" } },
        { manual_info = { az_id = "ru-central1-b", prefixes = [] } }
      ]
    }
  ]
`)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: routingInstanceTestFactories(t, api),
		CheckDestroy:             api.checkDestroyed,
		Steps: []resource.TestStep{
			{Config: config, Check: api.checkMutations(1, 0)},
			{
				PreConfig: func() {
					api.mu.Lock()
					defer api.mu.Unlock()
					api.reverseRead = true
				},
				Config:   config,
				PlanOnly: true,
			},
			{ResourceName: routingInstanceResourceName, ImportState: true, ImportStateVerify: true},
		},
	})
}

func TestCloudRouterRoutingInstance_unknownNetwork(t *testing.T) {
	api := &routingInstanceAPIStub{}
	config := func(networkID string) string {
		return fmt.Sprintf(`
resource "terraform_data" "network" {
  input = %q
}
`, networkID) + routingInstanceTestConfig("router-test", `
  vpc_info = [{
    vpc_network_id = terraform_data.network.output
    az_infos = [{ manual_info = { az_id = "ru-central1-a", prefixes = ["10.0.0.0/24"] } }]
  }]
`)
	}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: routingInstanceTestFactories(t, api),
		CheckDestroy:             api.checkDestroyed,
		Steps: []resource.TestStep{
			{
				Config: config("network-id"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(routingInstanceResourceName, "vpc_info.*", map[string]string{
						"vpc_network_id": "network-id",
					}),
					api.checkMutations(1, 0),
				),
			},
			{Config: config("network-id"), PlanOnly: true},
			{
				Config: config("other-network-id"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(routingInstanceResourceName, "vpc_info.*", map[string]string{
						"vpc_network_id": "other-network-id",
					}),
					api.checkUpdateMask("vpc_info"),
					api.checkMutations(1, 1),
				),
			},
			{Config: config("other-network-id"), PlanOnly: true},
		},
	})
}

func TestCloudRouterRoutingInstance_createError(t *testing.T) {
	api := &routingInstanceAPIStub{createError: status.Error(codes.PermissionDenied, "create denied")}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: routingInstanceTestFactories(t, api),
		Steps: []resource.TestStep{{
			Config:      routingInstanceTestConfig("router-test", ""),
			ExpectError: regexp.MustCompile("create denied"),
		}},
	})
}

func TestCloudRouterRoutingInstance_computedIDs(t *testing.T) {
	for _, attribute := range []string{"id", "routing_instance_id"} {
		t.Run(attribute, func(t *testing.T) {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: routingInstanceTestFactories(t, &routingInstanceAPIStub{}),
				Steps: []resource.TestStep{{
					Config:      routingInstanceTestConfig("router-test", attribute+` = "configured-id"`),
					ExpectError: regexp.MustCompile("Invalid Configuration for Read-Only Attribute"),
				}},
			})
		})
	}
}

func TestCloudRouterRoutingInstance_topologyUpdates(t *testing.T) {
	api := &routingInstanceAPIStub{}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: routingInstanceTestFactories(t, api),
		CheckDestroy:             api.checkDestroyed,
		Steps: []resource.TestStep{
			{Config: routingInstanceTestConfig("router-test", routingInstanceTestTopology)},
			{
				Config: routingInstanceTestConfig("router-test", routingInstanceTestUpdatedNetworks+`
  cic_private_connection_info = [
    { cic_private_connection_id = "private-connection-a" },
    { cic_private_connection_id = "private-connection-b" }
  ]
`),
				Check: resource.ComposeTestCheckFunc(
					api.checkUpdateMask("vpc_info"),
					api.checkMutations(1, 1),
				),
			},
			{
				Config: routingInstanceTestConfig("router-test", routingInstanceTestUpdatedTopology),
				Check: resource.ComposeTestCheckFunc(
					api.checkUpdateMask("cic_private_connection_info"),
					api.checkMutations(1, 2),
				),
			},
			{
				Config: routingInstanceTestConfig("router-test", `
  vpc_info = null
  cic_private_connection_info = null
`),
				PlanOnly: true,
			},
		},
	})
}

func TestCloudRouterRoutingInstance_moveError(t *testing.T) {
	for _, failure := range []string{"rpc", "operation"} {
		t.Run(failure, func(t *testing.T) {
			api := &routingInstanceAPIStub{}
			moved := routingInstanceTestConfig("router-test", `folder_id = "other-folder"`)
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: routingInstanceTestFactories(t, api),
				CheckDestroy:             api.checkDestroyed,
				Steps: []resource.TestStep{
					{Config: routingInstanceTestConfig("router-test", "")},
					{
						PreConfig: func() {
							api.mu.Lock()
							defer api.mu.Unlock()
							if failure == "rpc" {
								api.moveError = status.Error(codes.PermissionDenied, "move denied")
							} else {
								api.moveOperationError = status.Error(codes.FailedPrecondition, "move denied")
							}
						},
						Config:      moved,
						ExpectError: regexp.MustCompile("move denied"),
					},
					{
						PreConfig: func() {
							api.mu.Lock()
							defer api.mu.Unlock()
							api.moveError = nil
							api.moveOperationError = nil
						},
						Config: moved,
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(routingInstanceResourceName, "id", "routing-instance-1"),
							resource.TestCheckResourceAttr(routingInstanceResourceName, "folder_id", "other-folder"),
							api.checkMutations(1, 0),
						),
					},
				},
			})
		})
	}
}

func TestCloudRouterRoutingInstance_updateRecovery(t *testing.T) {
	api := &routingInstanceAPIStub{}
	updated := routingInstanceTestConfig("router-updated", "")
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: routingInstanceTestFactories(t, api),
		CheckDestroy:             api.checkDestroyed,
		Steps: []resource.TestStep{
			{Config: routingInstanceTestConfig("router-initial", "")},
			{
				PreConfig: func() {
					api.mu.Lock()
					defer api.mu.Unlock()
					api.updateOperationError = status.Error(codes.Internal, "update failed after remote change")
				},
				Config:      updated,
				ExpectError: regexp.MustCompile("update failed after remote change"),
			},
			{
				PreConfig: func() {
					api.mu.Lock()
					defer api.mu.Unlock()
					api.updateOperationError = nil
				},
				Config: updated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(routingInstanceResourceName, "id", "routing-instance-1"),
					resource.TestCheckResourceAttr(routingInstanceResourceName, "name", "router-updated"),
					api.checkUpdateMask("name"),
					api.checkMutations(1, 1),
				),
			},
		},
	})
}

func TestCloudRouterRoutingInstance_deletionProtection(t *testing.T) {
	api := &routingInstanceAPIStub{}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: routingInstanceTestFactories(t, api),
		CheckDestroy:             api.checkDestroyed,
		Steps: []resource.TestStep{
			{
				Config: routingInstanceTestConfig("router-protected", "deletion_protection = true"),
				Check:  resource.TestCheckResourceAttr(routingInstanceResourceName, "deletion_protection", "true"),
			},
			{
				Config:      routingInstanceTestConfig("router-protected", "deletion_protection = true"),
				Destroy:     true,
				ExpectError: regexp.MustCompile("deletion protection"),
			},
			{
				Config:   routingInstanceTestConfig("router-protected", ""),
				PlanOnly: true,
			},
			{
				Config: routingInstanceTestConfig("router-protected", "deletion_protection = false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(routingInstanceResourceName, "deletion_protection", "false"),
					api.checkUpdateMask("deletion_protection"),
				),
			},
		},
	})
}

type routingInstanceTestProvider struct {
	frameworkprovider.Provider
	config *providerconfig.Config
}

func (p *routingInstanceTestProvider) Configure(_ context.Context, _ frameworkprovider.ConfigureRequest, resp *frameworkprovider.ConfigureResponse) {
	resp.ResourceData = p.config
	resp.DataSourceData = p.config
}

func (p *routingInstanceTestProvider) Resources(_ context.Context) []func() frameworkresource.Resource {
	return []func() frameworkresource.Resource{yandex_cloudrouter_routing_instance.NewResource}
}

func (p *routingInstanceTestProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{yandex_cloudrouter_routing_instance.NewDataSource}
}

func routingInstanceTestFactories(t *testing.T, api *routingInstanceAPIStub) map[string]func() (tfprotov6.ProviderServer, error) {
	t.Helper()
	sdk, err := ycsdk.Build(context.Background(),
		options.WithCredentials(credentials.NoAuthentication()),
		options.WithEndpointsResolver(endpoints.NewSingleEndpointResolver("unused:443")),
		options.WithPlaintext(),
		options.WithCustomDialOptions(grpc.WithUnaryInterceptor(api.invoke)),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := sdk.Shutdown(context.Background()); err != nil {
			t.Error(err)
		}
	})
	return map[string]func() (tfprotov6.ProviderServer, error){
		"yandex": providerserver.NewProtocol6WithError(&routingInstanceTestProvider{
			Provider: yandexframework.NewFrameworkProvider(),
			config: &providerconfig.Config{
				SDKv2:         sdk,
				ProviderState: providerconfig.State{FolderID: types.StringValue("folder-id")},
			},
		}),
	}
}

type routingInstanceAPIStub struct {
	mu                   sync.Mutex
	instance             *cloudrouter.RoutingInstance
	pending              *operation.Operation
	lastUpdate           *cloudrouter.UpdateRoutingInstanceRequest
	createError          error
	readError            error
	moveError            error
	moveOperationError   error
	updateOperationError error
	reverseRead          bool
	creates              int
	updates              int
	polls                int
}

func (s *routingInstanceAPIStub) invoke(_ context.Context, method string, request, reply interface{}, _ *grpc.ClientConn, _ grpc.UnaryInvoker, _ ...grpc.CallOption) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var response proto.Message
	switch method {
	case cloudrouter.RoutingInstanceService_Get_FullMethodName:
		if s.readError != nil {
			return s.readError
		}
		if s.instance == nil || request.(*cloudrouter.GetRoutingInstanceRequest).RoutingInstanceId != s.instance.Id {
			return status.Error(codes.NotFound, "routing instance not found")
		}
		found := proto.Clone(s.instance).(*cloudrouter.RoutingInstance)
		if s.reverseRead {
			slices.Reverse(found.VpcInfo)
			slices.Reverse(found.CicPrivateConnectionInfo)
			for _, network := range found.VpcInfo {
				slices.Reverse(network.AzInfos)
				for _, zone := range network.AzInfos {
					slices.Reverse(zone.ManualInfo.Prefixes)
				}
			}
		}
		response = found
	case cloudrouter.RoutingInstanceService_Create_FullMethodName:
		if s.createError != nil {
			return s.createError
		}
		req := request.(*cloudrouter.CreateRoutingInstanceRequest)
		if req.FolderId == "" {
			return status.Error(codes.InvalidArgument, "folder_id is required")
		}
		s.creates++
		s.instance = &cloudrouter.RoutingInstance{
			Id: fmt.Sprintf("routing-instance-%d", s.creates), Name: req.Name, Description: req.Description,
			FolderId: req.FolderId, RegionId: "ru-central1", Status: cloudrouter.RoutingInstance_ACTIVE,
			CreatedAt: timestamppb.New(time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)),
			Labels:    req.Labels, DeletionProtection: req.DeletionProtection,
			VpcInfo: req.VpcInfo, CicPrivateConnectionInfo: req.CicPrivateConnectionInfo,
		}
		response = s.operation(&cloudrouter.CreateRoutingInstanceMetadata{RoutingInstanceId: s.instance.Id}, s.instance)
	case cloudrouter.RoutingInstanceService_Update_FullMethodName:
		req := request.(*cloudrouter.UpdateRoutingInstanceRequest)
		if s.instance == nil || req.RoutingInstanceId != s.instance.Id {
			return status.Error(codes.NotFound, "routing instance not found")
		}
		if len(req.GetUpdateMask().GetPaths()) == 0 {
			return status.Error(codes.InvalidArgument, "update_mask is required")
		}
		for _, path := range req.UpdateMask.Paths {
			switch path {
			case "name":
				s.instance.Name = req.Name
			case "description":
				s.instance.Description = req.Description
			case "labels":
				s.instance.Labels = req.Labels
			case "deletion_protection":
				s.instance.DeletionProtection = req.DeletionProtection
			case "vpc_info":
				s.instance.VpcInfo = req.VpcInfo
			case "cic_private_connection_info":
				s.instance.CicPrivateConnectionInfo = req.CicPrivateConnectionInfo
			default:
				return status.Errorf(codes.InvalidArgument, "unexpected update path %q", path)
			}
		}
		s.updates++
		s.lastUpdate = proto.Clone(req).(*cloudrouter.UpdateRoutingInstanceRequest)
		response = s.operation(&cloudrouter.UpdateRoutingInstanceMetadata{RoutingInstanceId: s.instance.Id}, s.instance)
		if s.updateOperationError != nil {
			s.pending.Result = &operation.Operation_Error{Error: status.Convert(s.updateOperationError).Proto()}
		}
	case cloudrouter.RoutingInstanceService_Move_FullMethodName:
		if s.moveError != nil {
			return s.moveError
		}
		req := request.(*cloudrouter.MoveRoutingInstanceRequest)
		if s.instance == nil || req.RoutingInstanceId != s.instance.Id {
			return status.Error(codes.NotFound, "routing instance not found")
		}
		if s.moveOperationError == nil {
			s.instance.FolderId = req.DestinationFolderId
		}
		response = s.operation(&cloudrouter.MoveRoutingInstanceMetadata{RoutingInstanceId: s.instance.Id}, s.instance)
		if s.moveOperationError != nil {
			s.pending.Result = &operation.Operation_Error{Error: status.Convert(s.moveOperationError).Proto()}
		}
	case cloudrouter.RoutingInstanceService_Delete_FullMethodName:
		req := request.(*cloudrouter.DeleteRoutingInstanceRequest)
		if s.instance == nil || req.RoutingInstanceId != s.instance.Id {
			return status.Error(codes.NotFound, "routing instance not found")
		}
		if s.instance.DeletionProtection {
			return status.Error(codes.FailedPrecondition, "deletion protection is enabled")
		}
		s.instance = nil
		response = s.operation(&cloudrouter.DeleteRoutingInstanceMetadata{RoutingInstanceId: req.RoutingInstanceId}, &emptypb.Empty{})
	case "/yandex.cloud.operation.OperationService/Get":
		if s.pending == nil || request.(*operation.GetOperationRequest).OperationId != s.pending.Id {
			return status.Error(codes.NotFound, "operation not found")
		}
		s.polls++
		response = s.pending
	default:
		return status.Errorf(codes.Unimplemented, "unexpected RPC %s", method)
	}
	proto.Merge(reply.(proto.Message), response)
	return nil
}

func (s *routingInstanceAPIStub) operation(metadata, response proto.Message) *operation.Operation {
	metadataAny, _ := anypb.New(metadata)
	responseAny, _ := anypb.New(response)
	s.pending = &operation.Operation{
		Id: "operation-id", Done: true, Metadata: metadataAny,
		Result: &operation.Operation_Response{Response: responseAny},
	}
	return &operation.Operation{Id: s.pending.Id, Metadata: metadataAny}
}

func (s *routingInstanceAPIStub) checkUpdateMask(paths ...string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		s.mu.Lock()
		defer s.mu.Unlock()
		got := slices.Clone(s.lastUpdate.GetUpdateMask().GetPaths())
		slices.Sort(got)
		slices.Sort(paths)
		if !reflect.DeepEqual(got, paths) {
			return fmt.Errorf("update mask = %v, want %v", got, paths)
		}
		return nil
	}
}

func (s *routingInstanceAPIStub) checkMutations(creates, updates int) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		s.mu.Lock()
		defer s.mu.Unlock()
		if s.creates != creates || s.updates != updates {
			return fmt.Errorf("mutations = (%d creates, %d updates), want (%d, %d)", s.creates, s.updates, creates, updates)
		}
		if s.polls < creates+updates {
			return fmt.Errorf("operations were not polled: got %d polls for %d mutations", s.polls, creates+updates)
		}
		return nil
	}
}

func (s *routingInstanceAPIStub) checkDestroyed(_ *terraform.State) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.instance != nil {
		return fmt.Errorf("routing instance %s still exists", s.instance.Id)
	}
	return nil
}

func routingInstanceTestConfig(name, attributes string) string {
	return fmt.Sprintf(`
resource "yandex_cloudrouter_routing_instance" "test" {
  name = %q
  %s
}
`, name, attributes)
}

const routingInstanceTestTopology = `
  vpc_info = [
    {
      vpc_network_id = "network-a"
      az_infos = [
        { manual_info = { az_id = "ru-central1-a", prefixes = ["10.0.0.0/24", "10.0.1.0/24"] } },
        { manual_info = { az_id = "ru-central1-b", prefixes = ["10.0.2.0/24"] } }
      ]
    },
    {
      vpc_network_id = "network-b"
      az_infos = [{ manual_info = { az_id = "ru-central1-a", prefixes = ["10.1.0.0/24"] } }]
    }
  ]
  cic_private_connection_info = [
    { cic_private_connection_id = "private-connection-a" },
    { cic_private_connection_id = "private-connection-b" }
  ]
`

const routingInstanceTestReorderedTopology = `
  vpc_info = [
    {
      vpc_network_id = "network-b"
      az_infos = [{ manual_info = { az_id = "ru-central1-a", prefixes = ["10.1.0.0/24"] } }]
    },
    {
      vpc_network_id = "network-a"
      az_infos = [
        { manual_info = { az_id = "ru-central1-b", prefixes = ["10.0.2.0/24"] } },
        { manual_info = { az_id = "ru-central1-a", prefixes = ["10.0.1.0/24", "10.0.0.0/24"] } }
      ]
    }
  ]
  cic_private_connection_info = [
    { cic_private_connection_id = "private-connection-b" },
    { cic_private_connection_id = "private-connection-a" }
  ]
`

const routingInstanceTestUpdatedNetworks = `
  vpc_info = [{
    vpc_network_id = "network-a"
    az_infos = [{ manual_info = { az_id = "ru-central1-a", prefixes = ["10.0.3.0/24"] } }]
  }]
`

const routingInstanceTestUpdatedTopology = routingInstanceTestUpdatedNetworks + `
  cic_private_connection_info = [{ cic_private_connection_id = "private-connection-b" }]
`
