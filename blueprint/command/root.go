package command

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "blueprint",
	Short: "Blueprint is a command for terraform scaffolding generation for resources and data sources for ya tf provider",
}

func Execute(err io.Writer) {
	if execErr := rootCmd.Execute(); execErr != nil {
		_, _ = fmt.Fprintf(err, "error happened during command execution: %s \n", execErr.Error())
		os.Exit(1)
	}
}

func AddSubCommand(sub *cobra.Command) {
	rootCmd.AddCommand(sub)
}
