package logging

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-log/tflogtest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	cic "github.com/yandex-cloud/go-genproto/yandex/cloud/cic/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestMaskContext(t *testing.T) {
	for i, secret := range []string{"secret-marker", "secret-marker\"\\\n\t雪"} {
		for _, protoNames := range []bool{false, true} {
			t.Run(fmt.Sprintf("secret=%d/protoNames=%t", i, protoNames), func(t *testing.T) {
				connection := &cic.PrivateConnection{Id: "visible-id", Ipv4Peering: &cic.Peering{PeerBgpMd5Key: secret}}
				response, err := anypb.New(connection)
				if err != nil {
					t.Fatal(err)
				}
				payload, err := (protojson.MarshalOptions{UseProtoNames: protoNames}).Marshal(&operation.Operation{
					Id: "visible-operation", Result: &operation.Operation_Response{Response: response},
				})
				if err != nil {
					t.Fatal(err)
				}
				var output bytes.Buffer
				ctx := tflogtest.RootLogger(context.Background(), &output)
				tflog.Debug(MaskContext(ctx), string(payload))
				entries, err := tflogtest.MultilineJSONDecode(&output)
				if err != nil || len(entries) != 1 {
					t.Fatalf("log entries = %v, error = %v", entries, err)
				}
				message := entries[0]["@message"].(string)
				if strings.Contains(message, "secret-marker") || !strings.Contains(message, "***") || !strings.Contains(message, "visible-id") {
					t.Fatalf("unexpected masked message: %s", message)
				}
				if connection.Ipv4Peering.PeerBgpMd5Key != secret {
					t.Fatal("masking modified the API message")
				}
				tflog.Debug(ctx, string(payload))
				if !strings.Contains(output.String(), "secret-marker") {
					t.Fatal("masking modified the parent logging context")
				}
			})
		}
	}
}

func TestMaskContextOmitsFailedProtoDump(t *testing.T) {
	message := &cic.Peering{PeerBgpMd5Key: "secret-marker\xff"}
	if _, err := protojson.Marshal(message); err == nil {
		t.Fatal("expected invalid UTF-8 to fail protobuf JSON encoding")
	}
	var output bytes.Buffer
	ctx := MaskContext(tflogtest.RootLogger(context.Background(), &output))
	tflog.Debug(ctx, validate.ProtoDump(message))
	if output.Len() != 0 {
		t.Fatal("failed protobuf dump was logged")
	}
	tflog.Debug(ctx, "Create private_connection x-server-request-id: visible-id")
	if !strings.Contains(output.String(), "visible-id") {
		t.Fatal("ordinary diagnostics were suppressed")
	}
}
