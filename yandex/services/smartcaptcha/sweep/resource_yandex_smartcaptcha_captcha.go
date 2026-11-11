package sweep

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/smartcaptcha/v1"
	smartcaptchasdk "github.com/yandex-cloud/go-sdk/services/smartcaptcha/v1"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

func registerResourceYandexSmartcaptchaCaptcha(add func(string, *resource.Sweeper)) {
	add("yandex_smartcaptcha_captcha", &resource.Sweeper{
		Name: "yandex_smartcaptcha_captcha",
		F:    sweepAllCaptcha,
	})
}

func sweepAllCaptcha(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	client := smartcaptchasdk.NewCaptchaClient(conf.SDK)
	resp, err := client.List(conf.Context(), &smartcaptcha.ListCaptchasRequest{
		FolderId: conf.FolderID,
	})
	if err != nil {
		return fmt.Errorf("error getting SmartCaptcha captchas: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Resources {
		if !sweepCaptcha(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep SmartCaptcha captcha %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepCaptcha(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepCaptchaOnce, conf, "Captcha", id)
}

func sweepCaptchaOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(time.Minute)
	defer cancel()

	op, err := smartcaptchasdk.NewCaptchaClient(conf.SDK).Delete(ctx, &smartcaptcha.DeleteCaptchaRequest{
		CaptchaId: id,
	})
	if err != nil {
		return err
	}
	_, err = op.Wait(ctx)
	return err
}
