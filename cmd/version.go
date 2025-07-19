package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"sigs.k8s.io/release-utils/version"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Long:  `Print the version number of this sbomz binary.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version.GetVersionInfo().GitVersion)
		},
	}
}
