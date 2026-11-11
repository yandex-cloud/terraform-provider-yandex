package yandex_compute_disk_placement_group_iam_binding_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type placementGroupClientStub struct {
	ids []string
	err error
}

func (c *placementGroupClientStub) Get(_ context.Context, req *compute.GetDiskPlacementGroupRequest, _ ...grpc.CallOption) (*compute.DiskPlacementGroup, error) {
	c.ids = append(c.ids, req.DiskPlacementGroupId)
	return &compute.DiskPlacementGroup{}, c.err
}
func TestDiskPlacementGroupDestroyUsesGroupClient(t *testing.T) {
	state := &terraform.State{Modules: []*terraform.ModuleState{{Path: []string{"root"}, Resources: map[string]*terraform.ResourceState{
		"group": {Type: "yandex_compute_disk_placement_group", Primary: &terraform.InstanceState{ID: "group-id"}},
		"disk":  {Type: "yandex_compute_disk", Primary: &terraform.InstanceState{ID: "disk-id"}},
	}}}}
	client := &placementGroupClientStub{}
	if err := checkDiskPlacementGroupDestroy(state, client); err == nil {
		t.Fatal("existing placement group must fail the destroy check")
	}
	if len(client.ids) != 1 || client.ids[0] != "group-id" {
		t.Fatalf("unexpected group lookups: %v", client.ids)
	}
	client.err = status.Error(codes.NotFound, "missing")
	if err := checkDiskPlacementGroupDestroy(state, client); err != nil {
		t.Fatal(err)
	}
}
