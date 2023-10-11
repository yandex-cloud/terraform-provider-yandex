package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/loadbalancer/v1"
)

const tgResource = "yandex_lb_target_group.test-tg"

func init() {
	resource.AddTestSweepers("yandex_lb_target_group", &resource.Sweeper{
		Name:         "yandex_lb_target_group",
		F:            testSweepLBTargetGroups,
		Dependencies: []string{},
	})
}

func testSweepLBTargetGroups(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &loadbalancer.ListNetworkLoadBalancersRequest{FolderId: conf.FolderID}
	nlbIt := conf.sdk.LoadBalancer().NetworkLoadBalancer().NetworkLoadBalancerIterator(conf.Context(), req)
	result := &multierror.Error{}
	for nlbIt.Next() {
		nlbId := nlbIt.Value().GetId()
		for _, tg := range nlbIt.Value().GetAttachedTargetGroups() {
			tgId := tg.TargetGroupId
			if !sweepLBNetworkLoadBalancerAttachments(conf, nlbId, tgId) {
				result = multierror.Append(
					result, fmt.Errorf("failed to sweep Attached Target Group %q for Network Load Balancer %q", nlbId, tgId),
				)
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

func sweepLBNetworkLoadBalancerAttachments(conf *Config, nlbId, tgId string) bool {
	return sweepWithRetryByFunc(
		conf, fmt.Sprintf("Attached Target Group %q for Network Load Balancer %q", nlbId, tgId),
		func(conf *Config) error {
			return sweepLBNetworkLoadBalancerAttachmentsOnce(conf, nlbId, tgId)
		},
	)
}

func sweepLBNetworkLoadBalancerAttachmentsOnce(conf *Config, nlbId, tgId string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexLBNetworkLoadBalancerDefaultTimeout)
	defer cancel()

	op, err := conf.sdk.LoadBalancer().NetworkLoadBalancer().DetachTargetGroup(
		ctx,
		&loadbalancer.DetachNetworkLoadBalancerTargetGroupRequest{
			NetworkLoadBalancerId: nlbId,
			TargetGroupId:         tgId,
		},
	)
	return handleSweepOperation(ctx, conf, op, err)
}

func sweepLBTargetGroup(conf *Config, id string) bool {
	return sweepWithRetry(sweepLBTargetGroupOnce, conf, "LB Target Group", id)
}

func sweepLBTargetGroupOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexIAMServiceAccountDefaultTimeout)
	defer cancel()

	op, err := conf.sdk.LoadBalancer().TargetGroup().Delete(ctx, &loadbalancer.DeleteTargetGroupRequest{
		TargetGroupId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

func targetGroupImportStep() resource.TestStep {
	return resource.TestStep{
		ResourceName:      tgResource,
		ImportState:       true,
		ImportStateVerify: true,
	}
}

func TestAccLBTargetGroup_basic(t *testing.T) {
	t.Parallel()

	var tg loadbalancer.TargetGroup
	tgName := acctest.RandomWithPrefix("tf-target-group")
	folderID := getExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLBTargetGroupBasic(tgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBTargetGroupExists(tgResource, &tg),
					resource.TestCheckResourceAttr(tgResource, "name", tgName),
					resource.TestCheckResourceAttrSet(tgResource, "folder_id"),
					resource.TestCheckResourceAttr(tgResource, "folder_id", folderID),
					testAccCheckLBTargetGroupContainsLabel(&tg, "tf-label", "tf-label-value"),
					testAccCheckLBTargetGroupContainsLabel(&tg, "empty-label", ""),
					testAccCheckCreatedAtAttr(tgResource),
					testAccCheckLBTargetGroupValues(&tg, []string{}),
				),
			},
			targetGroupImportStep(),
		},
	})
}

func TestAccLBTargetGroup_full(t *testing.T) {
	t.Parallel()

	var tg loadbalancer.TargetGroup
	targetPath := ""
	instancePrefix := acctest.RandomWithPrefix("tf-instance")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLBGeneralTGTemplate(
					"tf-target-group", "tf-descr", testAccLBBaseTemplate(instancePrefix), 1, false,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBTargetGroupExists(tgResource, &tg),
					testAccCheckLBTargetGroupValues(&tg, []string{fmt.Sprintf("%s-1", instancePrefix)}),
					testExistsFirstElementWithAttr(
						tgResource, "target", "subnet_id", &targetPath,
					),
					testCheckResourceSubAttrFn(
						tgResource, &targetPath, "subnet_id", func(value string) error {
							subnetID := tg.GetTargets()[0].SubnetId
							if value != subnetID {
								return fmt.Errorf("TargetGroup's target's sudnet_id doesnt't match. %s != %s", value, subnetID)
							}
							return nil
						},
					),
					testCheckResourceSubAttrFn(
						tgResource, &targetPath, "address", func(value string) error {
							address := tg.GetTargets()[0].Address
							if value != address {
								return fmt.Errorf("TargetGroup's target's address doesnt't match. %s != %s", value, address)
							}
							return nil
						},
					),
				),
			},
			targetGroupImportStep(),
		},
	})
}

func TestAccLBTargetGroup_update(t *testing.T) {
	var tg loadbalancer.TargetGroup
	instancePrefix := acctest.RandomWithPrefix("tf-instance")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLBNetworkLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLBGeneralTGTemplate(
					"tf-target-group", "tf-descr", testAccLBBaseTemplate(instancePrefix), 1, false,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBTargetGroupExists(tgResource, &tg),
					testAccCheckLBTargetGroupValues(&tg, []string{fmt.Sprintf("%s-1", instancePrefix)}),
				),
			},
			{
				Config: testAccLBGeneralTGTemplate(
					"tf-target-group-updated", "tf-descr-updated", testAccLBBaseTemplate(instancePrefix), 1, false,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBTargetGroupExists(tgResource, &tg),
					testAccCheckLBTargetGroupValues(&tg, []string{fmt.Sprintf("%s-1", instancePrefix)}),
				),
			},
			{
				Config: testAccLBGeneralTGTemplate(
					"tf-target-group-updated", "tf-descr-updated", testAccLBBaseTemplate(instancePrefix), 2, false,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBTargetGroupExists(tgResource, &tg),
					testAccCheckLBTargetGroupValues(&tg, []string{
						fmt.Sprintf("%s-1", instancePrefix), fmt.Sprintf("%s-2", instancePrefix),
					}),
				),
			},
			targetGroupImportStep(),
		},
	})
}

func testAccCheckLBTargetGroupDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_lb_target_group" {
			continue
		}

		_, err := config.sdk.LoadBalancer().TargetGroup().Get(context.Background(), &loadbalancer.GetTargetGroupRequest{
			TargetGroupId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("TargetGroup still exists")
		}
	}

	return nil
}

func testAccCheckLBTargetGroupExists(tgName string, tg *loadbalancer.TargetGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[tgName]
		if !ok {
			return fmt.Errorf("Not found: %s", tgName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.LoadBalancer().TargetGroup().Get(context.Background(), &loadbalancer.GetTargetGroupRequest{
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

func testAccCheckLBTargetGroupContainsLabel(tg *loadbalancer.TargetGroup, key string, value string) resource.TestCheckFunc {
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

func testAccLBTargetGroupBasic(name string) string {
	return fmt.Sprintf(`
resource "yandex_lb_target_group" "test-tg" {
  name		= "%s"

  labels = {
    tf-label    = "tf-label-value"
    empty-label = ""
  }
}
`, name)
}
