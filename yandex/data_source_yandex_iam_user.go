package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/grpc/codes"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
)

func dataSourceYandexIAMUser() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexLoginRead,
		Schema: map[string]*schema.Schema{
			"login": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"user_id"},
			},
			"user_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"login"},
			},
			"default_email": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceYandexLoginRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()
	var user *iam.UserAccount

	if v, ok := d.GetOk("login"); ok {
		login := v.(string)
		resp, err := config.sdk.IAM().YandexPassportUserAccount().GetByLogin(ctx, &iam.GetUserAccountByLoginRequest{
			Login: login,
		})

		if err != nil {
			if isStatusWithCode(err, codes.NotFound) {
				return fmt.Errorf("login not found: %s", login)
			}
			return err
		}

		user = resp
	} else if v, ok := d.GetOk("user_id"); ok {
		userID := v.(string)

		resp, err := config.sdk.IAM().UserAccount().Get(ctx, &iam.GetUserAccountRequest{
			UserAccountId: userID,
		})

		if err != nil {
			return fmt.Errorf("failed to find user with ID \"%s\": %s", userID, err)
		}

		user = resp
	} else {
		return fmt.Errorf("one of 'login' or 'user_id' must be set")
	}

	d.Set("user_id", user.Id)

	if user.UserAccount != nil {
		if yaUser := user.GetYandexPassportUserAccount(); yaUser != nil {
			d.Set("default_email", yaUser.DefaultEmail)
			d.Set("login", yaUser.Login)
		}
	}

	d.SetId(user.Id)

	return nil
}
