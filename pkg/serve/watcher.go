package serve

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/gobwas/glob"
	log "github.com/sirupsen/logrus"
)

func createWatcher(path string, matcher glob.Glob, ev chan fsnotify.Event) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalln("Failed to establish a watcher")
	}

	defer watcher.Close()
	handleCreate(watcher, path, matcher)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				log.WithError(err).Errorln("Error reading events")
				continue
			}

			if strings.HasPrefix(event.Name, ".") ||
				filepath.Ext(event.Name) != ".conf" {
				continue
			}

			if !matcher.Match(event.Name) {
				continue
			}

			if event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Rename) == 0 {
				continue
			}

			if event.Op&(fsnotify.Create|fsnotify.Rename) > 0 {
				log.WithFields(log.Fields{
					"path": event.Name,
					"op":   event.Op.String(),
				}).Infoln("CREATE/RENAME on a filepath")
				handleCreate(watcher, event.Name, matcher)
			}

			ev <- event
		case err, ok := <-watcher.Errors:
			if !ok {
				log.WithError(err).Errorln("Error reading events")
				continue
			}
			log.WithError(err).Error("Error while reading events")
		}
	}
}

func handleCreate(watcher *fsnotify.Watcher, path string, matcher glob.Glob) {
	st, err := os.Stat(path)
	if err != nil || !st.IsDir() {
		return
	}

	filepath.Walk(path, func(fp string, fd os.FileInfo, err error) error {
		if !fd.IsDir() || !matcher.Match(fp) {
			return err
		}

		return watcher.Add(fp)
	})
}
