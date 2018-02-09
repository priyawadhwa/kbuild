package util

import (
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

// CreateFile creates a file at path with contents specified
func CreateFile(path string, contents []byte) error {
	// Create directory path if it doesn't exist
	baseDir := filepath.Dir(path)
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		logrus.Debugf("baseDir %s for file %s does not exist. Creating.", baseDir, path)
		if err := os.MkdirAll(baseDir, 0755); err != nil {
			return err
		}
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	_, err = f.Write(contents)
	return err
}
