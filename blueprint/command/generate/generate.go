package generate

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/yandex-cloud/terraform-provider-yandex/blueprint/command"
)

var cmd = &cobra.Command{
	Use:   "generate",
	Short: "Use the sub-commands for datasource or resource generation",
}

func init() {
	dir, err := os.Getwd()
	if err != nil {
		dir = "."
	}

	command.AddSubCommand(cmd)

	cmd.PersistentFlags().StringVar(&Template, "template", "default", "use specific template for generation.")
	cmd.PersistentFlags().StringVar(&PathToRepo, "path", dir, "set project folder path.")
	cmd.PersistentFlags().StringVar(&ServiceName, "service-name", "", "set name for the service of generated resource.")
	cmd.PersistentFlags().BoolVar(&Override, "force", false, "set if you want to override existing files.")
	cmd.PersistentFlags().BoolVar(&SkipComments, "skip-comments", false, "set if you want to generate file without any tips for developers.")

	_ = cmd.MarkPersistentFlagRequired("service-name")
}

func AddSubCommand(sub *cobra.Command) {
	cmd.AddCommand(sub)
}
