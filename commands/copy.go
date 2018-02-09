package commands

import (
	"github.com/docker/docker/builder/dockerfile/instructions"
	"github.com/priyawadhwa/kbuild/contexts/dest"
	"github.com/priyawadhwa/kbuild/pkg/util"
	"github.com/sirupsen/logrus"

	"path/filepath"
)

type CopyCommand struct {
	cmd     *instructions.CopyCommand
	context dest.Context
}

func (c CopyCommand) ExecuteCommand() error {
	path := c.cmd.SourcesAndDest[0]
	path = filepath.Clean(path)
	dest := c.cmd.SourcesAndDest[1]
	logrus.Debugf("Executing copy command from %s to %s", path, dest)
	wildcard := containsWildcards(path)
	searchPath := path
	if wildcard {
		searchPath = ""
	}
	files, err := c.context.GetFilesFromSource(searchPath)
	if err != nil {
		return err
	}
	for file, contents := range files {
		logrus.Infof("Have file %s", file)
		if wildcard {
			matched, err := filepath.Match(path, file)
			logrus.Debugf("Tried to match %s to %s: %s", file, path, matched)
			if err != nil {
				return err
			}
			if !matched {
				continue
			}
		}
		relPath, err := filepath.Rel(path, file)
		if err != nil {
			return err
		}
		destPath := filepath.Join(dest, relPath)
		logrus.Infof("Creating file %s", destPath)
		err = util.CreateFile(destPath, contents)
		if err != nil {
			return err
		}
	}
	return nil
}

func containsWildcards(path string) bool {
	for i := 0; i < len(path); i++ {
		ch := path[i]
		// These are the wildcards that correspond to filepath.Match
		if ch == '*' || ch == '?' || ch == '[' {
			return true
		}
	}
	return false
}
