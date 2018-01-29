package main

import (
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/docker/docker/builder/dockerfile/instructions"
	"github.com/priyawadhwa/kbuild/appender"
	"github.com/priyawadhwa/kbuild/pkg/dockerfile"

	"github.com/priyawadhwa/kbuild/pkg/snapshot"
	"github.com/priyawadhwa/kbuild/pkg/util"
)

var dockerfilePath = flag.String("dockerfile", "/dockerfile/Dockerfile", "Path to Dockerfile.")

var dir = "/"

func main() {
	flag.Parse()

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

	// Unpack file system at /work-dir/img
	fmt.Println("Unpacking filesystem...")
	util.GetFileSystemFromImage(from)
	util.CopyTarsToFileSystem(from)
	if err != nil {
		panic(err)
	}

	commandsToRun := [][]string{}
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
				commandsToRun = append(commandsToRun, newCommand)
			}
		}
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

	for _, c := range commandsToRun {
		fmt.Println("cmd: ", c[0])
		fmt.Println("args: ", c[1:])
		if err != nil {
			panic(err)
		}
		cmd := exec.Command(c[0], c[1:]...)
		combout, err := cmd.CombinedOutput()
		if err != nil {
			panic(err)
		}
		fmt.Printf("Output from %s %s\n", cmd.Path, cmd.Args)
		fmt.Print(string(combout))

		if err := snapshotter.TakeSnapshot(); err != nil {
			panic(err)
		}
	}

	// Append layers and push image
	fmt.Println("Appending image")
	err = appender.AppendLayersAndPushImage(from, "gcr.io/priya-wadhwa/kbuild:final")
	if err != nil {
		panic(err)
	}

}
