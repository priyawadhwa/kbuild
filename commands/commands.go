package commands

import (
	"github.com/docker/docker/builder/dockerfile/instructions"
)

type DockerCommand interface {
	ExecuteCommand() error
}

func GetCommand(cmd instructions.Command) DockerCommand {
	switch c := cmd.(type) {
	case *instructions.RunCommand:
		return RunCommand{cmd: c}
	case *instructions.CopyCommand:
		return CopyCommand{cmd: c}
	}
	return nil
}
