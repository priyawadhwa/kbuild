package commands

import (
	"fmt"
	"github.com/docker/docker/builder/dockerfile/instructions"
	"os/exec"
	"strings"
)

type RunCommand struct {
	cmd *instructions.RunCommand
}

func (r RunCommand) ExecuteCommand() error {
	newCommand := []string{}
	if r.cmd.PrependShell {
		newCommand = []string{"sh", "-c"}
		newCommand = append(newCommand, strings.Join(r.cmd.CmdLine, " "))
	} else {
		newCommand = r.cmd.CmdLine
	}
	return execute(newCommand)
}

func execute(c []string) error {
	fmt.Println("cmd: ", c[0])
	fmt.Println("args: ", c[1:])
	cmd := exec.Command(c[0], c[1:]...)
	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	fmt.Printf("Output from %s %s\n", cmd.Path, cmd.Args)
	return nil
}
