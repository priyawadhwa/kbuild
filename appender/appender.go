package appender

import (
	"github.com/containers/image/copy"
	"github.com/containers/image/docker"
	"github.com/containers/image/signature"
	"github.com/containers/image/transports/alltransports"
	"github.com/priyawadhwa/kbuild/pkg/constants"
	"github.com/priyawadhwa/kbuild/pkg/image"
	"github.com/sirupsen/logrus"

	"io/ioutil"
	"os"
	"sort"
	"strings"
)

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
	if err := ms.SaveConfig(); err != nil {
		return err
	}
	return pushImage(dstImg)
}

func appendLayers() error {
	dir, err := os.Open(constants.WorkDir)
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
		contents, err := ioutil.ReadFile(constants.WorkDir + file)
		if err != nil {
			panic(err)
		}
		logrus.Debug("Appending layer ", file)
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
	logrus.Info("Pushing image to ", destImg)
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
		logrus.Debug("Error retrieving policy: %s", err)
		return nil, err
	}

	policyContext, err := signature.NewPolicyContext(policy)
	if err != nil {
		logrus.Debug("Error retrieving policy context: %s", err)
		return nil, err
	}
	return policyContext, nil
}
