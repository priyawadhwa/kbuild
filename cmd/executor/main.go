package main

import (
	"flag"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"

	"github.com/priyawadhwa/kbuild/appender"
	"github.com/priyawadhwa/kbuild/commands"
	"github.com/priyawadhwa/kbuild/pkg/dockerfile"
	"github.com/priyawadhwa/kbuild/pkg/env"
	"github.com/priyawadhwa/kbuild/pkg/snapshot"
	"github.com/priyawadhwa/kbuild/pkg/storage"
	"github.com/priyawadhwa/kbuild/pkg/util"
)

var dockerfilePath = flag.String("dockerfile", "/dockerfile/Dockerfile", "Path to Dockerfile.")
var source = flag.String("source", "kbuild-buckets-1518126874", "Source context location")
var destImg = flag.String("dest", "gcr.io/priya-wadhwa/kbuilder:finalimage", "Destination of final image")

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
	logrus.Infof("Unpacking filesystem of %s...", from)
	err = util.GetFileSystemFromImage(from)
	if err != nil {
		panic(err)
	}

	l := snapshot.NewLayeredMap(util.Hasher())
	snapshotter := snapshot.NewSnapshotter(l, dir)

	// Take initial snapshot
	if err := snapshotter.Init(); err != nil {
		panic(err)
	}
	// Save environment variables
	env.SetEnvironmentVariables(from)

	// Set source information
	storage.SetBucketname(*source)

	for _, s := range stages {
		for _, cmd := range s.Commands {
			dockerCommand := commands.GetCommand(cmd)
			if err := dockerCommand.ExecuteCommand(); err != nil {
				panic(err)
			}
			if err := snapshotter.TakeSnapshot(); err != nil {
				panic(err)
			}
		}
	}

	// Append layers and push image
	if err := appender.AppendLayersAndPushImage(from, *destImg); err != nil {
		panic(err)
	}
}
