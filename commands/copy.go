package commands

import (
	"github.com/docker/docker/builder/dockerfile/instructions"
	"github.com/pkg/errors"
	"github.com/priyawadhwa/kbuild/contexts/dest"
	"github.com/priyawadhwa/kbuild/pkg/util"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"strings"
)

type CopyCommand struct {
	cmd     *instructions.CopyCommand
	context dest.Context
}

func (c CopyCommand) ExecuteCommand() error {
	srcs := c.cmd.SourcesAndDest[:len(c.cmd.SourcesAndDest)-1]
	dest := c.cmd.SourcesAndDest[len(c.cmd.SourcesAndDest)-1]

	if len(srcs) > 1 {
		if !isDir(dest) {
			return errors.Errorf("When specifying multiple sources in a COPY command, destination must be a directory and end in '/'")
		}
	}

	if containsWildcards(srcs) {
		// If COPY cmd contains wildcards, we will need to look through the entire filesystem
		// So, pull in all files from the bucket, and check each one against the source
		files, err := c.context.GetFilesFromSource("")
		if err != nil {
			return err
		}
		addFiles, err := getFiles(srcs, files)
		if err != nil {
			return err
		}
		if isDir(dest) {
			for src, srcFiles := range addFiles {
				for _, file := range srcFiles {
					relPath, err := filepath.Rel(src, file)
					if err != nil {
						return err
					}
					destPath := filepath.Join(dest, relPath)
					logrus.Infof("Creating file %s", destPath)
					err = util.CreateFile(destPath, files[file])
					if err != nil {
						return err
					}
				}
			}
		} else {
			for _, srcFiles := range addFiles {
				if len(srcFiles) > 1 {
					return errors.Errorf("When specifying multiple sources in a COPY command, destination must be a directory and end in '/'")
				}
				for _, file := range srcFiles {
					logrus.Infof("Creating file %s", dest)
					err = util.CreateFile(dest, files[file])
					if err != nil {
						return err
					}
				}
			}
		}
		return nil
	}

	for _, src := range srcs {
		src = filepath.Clean(src)
		files, err := c.context.GetFilesFromSource(src)
		if err != nil {
			return err
		}
		for file, contents := range files {
			if !isDir(dest) {
				logrus.Infof("Creating file %s", dest)
				return util.CreateFile(dest, contents)
			}
			relPath, err := filepath.Rel(src, file)
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
	}
	return nil
}

func isDir(path string) bool {
	return strings.HasSuffix(path, "/")
}

func containsWildcards(paths []string) bool {
	for _, path := range paths {
		for i := 0; i < len(path); i++ {
			ch := path[i]
			// These are the wildcards that correspond to filepath.Match
			if ch == '*' || ch == '?' || ch == '[' {
				return true
			}
		}
	}
	return false
}

func getFiles(srcs []string, files map[string][]byte) (map[string][]string, error) {
	f := make(map[string][]string)
	for _, src := range srcs {
		src = filepath.Clean(src)
		addedFiles := []string{}
		for file := range files {
			matched, err := filepath.Match(src, file)
			keep := matched || strings.HasPrefix(file, src)
			logrus.Debugf("Tried to match %s to %s: %s", file, src, keep)
			if err != nil {
				return nil, err
			}
			if !keep {
				continue
			}
			logrus.Debug("Appending")
			addedFiles = append(addedFiles, file)
		}
		logrus.Debugf("Src %s and addedfiles %s", src, addedFiles)
		f[src] = addedFiles
	}
	logrus.Debug(f)
	return f, nil
}
