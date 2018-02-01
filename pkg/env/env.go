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
	return os.Setenv("PATH", "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin")
}
