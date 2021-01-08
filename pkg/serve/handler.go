package serve

import (
	"crypto"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/open-osquery/trailsc/internal/bundler"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

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

	// only send supported requests to the file server. If not send to the api
	// router.
	if isValidEnv(env) && fileExists(srv.fs, bundle) {
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

func (srv apiServer) reloadFsOnChange() {
	last := time.Now()
	for ev := range srv.events {
		if last.Add(time.Second).Before(time.Now()) {
			// Don't rebuild too quickly
			last = time.Now()
			log.WithFields(log.Fields{
				"name": ev.Name,
				"type": ev.Op,
			}).Infof("Changes detected, rebuilding")
			srv.bundle()
		}
	}
}

// bundle function is supposed to look at the filesystem for the "config" folder
// that contain the osquery configuration(s) in the layout as speficied by the
// namespacing spec. It will bundle up all the files and create a config.{format}
// file which may be compressed. After that, it will sign the bundle with the
// supplied private key. After that it will create another bundle with the
// config bundle, certificates and the signature written as a binary file into
// another file of the same format as the config bundle, i.e. tar.gz or tar or
// zip. This bundler function assumes the top level container to be named as
// "trails-config.{format}" which will be served by the http file server.
func (srv apiServer) bundle() {
	// TODO (prateeknischal) generalize this for all mime types
	rd, _ := bundler.StaticBundler(srv.fs, ".", "config", bundler.Gzip)
	buf, _ := ioutil.ReadAll(rd)

	srv.fs.MkdirAll(filepath.Join(srv.container, "certs"), 0755)
	afero.WriteFile(
		srv.fs, filepath.Join(srv.container, "config.tar.gz"), buf, 0644)

	// Write certs
	afero.WriteFile(
		srv.fs,
		filepath.Join(srv.container, "certs", "cert.pem"), srv.leaf, 0644)

	cert, _ := srv.fs.OpenFile(
		filepath.Join(srv.container, "certs", "intermediate.pem"),
		os.O_CREATE|os.O_TRUNC|os.O_WRONLY,
		0644,
	)
	for _, c := range srv.intermediate {
		cert.Write(c)
	}
	cert.Close()

	digest := sha256.Sum256(buf)
	signerBytes, _ := srv.signer.Sign(rand.Reader, digest[:], crypto.SHA256)

	afero.WriteFile(
		srv.fs, filepath.Join(srv.container, "config.sign"), signerBytes, 0644)

	rd, _ = bundler.StaticBundler(
		srv.fs, srv.container, srv.container, bundler.Gzip)

	buf, _ = ioutil.ReadAll(rd)
	bundleName := fmt.Sprintf("%s.tar.gz", srv.container)
	afero.WriteFile(srv.fs, bundleName, buf, 0644)

	st, _ := srv.fs.Stat(bundleName)
	log.WithFields(log.Fields{
		"lastModified": st.ModTime(),
		"size":         st.Size(),
	}).Infoln("Bundle built")
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

func fileExists(fs afero.Fs, name string) bool {
	if _, err := fs.Stat(name); err != nil {
		log.WithFields(log.Fields{
			"name": name,
		}).Info("File not found")
		return false
	}

	return true
}
