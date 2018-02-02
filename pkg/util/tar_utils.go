/*
Copyright 2017 Google, Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

var whitelist = []string{"/etc/resolv.conf", "/etc/hosts", "/sys"}
var symlinks = make(map[string]string)

func unpackTar(tr *tar.Reader, path string, symlinks map[string]string) (map[string]string, error) {
	for {
		header, err := tr.Next()
		if err == io.EOF {
			// end of tar archive
			break
		}
		if err != nil {
			logrus.Error("Error getting next tar header")
			return nil, err
		}

		if strings.Contains(header.Name, ".wh.") {
			rmPath := filepath.Join(path, header.Name)
			// Remove the .wh file if it was extracted.
			if _, err := os.Stat(rmPath); !os.IsNotExist(err) {
				if err := os.Remove(rmPath); err != nil {
					logrus.Error(err)
				}
			}

			// Remove the whited-out path.
			newName := strings.Replace(rmPath, ".wh.", "", 1)
			if err = os.RemoveAll(newName); err != nil {
				logrus.Error(err)
			}
			continue
		}

		target := filepath.Join(path, header.Name)
		if checkWhitelist(target) {
			continue
		}
		mode := header.FileInfo().Mode()
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); os.IsNotExist(err) {
				if err := os.MkdirAll(target, mode); err != nil {
					return nil, err
				}
			} else {
				if err := os.Chmod(target, mode); err != nil {
					return nil, err
				}
			}

		// if it's a file create it
		case tar.TypeReg:

			err = createFile(target)
			if err != nil {
				return nil, err
			}
			currFile, err := os.Create(target)
			if err != nil {
				fmt.Println("Error creating file %s %s", target, err)
				return nil, err
			}
			// manually set permissions on file, since the default umask (022) will interfere
			if err = os.Chmod(target, mode); err != nil {
				fmt.Println("Error updating file permissions on %s", target)
				return nil, err
			}
			_, err = io.Copy(currFile, tr)
			if err != nil {
				return nil, err
			}
			currFile.Close()
		// if it's a symlink also create it
		case tar.TypeSymlink:
			err = createFile(target)
			if err != nil {
				return nil, err
			}
			currFile, err := os.Create(target)
			if err != nil {
				fmt.Println("Error creating file %s %s", target, err)
				return nil, err
			}
			// manually set permissions on file, since the default umask (022) will interfere
			if err = os.Chmod(target, mode); err != nil {
				fmt.Println("Error updating file permissions on %s", target)
				return nil, err
			}
			_, err = io.Copy(currFile, tr)
			if err != nil {
				return nil, err
			}
			// Create symlink
			symlinks[target] = header.Linkname
			currFile.Close()
		}

	}
	return symlinks, nil
}

func checkWhitelist(target string) bool {
	for _, w := range whitelist {
		if strings.HasPrefix(target, w) {
			return true
		}
	}
	return false
}

func createFile(target string) error {
	// It's possible for a file to be included before the directory it's in is created.
	baseDir := filepath.Dir(target)
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		fmt.Println("baseDir %s for file %s does not exist. Creating.", baseDir, target)
		if err := os.MkdirAll(baseDir, 0755); err != nil {
			return err
		}
	}

	// It's possible we end up creating files that can't be overwritten based on their permissions.
	// Explicitly delete an existing file before continuing.
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		fmt.Println("Removing %s for overwrite.", target)
		if err := os.Remove(target); err != nil {
			return err
		}
	}
	return nil
}
