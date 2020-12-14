package serve

import (
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"net/http"
	"os"

	"github.com/fatih/color"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
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
	path string
	addr string

	fileHandler  http.Handler
	accessLogger *logrus.Logger
}

// Listen starts the local development server hosting the trails config bundles.
func Listen(path, addr string, raw bool) {
	if _, err := os.Stat(path); err != nil {
		log.Fatalf("Unable to stat directory: %s", path)
	}

	accessLogger := log.New()
	accessLogger.SetOutput(os.Stdout)
	accessLogger.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339,
	})

	srv := apiServer{
		path:         path,
		addr:         addr,
		accessLogger: accessLogger,
		fileHandler:  http.FileServer(http.Dir(path)),
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

func (srv apiServer) handler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var env, bundle string
	var ok bool

	if env, ok = vars["env"]; !ok {
		env = "qa"
	}

	if bundle, ok = vars["bundle"]; !ok {
		http.NotFound(w, r)
		return
	}

	if isValidEnv(env) && fileExists(srv.path, bundle) {
		log.WithFields(log.Fields{
			"env":    env,
			"bundle": bundle,
		}).Info("Sending to fileserver")
		http.StripPrefix(fmt.Sprintf("/%s/", env), srv.fileHandler).ServeHTTP(w, r)
		return
	}

	// The API router is not implemented yet.
	w.WriteHeader(http.StatusNotImplemented)
}

func icoHandler(w http.ResponseWriter, r *http.Request) {
	b, _ := ioutil.ReadAll(getFavicon())
	w.Header().Add("Content-type", "image/x-icon")
	w.Write(b)
}

func logFormatter(w io.Writer, params handlers.LogFormatterParams) {
	var fn *color.Color = red
	if params.StatusCode >= 200 && params.StatusCode < 299 {
		fn = green
	}

	fn.Fprintf(
		w,
		"method=%s url=%s status=%d\n",
		params.Request.Method,
		params.URL.String(),
		params.StatusCode,
	)
}

func isValidEnv(env string) bool {
	env = strings.ToLower(env)
	for _, e := range envsWhitelist {
		// Searching for prefixes and not exact string matches for extra
		// flexibility. This would allow using environments as dev-foo or prod-a
		// like environments.
		if strings.HasPrefix(env, e) {
			return true
		}
	}

	return false
}

func fileExists(dir, name string) bool {
	if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
		log.WithFields(log.Fields{
			"dir":  dir,
			"name": name,
		}).Info("File not found")
		return false
	}

	return true
}
