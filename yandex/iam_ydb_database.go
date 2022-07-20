package yandex

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
)

const yandexIAMYDBDefaultTimeout = 1 * time.Minute

var IamYDBDatabaseSchema = map[string]*schema.Schema{
	"database_id": {
		Type:     schema.TypeString,
		Required: true,
		ForceNew: true,
	},
}

type YDBDatabaseIamUpdater struct {
	databaseID string
	Config     *Config
}

func newYDBDatabaseIamUpdater(d *schema.ResourceData, config *Config) (ResourceIamUpdater, error) {
	return &YDBDatabaseIamUpdater{
		databaseID: d.Get("database_id").(string),
		Config:     config,
	}, nil
}

func ydbDatabaseIDParseFunc(d *schema.ResourceData, _ *Config) error {
	d.Set("database_id", d.Id())
	return nil
}

func (u *YDBDatabaseIamUpdater) GetResourceIamPolicy() (*Policy, error) {
	bindings, err := getYDBDatabaseAccessBindings(u.Config, u.GetResourceID())
	if err != nil {
		return nil, err
	}
	return &Policy{bindings}, nil
}

func (u *YDBDatabaseIamUpdater) SetResourceIamPolicy(policy *Policy) error {
	req := &access.SetAccessBindingsRequest{
		ResourceId:     u.databaseID,
		AccessBindings: policy.Bindings,
	}

	ctx, cancel := context.WithTimeout(u.Config.Context(), yandexIAMYDBDefaultTimeout)
	defer cancel()

	op, err := u.Config.sdk.WrapOperation(u.Config.sdk.YDB().Database().SetAccessBindings(ctx, req))
	if err != nil {
		return fmt.Errorf("Error setting IAM policy for %s: %s", u.DescribeResource(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error setting IAM policy for %s: %s", u.DescribeResource(), err)
	}

	return nil
}

func (u *YDBDatabaseIamUpdater) GetResourceID() string {
	return u.databaseID
}

func (u *YDBDatabaseIamUpdater) GetMutexKey() string {
	return fmt.Sprintf("iam-ydb-database-%s", u.databaseID)
}

func (u *YDBDatabaseIamUpdater) DescribeResource() string {
	return fmt.Sprintf("YDB Database '%s'", u.databaseID)
}

func getYDBDatabaseAccessBindings(config *Config, databaseID string) ([]*access.AccessBinding, error) {
	bindings := []*access.AccessBinding{}
	pageToken := ""
	ctx := config.Context()

	for {
		resp, err := config.sdk.YDB().Database().ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
			ResourceId: databaseID,
			PageSize:   defaultListSize,
			PageToken:  pageToken,
		})

		if err != nil {
			return nil, fmt.Errorf("Error retrieving IAM access bindings for YDB Database %s: %s", databaseID, err)
		}

		bindings = append(bindings, resp.AccessBindings...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}
	return bindings, nil
}
