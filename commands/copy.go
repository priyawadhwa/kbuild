package commands

import (
	"github.com/docker/docker/builder/dockerfile/instructions"
	"github.com/priyawadhwa/kbuild/pkg/storage"
	"github.com/priyawadhwa/kbuild/pkg/util"
	"github.com/sirupsen/logrus"

	"path/filepath"
)

type CopyCommand struct {
	cmd *instructions.CopyCommand
}

func (c CopyCommand) ExecuteCommand() error {
	path := c.cmd.SourcesAndDest[0]
	dest := c.cmd.SourcesAndDest[1]
	logrus.Infof("Getting files from %s", filepath.Clean(path))
	files, err := storage.GetFilesFromStorageBucket(filepath.Clean(path))
	if err != nil {
		return err
	}
	for file, contents := range files {
		logrus.Infof("Creating file %s", file)
		destPath := filepath.Join(dest, file)
		err := util.CreateFile(destPath, contents)
		if err != nil {
			return err
		}
	}
	return nil
}
