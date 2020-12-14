package cmd

import (
	"os"
	"time"

	"github.com/open-osquery/trailsc/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:     "trailsc",
		Short:   "Manage and build configuration for open-osquery",
		Version: config.GetVersion(),
	}
	verbose bool
)

// Execute starts the main cli
func Execute() {
	setupLogger()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(cobraInit)
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Increase verbosity")
}

func cobraInit() {
}

func setupLogger() {
	if verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: time.RFC3339,
		FullTimestamp:   true,
	})
}
