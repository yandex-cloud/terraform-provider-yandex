package yandex

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/loadbalancer/v1"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/apploadbalancer/v1"
)

const albTGResource = "yandex_alb_target_group.test-tg"

func init() {
	resource.AddTestSweepers("yandex_alb_target_group", &resource.Sweeper{
		Name:         "yandex_alb_target_group",
		F:            testSweepALBTargetGroups,
		Dependencies: []string{},
	})
}

func testSweepALBTargetGroups(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &apploadbalancer.ListLoadBalancersRequest{FolderId: conf.FolderID}
	albIt := conf.sdk.ApplicationLoadBalancer().LoadBalancer().LoadBalancerIterator(conf.Context(), req)
	result := &multierror.Error{}
	for albIt.Next() {
		albId := albIt.Value().GetId()
		for _, l := range albIt.Value().GetListeners() {
			routerId := l.GetHttp().Handler.HttpRouterId
			routerReq := &apploadbalancer.GetHttpRouterRequest{HttpRouterId: routerId}
			router, err := conf.sdk.ApplicationLoadBalancer().HttpRouter().Get(conf.Context(), routerReq)
			if err != nil {
				result = multierror.Append(
					result, fmt.Errorf("failed to get Router %q for Application Load Balancer %q", routerId, albId),
				)
			} else {
				for _, vh := range router.GetVirtualHosts() {
					for _, route := range vh.GetRoutes() {
						bgId := route.GetHttp().GetRoute().GetBackendGroupId() // grpc?
						bgReq := &apploadbalancer.GetBackendGroupRequest{BackendGroupId: bgId}
						bg, err := conf.sdk.ApplicationLoadBalancer().BackendGroup().Get(conf.Context(), bgReq)
						if err != nil {
							result = multierror.Append(
								result, fmt.Errorf("failed to get Backend Group %q for Application Load Balancer %q", bgId, albId),
							)
						} else {
							for _, backend := range bg.GetHttp().GetBackends() {
								for _, tgId := range backend.GetTargetGroups().GetTargetGroupIds() {
									if !sweepALBNetworkLoadBalancerAttachments(conf, albId, l.GetName(), routerId, route.GetName(), vh.GetName(), bg.GetId(), backend.GetName(), tgId) {
										result = multierror.Append(
											result, fmt.Errorf("failed to sweep Attached Target Group %q for Application Load Balancer %q", tgId, albId),
										)
									}
								}
							}
						}
					}
				}
			}
		}
	}

	if err := result.ErrorOrNil(); err != nil {
		return err
	}

	reqTg := &loadbalancer.ListTargetGroupsRequest{FolderId: conf.FolderID}
	it := conf.sdk.LoadBalancer().TargetGroup().TargetGroupIterator(conf.Context(), reqTg)
	for it.Next() {
		id := it.Value().GetId()
		if !sweepLBTargetGroup(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep LB Target Group %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepALBNetworkLoadBalancerAttachments(conf *Config, albId, lName, rId, rName, vhName, bgId, bName, tgId string) bool {
	return sweepWithRetryByFunc(
		conf, fmt.Sprintf("Attached Target Group %q for Application Load Balancer %q", tgId, albId),
		func(conf *Config) error {
			return sweepALBNetworkLoadBalancerAttachmentsOnce(conf, albId, lName, rId, rName, vhName, bgId, bName, tgId)
		},
	)
}

func sweepALBNetworkLoadBalancerAttachmentsOnce(conf *Config, albId, lName, rId, rName, vhName, bgId, bName, tgId string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexLBNetworkLoadBalancerDefaultTimeout)
	defer cancel()

	op, err := conf.sdk.ApplicationLoadBalancer().BackendGroup().RemoveBackend(
		ctx,
		&apploadbalancer.RemoveBackendRequest{
			BackendGroupId: bgId,
			BackendName:    bName,
		},
	)

	err1 := handleSweepOperation(ctx, conf, op, err)
	if err1 != nil {
		return err1
	}

	op, err = conf.sdk.ApplicationLoadBalancer().VirtualHost().RemoveRoute(
		ctx,
		&apploadbalancer.RemoveRouteRequest{
			HttpRouterId:    rId,
			RouteName:       rName,
			VirtualHostName: vhName,
		},
	)

	err1 = handleSweepOperation(ctx, conf, op, err)
	if err1 != nil {
		return err1
	}

	op, err = conf.sdk.ApplicationLoadBalancer().LoadBalancer().RemoveListener(
		ctx,
		&apploadbalancer.RemoveListenerRequest{
			LoadBalancerId: albId,
			Name:           lName,
		},
	)

	return handleSweepOperation(ctx, conf, op, err)
}

func sweepALBTargetGroup(conf *Config, id string) bool {
	return sweepWithRetry(sweepLBTargetGroupOnce, conf, "ALB Target Group", id)
}

func sweepALBTargetGroupOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexIAMServiceAccountDefaultTimeout)
	defer cancel()

	op, err := conf.sdk.ApplicationLoadBalancer().TargetGroup().Delete(ctx, &apploadbalancer.DeleteTargetGroupRequest{
		TargetGroupId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

func albTargetGroupImportStep() resource.TestStep {
	return resource.TestStep{
		ResourceName:      albTGResource,
		ImportState:       true,
		ImportStateVerify: true,
	}
}

func TestAccALBTargetGroup_basic(t *testing.T) {
	t.Parallel()

	var tg apploadbalancer.TargetGroup
	tgName := acctest.RandomWithPrefix("tf-target-group")
	folderID := getExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccALBTargetGroupBasic(tgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBTargetGroupExists(albTGResource, &tg),
					resource.TestCheckResourceAttr(albTGResource, "name", tgName),
					resource.TestCheckResourceAttrSet(albTGResource, "folder_id"),
					resource.TestCheckResourceAttr(albTGResource, "folder_id", folderID),
					testAccCheckALBTargetGroupContainsLabel(&tg, "tf-label", "tf-label-value"),
					testAccCheckALBTargetGroupContainsLabel(&tg, "empty-label", ""),
					testAccCheckCreatedAtAttr(albTGResource),
					//testAccCheckALBTargetGroupValues(&tg, []string{}),
				),
			},
			albTargetGroupImportStep(),
		},
	})
}

func testAccCheckALBTargetGroupDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_alb_target_group" {
			continue
		}

		_, err := config.sdk.ApplicationLoadBalancer().TargetGroup().Get(context.Background(), &apploadbalancer.GetTargetGroupRequest{
			TargetGroupId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("TargetGroup still exists")
		}
	}

	return nil
}

func testAccCheckALBTargetGroupExists(tgName string, tg *apploadbalancer.TargetGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[tgName]
		if !ok {
			return fmt.Errorf("Not found: %s", tgName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.ApplicationLoadBalancer().TargetGroup().Get(context.Background(), &apploadbalancer.GetTargetGroupRequest{
			TargetGroupId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("TargetGroup not found")
		}

		*tg = *found

		return nil
	}
}

func testAccCheckALBTargetGroupContainsLabel(tg *apploadbalancer.TargetGroup, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v, ok := tg.Labels[key]
		if !ok {
			return fmt.Errorf("Expected label with key '%s' not found", key)
		}
		if v != value {
			return fmt.Errorf("Incorrect label value for key '%s': expected '%s' but found '%s'", key, value, v)
		}
		return nil
	}
}

func testAccALBTargetGroupBasic(name string) string {
	return fmt.Sprintf(`
resource "yandex_alb_target_group" "test-tg" {
  name		= "%s"

  labels = {
    tf-label    = "tf-label-value"
    empty-label = ""
  }
}
`, name)
}
