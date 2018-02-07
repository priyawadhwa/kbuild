package env

import (
	// "github.com/containers/image/docker"
	"os"
)

func SetEnvironmentVariables(image string) error {
	// ref, err := docker.ParseReference("//" + image)
	// if err != nil {
	// 	return err
	// }
	// img, err := ref.NewImage(nil)
	// if err != nil {
	// 	return err
	// }
	os.Setenv("HOME", "/root")
	return os.Setenv("PATH", "/opt/python3.5/bin:/opt/python3.6/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin")
}
