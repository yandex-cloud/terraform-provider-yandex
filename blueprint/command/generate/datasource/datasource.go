package datasource

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yandex-cloud/terraform-provider-yandex/blueprint/command/generate"
	"github.com/yandex-cloud/terraform-provider-yandex/blueprint/generator"
)

var cmd = &cobra.Command{
	Use:   "datasource",
	Short: "Use for terraform datasource scaffolding generation",
	RunE: func(cmd *cobra.Command, _ []string) error {
		gen := generator.New(
			generate.ServiceName,
			generate.DatasourceName,
			generator.WithTemplateType("datasource"),
			generator.WithTemplateName(generate.Template),
			generator.WithOverrideFiles(generate.Override),
			generator.WithSkipComments(generate.SkipComments),
		)

		if err := gen.Generate(cmd.Context(), cmd.OutOrStdout()); err != nil {
			return fmt.Errorf("execute cmd: %w", err)
		}

		return nil
	},
	Args: cobra.NoArgs,
}

func init() {
	cmd.Flags().StringVar(&generate.ResourceName, "name", "", "set name for generated datasource")
	_ = cmd.MarkFlagRequired("name")

	generate.AddSubCommand(cmd)
}
