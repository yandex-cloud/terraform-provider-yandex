package sweep

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	idp "github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1/idp"
	idpsdk "github.com/yandex-cloud/go-sdk/services/organizationmanager/v1/idp"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/internal/provider/framework/config"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

const (
	userSweepPageSize             = 1000
	userpoolSweepPageSizeIdpUser  = 1000
	userSweepDeleteTimeout        = 15 * time.Minute
	testResourceNamePrefixIdpUser = "tf-acc-test-user"
)

func registerYandexOrganizationmanagerIdpUserResource(add func(string, *resource.Sweeper)) {
	add("yandex_organizationmanager_idp_user", &resource.Sweeper{
		Name:         "yandex_organizationmanager_idp_user",
		F:            sweepAllIdpUser,
		Dependencies: []string{},
	})
}

func sweepAllIdpUser(_ string) error {
	conf, err := test.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	organizationID := test.GetExampleOrganizationID()
	if organizationID == "" {
		log.Printf("[WARN] organization ID is not set, skipping user sweep")
		return nil
	}

	// First, we need to get all userpools to iterate through their users
	userpoolClient := idpsdk.NewUserpoolClient(conf.SDKv2)
	userpoolReq := &idp.ListUserpoolsRequest{
		OrganizationId: organizationID,
		PageSize:       userpoolSweepPageSizeIdpUser,
	}

	userpoolResp, err := userpoolClient.List(context.Background(), userpoolReq)
	if err != nil {
		return fmt.Errorf("error getting list of userpools: %s", err)
	}

	result := &multierror.Error{}
	userClient := idpsdk.NewUserClient(conf.SDKv2)

	// Collect all userpools (handle pagination)
	userpools := userpoolResp.Userpools
	for userpoolResp.NextPageToken != "" {
		userpoolReq.PageToken = userpoolResp.NextPageToken
		userpoolResp, err = userpoolClient.List(context.Background(), userpoolReq)
		if err != nil {
			result = multierror.Append(result, fmt.Errorf("error getting next page of userpools: %s", err))
			break
		}
		userpools = append(userpools, userpoolResp.Userpools...)
	}

	// For each userpool, get and sweep users
	for _, pool := range userpools {
		userReq := &idp.ListUsersRequest{
			UserpoolId: pool.Id,
			PageSize:   userSweepPageSize,
		}

		userResp, err := userClient.List(context.Background(), userReq)
		if err != nil {
			result = multierror.Append(result, fmt.Errorf("error getting list of users for userpool %q: %s", pool.Id, err))
			continue
		}

		// Sweep users with test prefixes
		for _, user := range userResp.Users {
			username := user.Username
			if strings.HasPrefix(username, testResourceNamePrefixIdpUser) || strings.Contains(username, "@tf-acc-test-") {
				if !sweepIdpUser(conf, user.Id) {
					result = multierror.Append(result, fmt.Errorf("failed to sweep Idp User %q", user.Id))
				}
			}
		}

		// Handle pagination for users
		for userResp.NextPageToken != "" {
			userReq.PageToken = userResp.NextPageToken
			userResp, err = userClient.List(context.Background(), userReq)
			if err != nil {
				result = multierror.Append(result, fmt.Errorf("error getting next page of users for userpool %q: %s", pool.Id, err))
				break
			}

			for _, user := range userResp.Users {
				username := user.Username
				if strings.HasPrefix(username, testResourceNamePrefixIdpUser) || strings.Contains(username, "@tf-acc-test-") {
					if !sweepIdpUser(conf, user.Id) {
						result = multierror.Append(result, fmt.Errorf("failed to sweep Idp User %q", user.Id))
					}
				}
			}
		}
	}

	return result.ErrorOrNil()
}

func sweepIdpUser(conf *provider_config.Config, id string) bool {
	return test.SweepWithRetry(sweepIdpUserOnce, conf, "Idp User", id)
}

func sweepIdpUserOnce(conf *provider_config.Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), userSweepDeleteTimeout)
	defer cancel()

	client := idpsdk.NewUserClient(conf.SDKv2)
	op, err := client.Delete(ctx, &idp.DeleteUserRequest{
		UserId: id,
	})
	if err != nil {
		return err
	}

	_, err = op.Wait(ctx)
	return err
}
