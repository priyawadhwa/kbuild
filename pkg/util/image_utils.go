// Pull the image in the FROM part of the Dockerfile

package util

import (
	"archive/tar"
	"fmt"
	"os"
	"strings"

	"github.com/containers/image/docker"
	"github.com/containers/image/pkg/compression"
	"github.com/containers/image/types"
	"github.com/sirupsen/logrus"
)

var dir = "/Users/priyawadhwa/go/src/github.com/priyawadhwa/kbuild/testexec"

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
		fmt.Println("get file system")
		fmt.Println(reader, tr)
		fmt.Println(bi)
		copyBlobToDir(b, *tr)
		err = unpackTar(tr, dir)
		if err != nil {
			logrus.Errorf("Failed to untar layer with error: %s", err)
		}
	}
	return nil
}

func copyBlobToDir(b types.BlobInfo, tr tar.Reader) error {
	digest := strings.Split(b.Digest.String(), ":")[1]
	destFile, err := os.Create(dir + "/work-dir/" + digest + ".tar")
	if err != nil {
		return err
	}
	defer destFile.Close()
	// First copy tar file into dir
	// _, err = io.Copy(destFile, tr)
	return err
}

// GetFileSystemFromImage pulls an image and unpacks it to a file system at root
func GetFileSystemFromImage(img string) error {
	ref, err := docker.ParseReference("//" + img)
	if err != nil {
		return err
	}
	return getFileSystemFromReference(ref)
}
