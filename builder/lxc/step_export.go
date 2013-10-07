package lxc

import (
	"github.com/mitchellh/multistep"
	"fmt"
	"github.com/mitchellh/packer/packer"
	"bytes"
	"os/exec"
	"log"
	"strings"
	"path/filepath"
)

type stepExport struct{}

func (s *stepExport) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*config)
	ui := state.Get("ui").(packer.Ui)

	name := config.ContainerName

	containerDir := fmt.Sprintf("/var/lib/lxc/%s", name)
	outputPath := filepath.Join(config.OutputDir, "rootfs.tar.gz")

	commands := make([][]string, 5)
	commands[0] = []string{
		"tar", "-C", containerDir, "--numeric-owner", "-czf", outputPath, "./rootfs",
	}
	commands[1] = []string{
		"wget", "https://raw.github.com/fgrehm/vagrant-lxc/master/boxes/common/lxc-template",
		"-O", filepath.Join(config.OutputDir, "lxc-template"),
	}
	commands[2] = []string{
		"wget", "https://raw.github.com/fgrehm/vagrant-lxc/master/boxes/common/lxc.conf",
		"-O", filepath.Join(config.OutputDir, "lxc.conf"),
	}
	commands[3] = []string{
		"chmod", "+x", filepath.Join(config.OutputDir, "lxc-template"),
	}
	commands[4] = []string{
		"sh", "-c", "chown $USER:`id -gn` " + filepath.Join(config.OutputDir, "*"),
	}

	ui.Say("Exporting containter...")
	for _, command := range commands {
		err := s.SudoCommand(command...)
		if err != nil {
			err := fmt.Errorf("Error exporting container: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *stepExport) Cleanup(state multistep.StateBag) {}


func (s *stepExport) SudoCommand(args ...string) error {
	var stdout, stderr bytes.Buffer

	log.Printf("Executing sudo command: %#v", args)
	cmd := exec.Command("sudo", args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	stdoutString := strings.TrimSpace(stdout.String())
	stderrString := strings.TrimSpace(stderr.String())

	if _, ok := err.(*exec.ExitError); ok {
		err = fmt.Errorf("Sudo command error: %s", stderrString)
	}

	log.Printf("stdout: %s", stdoutString)
	log.Printf("stderr: %s", stderrString)

	return err
}