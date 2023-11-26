package sidecar

import (
	"fmt"
	"os"
	"os/exec"
	"path"
)

// Starts docker compose on 'docker-compose.yml' from configuration folder
// Does not wait finish.
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

// Loads docker-compose.yml file to configuration folder with name "docker-compose.yml"
// for further using by StartServices
func LoadDockerComposeFile(url string) error {
	if len(url) == 0 {
		return fmt.Errorf("wrong url")
	}

	confFolder, err := ConfFolder()
	if err != nil {
		return err
	}

	composePath := path.Join(confFolder, "docker-compose.yml")

	if _, err = os.ReadFile(composePath); err == nil {
		os.Remove(composePath)
	}

	// curl -s https://memphisdev.github.io/memphis-docker/docker-compose-dev.yml -o docker-compose.yaml
	osCMD := exec.Command("curl")
	osCMD.Args = append(osCMD.Args, "-s")
	osCMD.Args = append(osCMD.Args, url)
	osCMD.Args = append(osCMD.Args, "-o")
	osCMD.Args = append(osCMD.Args, composePath)
	osCMD.Run()

	_, err = os.ReadFile(composePath)

	return err
}
