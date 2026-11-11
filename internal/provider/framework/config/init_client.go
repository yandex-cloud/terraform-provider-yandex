package config

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

const (
	methodPath  = "/yandex.cloud.priv.devtools.ycvc.v1.VersionControlService/Init"
	serviceName = "yandex.cloud.priv.devtools.ycvc.v1.VersionControlService"
	methodName  = "Init"
	endpoint    = "yc-tools-version-control.cloud.yandex.net:443"
)

type InitRequest struct {
	ToolName string `json:"tool_name"`
	Version  string `json:"version"`
}

type InitResponse struct {
	DeprecationWarning *DeprecationWarning `json:"deprecation_warning,omitempty"`
}

type DeprecationWarning struct {
	Reason string `json:"reason"`
}

type VersionControlClient struct {
	conn *grpc.ClientConn
}

func NewVersionControlClient(iamToken string, userAgent string) (*VersionControlClient, error) {
	var dialOpts []grpc.DialOption

	dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	dialOpts = append(dialOpts, grpc.WithUnaryInterceptor(
		func(ctx context.Context, method string, req, reply interface{}, conn *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+iamToken)
			return invoker(ctx, method, req, reply, conn, opts...)
		}))
	dialOpts = append(dialOpts, grpc.WithUserAgent(userAgent))

	conn, err := grpc.NewClient(endpoint, dialOpts...)
	if err != nil {
		return nil, err
	}

	return &VersionControlClient{conn: conn}, nil
}

func (c *VersionControlClient) Init(ctx context.Context, req *InitRequest) (*InitResponse, error) {
	refClient := grpcreflect.NewClientV1Alpha(ctx, grpc_reflection_v1alpha.NewServerReflectionClient(c.conn))
	defer refClient.Reset()

	svcDesc, err := refClient.ResolveService(serviceName)
	if err != nil {
		return nil, fmt.Errorf("server does not support reflection or service not found: %w", err)
	}

	methodDesc := svcDesc.FindMethodByName(methodName)
	if methodDesc == nil {
		return nil, fmt.Errorf("method %s not found in service", methodName)
	}

	reqMsg := dynamic.NewMessage(methodDesc.GetInputType())

	reqMsg.SetFieldByName("tool_name", req.ToolName)
	reqMsg.SetFieldByName("version", req.Version)

	respMsg := dynamic.NewMessage(methodDesc.GetOutputType())

	err = c.conn.Invoke(ctx, methodPath, reqMsg, respMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to invoke method: %w", err)
	}

	resp := &InitResponse{}

	if respMsg != nil {
		if warning, err := respMsg.TryGetFieldByName("deprecation_warning"); err == nil && warning != nil {
			if warnMsg, ok := warning.(*dynamic.Message); ok && warnMsg != nil {
				if reason, err := warnMsg.TryGetFieldByName("reason"); err == nil && reason != nil {
					if reasonStr, ok := reason.(string); ok {
						resp.DeprecationWarning = &DeprecationWarning{
							Reason: reasonStr,
						}
					}
				}
			}
		}
	}

	return resp, nil
}

func (c *VersionControlClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
