// Copyright (c) 2017 Yandex LLC. All rights reserved.
// Author: Alexey Baranov <baranovich@yandex-team.ru>

package sdk

import (
	"context"
	"crypto/tls"
	"fmt"
	"sort"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/endpoint"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	apiendpoint "github.com/yandex-cloud/go-sdk/apiendpoint"
	"github.com/yandex-cloud/go-sdk/compute"
	"github.com/yandex-cloud/go-sdk/iam"
	sdk_operation "github.com/yandex-cloud/go-sdk/operation"
	"github.com/yandex-cloud/go-sdk/pkg/grpcclient"
	"github.com/yandex-cloud/go-sdk/pkg/singleflight"
	"github.com/yandex-cloud/go-sdk/resourcemanager"
	"github.com/yandex-cloud/go-sdk/sdkerrors"
	"github.com/yandex-cloud/go-sdk/vpc"
)

type Endpoint string

const (
	ComputeServiceID            Endpoint = "compute"
	IAMServiceID                Endpoint = "iam"
	OperationServiceID          Endpoint = "operation"
	ResourceManagementServiceID Endpoint = "resourcemanager"
	//revive:disable:var-naming
	ApiEndpointServiceID Endpoint = "endpoint"
	//revive:enable:var-naming
	VpcServiceID Endpoint = "vpc"
)

type Config struct {
	OAuthToken string

	Endpoint           string
	Plaintext          bool
	TLSConfig          *tls.Config
	DialContextTimeout time.Duration
}

type lazyConn func(ctx context.Context) (*grpc.ClientConn, error)

type requestContext struct {
	conf    Config
	getConn lazyConn
}

type SDK struct {
	conf      Config
	cc        grpcclient.ConnContext
	endpoints struct {
		initDone bool
		mu       sync.Mutex
		ep       map[Endpoint]*endpoint.ApiEndpoint
	}

	initErr  error
	initCall singleflight.Call
	muErr    sync.Mutex
}

func Build(ctx context.Context, conf Config, dialOpts ...grpc.DialOption) (*SDK, error) {
	sdk := build(conf, dialOpts...)
	return sdk, nil
}

func build(conf Config, dialOpts ...grpc.DialOption) *SDK {
	creds := creds(conf)
	dialOpts = append(dialOpts, grpc.WithPerRPCCredentials(creds))
	if conf.DialContextTimeout > 0 {
		dialOpts = append(dialOpts, grpc.WithBlock(), grpc.WithTimeout(conf.DialContextTimeout))
	}
	if conf.Plaintext {
		dialOpts = append(dialOpts, grpc.WithInsecure())
	} else {
		tlsConfig := conf.TLSConfig
		if tlsConfig == nil {
			tlsConfig = &tls.Config{}
		}
		creds := credentials.NewTLS(tlsConfig)
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(creds))
	}
	cc := grpcclient.NewLazyConnContext(grpcclient.DialOptions(dialOpts...))
	sdk := &SDK{
		cc:   cc,
		conf: conf,
	}
	creds.Init(sdk.getConn(IAMServiceID), time.Now, 30*time.Second)
	return sdk
}

// Shutdown shutdowns SDK and closes all open connections.
func (sdk *SDK) Shutdown(ctx context.Context) error {
	return sdk.cc.Shutdown(ctx)
}

func (sdk *SDK) WrapOperation(op *operation.Operation, err error) (*sdk_operation.Operation, error) {
	if err != nil {
		return nil, err
	}
	return sdk_operation.New(sdk.Operation(), op), nil
}

// IAM returns IAM object that is used to operate on Yandex Cloud Identity and Access Manager
func (sdk *SDK) IAM() *iam.IAM {
	return iam.NewIAM(sdk.requestContext(IAMServiceID).getConn)
}

// Compute returns Compute object that is used to operate on Yandex Compute Cloud
func (sdk *SDK) Compute() *compute.Compute {
	return compute.NewCompute(sdk.requestContext(ComputeServiceID).getConn)
}

// VPC returns VPC object that is used to operate on Yandex VPC Cloud
func (sdk *SDK) VPC() *vpc.VPC {
	return vpc.NewVPC(sdk.requestContext(VpcServiceID).getConn)
}

