package sweep

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func Register(add func(string, *resource.Sweeper)) {
	registerResourceYandexStorageBucket(add)
	registerResourceYandexStorageObject(add)
}
