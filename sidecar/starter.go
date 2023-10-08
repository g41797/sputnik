package sidecar

import (
	"fmt"
	"os"
	"os/exec"
	"path"
)

func StartServices() (stopServices func(), err error) {

	confFolder, err := ConfFolder()
	if err != nil {
		return nil, err
	}

	composePath := path.Join(confFolder, "docker-compose.yml")

	return StartServicesWithCompose(composePath)
}

func StartServicesWithCompose(composePath string) (stopServices func(), err error) {
	_, err = os.ReadFile(composePath)

	if err != nil {
		fmt.Println("docker-compose.yml does not exist. Please start required services manually")
		return func() {}, nil
	}

	go newCompositeCmd(composePath, "up").Run()

	return newCompositeCmd(composePath, "down").Run, nil
}

type compositeCmd struct {
	composePath string
	command     string
}

func newCompositeCmd(composePath, command string) *compositeCmd {
	result := compositeCmd{command: command, composePath: composePath}
	return &result
}

func (cmd *compositeCmd) Run() {
	osCMD := exec.Command("docker")
	osCMD.Args = append(osCMD.Args, "compose")
	osCMD.Args = append(osCMD.Args, "-f")
	osCMD.Args = append(osCMD.Args, cmd.composePath)
	osCMD.Args = append(osCMD.Args, cmd.command)
	osCMD.Run()
	return
}
