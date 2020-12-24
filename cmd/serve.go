package cmd

import (
	"os"

	"github.com/open-osquery/trailsc/pkg/serve"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Serve the osquery config bundle over http",
		Run:   run,
	}

	serveRaw  bool
	addr      string
	dir       string
	cert      string
	container string
)

func init() {
	wd, _ := os.Getwd()
	serveCmd.Flags().BoolVarP(&serveRaw, "raw", "r", false, "Serve raw directory")
	serveCmd.Flags().StringVarP(&addr, "addr", "a", "localhost:9000", "IP:PORT to serve the config on")
	serveCmd.Flags().StringVarP(&dir, "dir", "d", wd, "Directory to serve from")
	serveCmd.Flags().StringVarP(&cert, "cert", "c", "cert.pem", "Config signer key and certificate")
	serveCmd.Flags().StringVarP(&container, "container", "n", "trails-config", "The config bundle container name")

	rootCmd.AddCommand(serveCmd)
}

func run(cmd *cobra.Command, args []string) {
	if verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	serve.Listen(dir, addr, cert, container, serveRaw)
}
