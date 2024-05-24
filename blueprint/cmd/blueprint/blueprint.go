package main

import (
	"os"

	"github.com/yandex-cloud/terraform-provider-yandex/blueprint/command"

	_ "github.com/yandex-cloud/terraform-provider-yandex/blueprint/command/generate"
	_ "github.com/yandex-cloud/terraform-provider-yandex/blueprint/command/generate/datasource"
	_ "github.com/yandex-cloud/terraform-provider-yandex/blueprint/command/generate/resource"
)

func main() {
	command.Execute(os.Stderr)
}
