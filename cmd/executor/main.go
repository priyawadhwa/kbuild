package main

import (
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/docker/docker/builder/dockerfile/instructions"
	"github.com/priyawadhwa/kbuild/appender"
	"github.com/priyawadhwa/kbuild/pkg/dockerfile"
	"github.com/priyawadhwa/kbuild/pkg/env"
	"github.com/priyawadhwa/kbuild/pkg/snapshot"
	"github.com/priyawadhwa/kbuild/pkg/storage"
	"github.com/priyawadhwa/kbuild/pkg/util"
)

var dockerfilePath = flag.String("dockerfile", "/dockerfile/Dockerfile", "Path to Dockerfile.")
var source = flag.String("source", "kbuild-buckets-1518126874", "Source context location")

var dir = "/"

func main() {
	flag.Parse()
	logrus.SetLevel(logrus.DebugLevel)
	// Read and parse dockerfile
	b, err := ioutil.ReadFile(*dockerfilePath)
	if err != nil {
		panic(err)
	}

	stages, err := dockerfile.Parse(b)
	if err != nil {
		panic(err)
	}
	from := stages[0].BaseName
	// Unpack file system to root
	logrus.Info("Unpacking filesystem...", from)
	err = util.GetFileSystemFromImage(from)
	if err != nil {
		panic(err)
	}

	hasher := func(p string) string {
		h := md5.New()
		fi, err := os.Lstat(p)
		if err != nil {
			panic(err)
		}
		h.Write([]byte(fi.Mode().String()))
		h.Write([]byte(fi.ModTime().String()))

		if fi.Mode().IsRegular() {
			f, err := os.Open(p)
			if err != nil {
				panic(err)
			}
			defer f.Close()
			if _, err := io.Copy(h, f); err != nil {
				panic(err)
			}
		}

		return hex.EncodeToString(h.Sum(nil))
	}

	l := snapshot.NewLayeredMap(hasher)
	snapshotter := snapshot.NewSnapshotter(l, dir)

	// Take initial snapshot
	if err := snapshotter.Init(); err != nil {
		panic(err)
	}
	// Save environment variables
	env.SetEnvironmentVariables(from)
	logrus.Info("Environment variable is ", os.Getenv("PATH"))
	for _, s := range stages {
		for _, cmd := range s.Commands {
			switch c := cmd.(type) {
			case *instructions.RunCommand:
				newCommand := []string{}
				if c.PrependShell {
					newCommand = []string{"sh", "-c"}
					newCommand = append(newCommand, strings.Join(c.CmdLine, " "))
				} else {
					newCommand = c.CmdLine
				}
				if err := executeCommand(newCommand); err != nil {
					panic(err)
				}
			case *instructions.CopyCommand:
				path := c.SourcesAndDest[0]
				dest := c.SourcesAndDest[1]
				logrus.Infof("Getting files from %s", filepath.Clean(path))
				files, err := storage.GetFilesFromStorageBucket(*source, filepath.Clean(path))
				if err != nil {
					panic(err)
				}
				for file, contents := range files {
					logrus.Infof("Creating file %s", file)
					destPath := filepath.Join(dest, file)
					err := util.CreateFile(destPath, contents)
					if err != nil {
						panic(err)
					}
				}
			}
			if err := snapshotter.TakeSnapshot(); err != nil {
				panic(err)
			}

		}
	}

	// Append layers and push image
	// Get name of final image

	destImg := os.Getenv("KBUILD_DEST_IMAGE")
	fmt.Println("Appending image to ", destImg)
	destImg = "gcr.io/priya-wadhwa/kbuilder:finalimage"
	err = appender.AppendLayersAndPushImage(from, destImg)
	if err != nil {
		panic(err)
	}

}

func executeCommand(c []string) error {
	fmt.Println("cmd: ", c[0])
	fmt.Println("args: ", c[1:])
	cmd := exec.Command(c[0], c[1:]...)
	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	fmt.Printf("Output from %s %s\n", cmd.Path, cmd.Args)
	return nil
}
