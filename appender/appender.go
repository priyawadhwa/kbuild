package appender

import (
	"fmt"
	"github.com/containers/image/copy"
	"github.com/containers/image/docker"
	"github.com/containers/image/signature"
	"github.com/containers/image/transports/alltransports"
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
		ms.AppendLayer(contents)
	}
	return nil

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
	srcRef, err := image.NewProxyReference(nil, ms)

	destRef, err := alltransports.ParseImageName("docker://" + destImg)
	if err != nil {
		return err
	}

	policyContext, err := getPolicyContext()
	if err != nil {
		return err
	}

	err = copy.Image(policyContext, destRef, srcRef, nil)
	return err
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
