package cmd

import (
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/spf13/cobra"
)

func NewRootCmd(opts remote.Option) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "sbomz",
		Short:        "SBOM generator for multi-platform image indexes",
		Long:         `SBOM generator for multi-platform container image indexes.`,
		SilenceUsage: true,
	}

	cmd.AddCommand(newGenerateCmd(opts))
	cmd.AddCommand(newGetCmd(opts))
	cmd.AddCommand(newVersionCmd())

	return cmd
}