// MDB returns MDB object that is used to operate on Yandex MDB Cloud
func (sdk *SDK) MDB() *MDB {
	return &MDB{sdk: sdk}
}

// Operation gets OperationService client
func (sdk *SDK) Operation() *OperationServiceClient {
	return &OperationServiceClient{getConn: sdk.getConn(OperationServiceID)}
}

func (sdk *SDK) ResourceManager() *resourcemanager.ResourceManager {
	return resourcemanager.NewResourceManager(sdk.requestContext(ResourceManagementServiceID).getConn)
}

//revive:disable:var-naming

// ApiEndpoint gets ApiEndpointService client
func (sdk *SDK) ApiEndpoint() *apiendpoint.APIEndpoint {
	return apiendpoint.NewAPIEndpoint(sdk.requestContext(ApiEndpointServiceID).getConn)
}

//revive:enable:var-naming

func (sdk *SDK) Resolve(ctx context.Context, r ...Resolver) error {
	args := make([]func() error, len(r))
	for k, v := range r {
		resolver := v
		args[k] = func() error {
			return resolver.Run(ctx, sdk)
		}
	}
	return sdkerrors.CombineGoroutines(args...)
}

func (sdk *SDK) requestContext(serviceID Endpoint) *requestContext {
	return &requestContext{
		conf:    sdk.conf,
		getConn: sdk.getConn(serviceID),
	}
}

func (sdk *SDK) getConn(serviceID Endpoint) func(ctx context.Context) (*grpc.ClientConn, error) {
	return func(ctx context.Context) (*grpc.ClientConn, error) {
		if !sdk.initDone() {
			sdk.initCall.Do(func() interface{} {
				sdk.muErr.Lock()
				sdk.initErr = sdk.initConns(ctx)
				sdk.muErr.Unlock()
				return nil
			})
			if err := sdk.InitErr(); err != nil {
				return nil, err
			}
		}
		endpoint, endpointExist := sdk.Endpoint(serviceID)
		if !endpointExist {
			return nil, fmt.Errorf("server doesn't know service \"%v\". Known services: %v",
				serviceID,
				sdk.KnownServices())
		}
		return sdk.cc.GetConn(ctx, endpoint.Address)
	}
}

func (sdk *SDK) initDone() (b bool) {
	sdk.endpoints.mu.Lock()
	b = sdk.endpoints.initDone
	sdk.endpoints.mu.Unlock()
	return
}

func (sdk *SDK) KnownServices() []string {
	sdk.endpoints.mu.Lock()
	result := make([]string, 0, len(sdk.endpoints.ep))
	for k := range sdk.endpoints.ep {
		result = append(result, string(k))
	}
	sdk.endpoints.mu.Unlock()
	sort.Strings(result)
	return result
}

func (sdk *SDK) Endpoint(endpointName Endpoint) (ep *endpoint.ApiEndpoint, exist bool) {
	sdk.endpoints.mu.Lock()
	ep, exist = sdk.endpoints.ep[endpointName]
	sdk.endpoints.mu.Unlock()
	return
}

func (sdk *SDK) InitErr() error {
	sdk.muErr.Lock()
	defer sdk.muErr.Unlock()
	return sdk.initErr
}

func (sdk *SDK) initConns(ctx context.Context) error {
	discoveryConn, err := sdk.cc.GetConn(ctx, sdk.conf.Endpoint)
	if err != nil {
		return err
	}
	ec := endpoint.NewApiEndpointServiceClient(discoveryConn)
	const defaultEndpointPageSize = 100
	listResponse, err := ec.List(ctx, &endpoint.ListApiEndpointsRequest{
		PageSize: defaultEndpointPageSize,
	})
	if err != nil {
		return err
	}
	sdk.endpoints.mu.Lock()
	sdk.endpoints.ep = make(map[Endpoint]*endpoint.ApiEndpoint, len(listResponse.Endpoints))
	for _, e := range listResponse.Endpoints {
		sdk.endpoints.ep[Endpoint(e.Id)] = e
	}
	sdk.endpoints.initDone = true
	sdk.endpoints.mu.Unlock()
	return nil
}
