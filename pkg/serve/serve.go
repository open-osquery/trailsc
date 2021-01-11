package serve

import (
	"crypto"
	"path/filepath"
	"time"

	"net/http"
	"os"

	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
	"github.com/gobwas/glob"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/open-osquery/trailsc/internal/signer"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

var (
	green = color.New(color.FgHiGreen)
	red   = color.New(color.FgHiRed)
)

var envsWhitelist = []string{
	"prod",
	"qa",
	"dev",
}

type apiServer struct {
	fs        afero.Fs
	path      string
	addr      string
	container string

	fileHandler  http.Handler
	accessLogger *logrus.Logger
	events       chan fsnotify.Event

	signer       crypto.Signer
	leaf         []byte
	intermediate [][]byte
}

// Listen starts the local development server hosting the trails config bundles.
func Listen(path, addr, cert, container string, raw bool) {
	if _, err := os.Stat(path); err != nil {
		log.Fatalf("Unable to stat directory: %s", path)
	}

	accessLogger := log.New()
	accessLogger.SetOutput(os.Stdout)
	accessLogger.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339,
	})

	baseFs := afero.NewBasePathFs(afero.NewOsFs(), path)
	roFs := afero.NewReadOnlyFs(baseFs)
	cowFs := afero.NewCopyOnWriteFs(roFs, afero.NewMemMapFs())
	httpFs := afero.NewHttpFs(cowFs)

	srv := apiServer{
		path:         path,
		fs:           cowFs,
		addr:         addr,
		container:    container,
		accessLogger: accessLogger,
		fileHandler:  http.FileServer(httpFs.Dir(".")),
		events:       make(chan fsnotify.Event),
	}

	var err error
	f, err := os.OpenFile(cert, os.O_RDONLY, 0644)
	if err != nil {
		log.WithError(err).WithField("cert", cert).Fatalln("Failed to open cert file")
	}

	srv.signer, srv.leaf, srv.intermediate, err = signer.ParseCertificates(f)
	if err != nil {
		log.WithError(err).Fatalln("Failed to parse certificate")
	}

	log.Infoln("Building the config bundles")
	srv.bundle()

	if !raw {
		log.Infoln("Creating the config change listener")
		go createWatcher(
			path,
			// TODO (prateeknischal) Fix this glob to a configurable one
			glob.MustCompile(filepath.Join(path, "**")),
			srv.events,
		)

		go srv.reloadFsOnChange()
	}

	srv.listen()
}

func (srv apiServer) listen() {
	router := mux.NewRouter()
	router.HandleFunc("/favicon.ico", icoHandler)
	router.HandleFunc("/{env}/{bundle}", srv.handler).Methods(
		http.MethodGet, http.MethodHead)

	withLog := handlers.CustomLoggingHandler(
		srv.accessLogger.Writer(), router, logFormatter)

	log.Infoln("trailsc listening on", srv.addr)
	log.Fatalln(http.ListenAndServe(srv.addr, withLog))
}
