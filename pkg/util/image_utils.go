// Pull the image in the FROM part of the Dockerfile

package util

import (
	"archive/tar"
	"io"
	"os"
	"strings"

	"github.com/containers/image/docker"
	"github.com/containers/image/pkg/compression"
	"github.com/containers/image/types"
	"github.com/sirupsen/logrus"
)

var dir = "/Users/priyawadhwa/go/src/github.com/priyawadhwa/kbuild/exec"

func getFileSystemFromReference(ref types.ImageReference) error {
	img, err := ref.NewImage(nil)
	if err != nil {
		logrus.Error(err)
		return err
	}
	defer img.Close()

	imgSrc, err := ref.NewImageSource(nil)
	if err != nil {
		logrus.Error(err)
		return err
	}

	for _, b := range img.LayerInfos() {
		bi, _, err := imgSrc.GetBlob(b)
		if err != nil {
			logrus.Errorf("Failed to pull image layer: %s", err)
			return err
		}
		// try and detect layer compression
		f, reader, err := compression.DetectCompression(bi)
		if err != nil {
			logrus.Errorf("Failed to detect image compression: %s", err)
			return err
		}
		if f != nil {
			// decompress if necessary
			reader, err = f(reader)
			if err != nil {
				logrus.Errorf("Failed to decompress image: %s", err)
				return err
			}
		}
		tr := tar.NewReader(reader)
		if err = copyTar(bi, b); err != nil {
			return err
		}
		err = unpackTar(tr, dir)
		if err != nil {
			logrus.Errorf("Failed to untar layer with error: %s", err)
		}
	}
	return nil
}

func copyTar(bi io.ReadCloser, b types.BlobInfo) error {
	digest := strings.Split(b.Digest.String(), ":")[1]
	tarDestPath := dir + "/work-dir/" + digest + ".tar"
	tarDest, err := os.Create(tarDestPath)
	if err != nil {
		return err
	}
	defer tarDest.Close()

	_, err = io.Copy(tarDest, bi)

	return nil
}

// GetFileSystemFromImage pulls an image and unpacks it to a file system at root
func GetFileSystemFromImage(img string) error {
	ref, err := docker.ParseReference("//" + img)
	if err != nil {
		return err
	}
	return getFileSystemFromReference(ref)
}
