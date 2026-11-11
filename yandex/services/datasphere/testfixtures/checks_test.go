package testfixtures

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/datasphere/v2"
	"google.golang.org/grpc"
)

type projectClientStub struct {
	ids []string
	err error
}

func (c *projectClientStub) Get(_ context.Context, req *datasphere.GetProjectRequest, _ ...grpc.CallOption) (*datasphere.Project, error) {
	c.ids = append(c.ids, req.ProjectId)
	return &datasphere.Project{Id: req.ProjectId}, c.err
}

func TestProjectDestroyChecksProjectResource(t *testing.T) {
	client := &projectClientStub{}
	calls := 0
	check := AccCheckProjectDestroy(func() ProjectClient { calls++; return client })
	if calls != 0 {
		t.Fatal("client must be resolved after provider configuration")
	}
	state := &terraform.State{Modules: []*terraform.ModuleState{{Path: []string{"root"}, Resources: map[string]*terraform.ResourceState{
		"project": {Type: "yandex_datasphere_project", Primary: &terraform.InstanceState{ID: "project-id"}},
		"other":   {Type: "yandex_datasphere_community", Primary: &terraform.InstanceState{ID: "community-id"}},
	}}}}
	if err := check(state); err == nil {
		t.Fatal("existing project must fail the destroy check")
	}
	if len(client.ids) != 1 || client.ids[0] != "project-id" {
		t.Fatalf("unexpected project lookups: %v", client.ids)
	}
	client.err = errors.New("API error")
	if err := check(state); err != nil {
		t.Fatalf("legacy DataSphere error policy changed: %v", err)
	}
}
