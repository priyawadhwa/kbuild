// Pull the image in the FROM part of the Dockerfile

package util

import (
	"archive/tar"
	"fmt"
	"github.com/containers/image/docker"
	"github.com/containers/image/pkg/compression"
	"github.com/containers/image/signature"
	"github.com/containers/image/types"
	"os"
)

var dir = "/"

func getFileSystemFromReference(ref types.ImageReference, imgSrc types.ImageSource, path string) error {
	img, err := ref.NewImage(nil)
	if err != nil {
		return err
	}
	defer img.Close()
	fmt.Println("layer infos: ", img.LayerInfos())
	symlinks := make(map[string]string)
	for _, b := range img.LayerInfos() {
		fmt.Println("Unpacking ", b)
		bi, _, err := imgSrc.GetBlob(b)
		if err != nil {
			return err
		}
		defer bi.Close()
		f, reader, err := compression.DetectCompression(bi)
		if err != nil {
			return err
		}
		// Decompress if necessary.
		if f != nil {
			reader, err = f(reader)
			if err != nil {
				return err
			}
		}
		tr := tar.NewReader(reader)
		symlinks, err = unpackTar(tr, path, symlinks)
		if err != nil {
			return err
		}
	}
	return createSymlinks(symlinks)
}

func createSymlinks(symlinks map[string]string) error {
	for newname, oldname := range symlinks {

		err := os.Symlink(newname, oldname)
		if err != nil {
			fmt.Println(err, newname, oldname)
		}
	}
	return nil
}

func getPolicyContext() (*signature.PolicyContext, error) {
	policy, err := signature.DefaultPolicy(nil)
	if err != nil {
		fmt.Println("Error retrieving policy")
		return nil, err
	}

	policyContext, err := signature.NewPolicyContext(policy)
	if err != nil {
		fmt.Println("Error retrieving policy context")
		return nil, err
	}
	return policyContext, nil
}

// GetFileSystemFromImage pulls an image and unpacks it to a file system at root
func GetFileSystemFromImage(img string) error {
	ref, err := docker.ParseReference("//" + img)
	if err != nil {
		return err
	}
	imgSrc, err := ref.NewImageSource(nil)
	if err != nil {
		return err
	}
	return getFileSystemFromReference(ref, imgSrc, dir)
}
