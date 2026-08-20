package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	ydbsdk "github.com/yandex-cloud/go-sdk/services/ydb/v1"
)

const yandexIAMYDBDefaultTimeout = 1 * time.Minute
const yandexIAMYDBUpdateAccessBindingsBatchSize = 1000

var IamYDBDatabaseSchema = map[string]*schema.Schema{
	"database_id": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: "The [Managed Service YDB instance](https://yandex.cloud/docs/ydb/) Database ID to apply a binding to.",
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

func (u *YDBDatabaseIamUpdater) GetResourceIamPolicy(ctx context.Context) (*Policy, error) {
	bindings, err := getYDBDatabaseAccessBindings(ctx, u.Config, u.GetResourceID())
	if err != nil {
		return nil, err
	}
	return &Policy{bindings}, nil
}

func (u *YDBDatabaseIamUpdater) SetResourceIamPolicy(ctx context.Context, policy *Policy) error {
	req := &access.SetAccessBindingsRequest{
		ResourceId:     u.databaseID,
		AccessBindings: policy.Bindings,
	}

	ctx, cancel := context.WithTimeout(ctx, yandexIAMYDBDefaultTimeout)
	defer cancel()

	client := ydbsdk.NewDatabaseClient(u.Config.SDK)

	op, err := client.SetAccessBindings(ctx, req)
	if err != nil {
		return fmt.Errorf("Error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	_, err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	return nil
}

func (u *YDBDatabaseIamUpdater) UpdateResourceIamPolicy(ctx context.Context, policy *PolicyDelta) error {
	bSize := yandexIAMYDBUpdateAccessBindingsBatchSize
	deltas := policy.Deltas
	dLen := len(deltas)

	for i := 0; i < countBatches(dLen, bSize); i++ {
		req := &access.UpdateAccessBindingsRequest{
			ResourceId:          u.databaseID,
			AccessBindingDeltas: deltas[i*bSize : min((i+1)*bSize, dLen)],
		}

		client := ydbsdk.NewDatabaseClient(u.Config.SDK)

		op, err := client.UpdateAccessBindings(ctx, req)
		if err != nil {
			if reqID, ok := isRequestIDPresent(err); ok {
				log.Printf("[DEBUG] request ID is %s\n", reqID)
			}
			return fmt.Errorf("Error updating access bindings of %s: %w", u.DescribeResource(), err)
		}

		_, err = op.Wait(ctx)
		if err != nil {
			return fmt.Errorf("Error updating access bindings of %s: %w", u.DescribeResource(), err)
		}
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

func getYDBDatabaseAccessBindings(ctx context.Context, config *Config, databaseID string) ([]*access.AccessBinding, error) {
	bindings := []*access.AccessBinding{}
	pageToken := ""

	for {
		client := ydbsdk.NewDatabaseClient(config.SDK)

		resp, err := client.ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
			ResourceId: databaseID,
			PageSize:   defaultListSize,
			PageToken:  pageToken,
		})

		if err != nil {
			return nil, fmt.Errorf("Error retrieving access bindings of YDB Database %s: %w", databaseID, err)
		}

		bindings = append(bindings, resp.AccessBindings...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}
	return bindings, nil
}
