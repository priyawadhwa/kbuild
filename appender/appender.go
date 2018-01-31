package appender

import (
	"fmt"
	"github.com/containers/image/copy"
	"github.com/containers/image/docker"
	"github.com/containers/image/signature"
	"github.com/containers/image/transports/alltransports"
	digest "github.com/opencontainers/go-digest"
	"github.com/priyawadhwa/kbuild/pkg/image"

	"io/ioutil"
	"os"
	"sort"
	"strings"
)

var directory = "/Users/priyawadhwa/go/src/github.com/priyawadhwa/kbuild/exec/work-dir/"

var ms image.MutableSource

// AppendLayersAndPushImage appends layers taken from snapshotter
// and then pushes the image to the specified destination
func AppendLayersAndPushImage(srcImg, dstImg string) error {
	if err := initializeMutableSource(srcImg); err != nil {
		return err
	}
	if err := appendLayers(); err != nil {
		return err
	}
	return pushImage(dstImg)
}

func appendLayers() error {
	fmt.Println("Appending layers")
	dir, err := os.Open(directory)
	if err != nil {
		panic(err)
	}
	defer dir.Close()
	files, err := dir.Readdir(0)
	var tars []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".tar") && strings.HasPrefix(file.Name(), "layer") {
			tars = append(tars, file.Name())
		}
	}
	sort.Strings(tars)

	for _, file := range tars {
		contents, err := ioutil.ReadFile(directory + file)
		if err != nil {
			panic(err)
		}
		// Rename file
		d := digest.FromBytes(contents).String()
		diffID := strings.Split(d, ":")[1]
		fmt.Println("Renaming ", directory+file, " to ", directory+diffID+".tar")
		os.Rename(directory+file, directory+diffID+".tar")
		ms.AppendLayer(contents)
	}
	err = ms.WriteConfig(directory)
	if err != nil {
		return err
	}
	return ms.WriteManifest(directory + "manifest.json")

}

func initializeMutableSource(img string) error {
	ref, err := docker.ParseReference("//" + img)

	if err != nil {
		return err
	}
	m, err := image.NewMutableSource(ref)
	if err != nil {
		return err
	}
	ms = *m
	return nil
}

func pushImage(destImg string) error {
	srcRef, err := alltransports.ParseImageName("dir:" + directory)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	destRef, err := alltransports.ParseImageName("docker://" + destImg)
	if err != nil {
		fmt.Println("failed to get destRef: ", err)
		os.Exit(1)
	}

	policyContext, err := getPolicyContext()
	if err != nil {
		fmt.Println("failed to get policy context: ", err)
		os.Exit(1)
	}

	err = copy.Image(policyContext, destRef, srcRef, nil)

	if err != nil {
		fmt.Println("failed to copy image: ", err)
		os.Exit(1)
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
