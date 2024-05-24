package resource

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yandex-cloud/terraform-provider-yandex/blueprint/command/generate"
	"github.com/yandex-cloud/terraform-provider-yandex/blueprint/generator"
)

var cmd = &cobra.Command{
	Use:   "resource",
	Short: "Use for terraform resource scaffolding generation",
	RunE: func(cmd *cobra.Command, _ []string) error {
		gen := generator.New(
			generate.ServiceName,
			generate.ResourceName,
			generator.WithTemplateType("resource"),
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
	cmd.Flags().StringVar(&generate.ResourceName, "name", "", "set name for generated resource")
	_ = cmd.MarkFlagRequired("name")

	generate.AddSubCommand(cmd)
}
