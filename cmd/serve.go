package cmd

import (
	"os"

	"github.com/open-osquery/trailsc/pkg/serve"
	"github.com/spf13/cobra"
)

var (
	serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Serve the osquery config bundle over http",
		Run:   run,
	}

	serveRaw bool
	addr     string
	dir      string
)

func init() {
	wd, _ := os.Getwd()
	serveCmd.Flags().BoolVarP(&serveRaw, "raw", "r", false, "Serve raw directory")
	serveCmd.Flags().StringVarP(&addr, "addr", "a", "localhost:9000", "IP:PORT to serve the config on")
	serveCmd.Flags().StringVarP(&dir, "dir", "d", wd, "Directory to serve from")

	rootCmd.AddCommand(serveCmd)
}

func run(cmd *cobra.Command, args []string) {
	serve.Listen(dir, addr, serveRaw)
}
