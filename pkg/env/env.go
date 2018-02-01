package env

import (
	"fmt"
	"github.com/containers/image/docker"
)

func SetEnvironmentVariables(image string) error {
	ref, err := docker.ParseReference("//" + image)
	if err != nil {
		return err
	}
	img, err := ref.NewImage(nil)
	if err != nil {
		return err
	}
	config := img.ConfigInfo()
	fmt.Println("CONFIG STUFF")
	fmt.Println(config)
	return nil
}
