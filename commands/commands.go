package commands

import (
	"github.com/docker/docker/builder/dockerfile/instructions"
	"github.com/priyawadhwa/kbuild/contexts/dest"
)

type DockerCommand interface {
	ExecuteCommand() error
}

func GetCommand(cmd instructions.Command, context dest.Context) DockerCommand {
	switch c := cmd.(type) {
	case *instructions.RunCommand:
		return RunCommand{cmd: c}
	case *instructions.CopyCommand:
		return CopyCommand{cmd: c, context: context}
	}
	return nil
}
