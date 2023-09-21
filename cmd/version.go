package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Display the Version",
		Long:  ``,
		RunE:  RunVersionCmdE,
	}
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

// RunVersionCmdE runs the command to generate the cli docs
func RunVersionCmdE(cmd *cobra.Command, args []string) (err error) {
	fmt.Println("Version:", version)
	return nil
}
