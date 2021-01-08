package bundler

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

// BundleType specifies the type of bundling for the trails-config
type BundleType string

var (
	// Gzip content type for application/gzip
	Gzip BundleType = "gzip"

	// Tar content type for application/tar
	Tar BundleType = "tar"

	// Zip content type for application/zip
	Zip BundleType = "zip"
)

// StaticBundler creates a file bundle specified by the bundle type.
func StaticBundler(
	fs afero.Fs, root, containerName string, typ BundleType,
) (io.Reader, error) {
	switch typ {
	case Gzip:
		return gzipBundler(fs, root, containerName)
	case Tar:
		return tarBundler(fs, root, containerName)
	default:
		return nil, errors.New("not implemented")
	}
}

func tarBundler(fs afero.Fs, root, containerName string) (io.Reader, error) {
	var buf bytes.Buffer
	twr := tar.NewWriter(&buf)

	err := afero.Walk(fs, root, func(fp string, fd os.FileInfo, err error) error {
		var hdrPath string
		if hdrPath, err = filepath.Rel(root, fp); err != nil {
			return err
		}

		hdrPath = filepath.Join(containerName, hdrPath)
		hdr := tar.Header{
			Typeflag: byte(tar.TypeReg),
			Name:     hdrPath,
			Size:     fd.Size(),
			Mode:     int64(fd.Mode()) & 0o0777,
			Uid:      os.Getuid(),
			Gid:      os.Getgid(),
			ModTime:  fd.ModTime(),
		}

		if hdr.ModTime.Unix() == 0 {
			hdr.ModTime = time.Now()
		}

		if fd.IsDir() {
			hdr.Typeflag = tar.TypeDir
		}

		if b, err := afero.ReadFile(fs, fp); err == nil {
			log.WithField("path", hdr.Name).Debugln("Bundling")
			twr.WriteHeader(&hdr)
			twr.Write(b)
		}

		return err
	})

	if err != nil {
		return nil, err
	}

	twr.Flush()
	twr.Close()

	return bufio.NewReader(&buf), nil
}

func gzipBundler(fs afero.Fs, root, containerName string) (io.Reader, error) {
	var twr io.Reader
	var err error

	if twr, err = tarBundler(fs, root, containerName); err != nil {
		return nil, err
	}

	var wr bytes.Buffer
	gzwr := gzip.NewWriter(&wr)
	if _, err = io.Copy(gzwr, twr); err != nil {
		log.WithError(err).Errorln("Failed to copy")
		return nil, err
	}
	gzwr.Flush()
	gzwr.Close()

	return bufio.NewReader(&wr), nil
}
