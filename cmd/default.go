package cmd

import (
	"github.com/open-osquery/trailsc/internal/config"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:     "trailsc",
		Short:   "Manage build configuration for open-osquery",
		Version: config.GetVersion(),
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(cobraInit)
}

func cobraInit() {
}
