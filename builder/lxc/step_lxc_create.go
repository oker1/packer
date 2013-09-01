package lxc

import (
	"github.com/mitchellh/multistep"
	"fmt"
	"github.com/mitchellh/packer/packer"
	"bytes"
	"os/exec"
	"log"
	"strings"
)

type stepLxcCreate struct{}

func (s *stepLxcCreate) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*config)
	ui := state.Get("ui").(packer.Ui)

	name := config.ContainerName

	rootfs := fmt.Sprintf("/var/lib/lxc/%s/rootfs", name)

	commands := make([][]string, 5)
	commands[0] = []string{
		"lxc-create", "-n", fmt.Sprintf("%s", name), "-t", "debian",
	}
	commands[1] = []string{
		"sed", "-i", "-e",
		fmt.Sprintf("s/\\(127.0.0.1\\s\\+localhost\\)/\\1\\n127.0.1.1\\t%s\\n/g", name),
		fmt.Sprintf("%s/etc/hosts", rootfs),
	}
	commands[2] = []string{
		"chroot", rootfs, "/usr/sbin/update-rc.d", "-f", "checkroot-bootclean.sh", "remove",
	}
	commands[3] = []string{
		"chroot", rootfs, "/usr/sbin/update-rc.d", "-f", "mountall-bootclean.sh", "remove",
	}
	commands[4] = []string{
		"chroot", rootfs, "/usr/sbin/update-rc.d", "-f", "mountnfs-bootclean.sh", "remove",
	}

	ui.Say("Creating containter...")
	for _, command := range commands {
		err := s.SudoCommand(command...)
		if err != nil {
			err := fmt.Errorf("Error creating container: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *stepLxcCreate) Cleanup(state multistep.StateBag) {
	config := state.Get("config").(*config)
	ui := state.Get("ui").(packer.Ui)

	command := []string{
		"lxc-destroy", "-f", "-n", config.ContainerName,
	}

	ui.Say("Unregistering and deleting virtual machine...")
	if err := s.SudoCommand(command...); err != nil {
		ui.Error(fmt.Sprintf("Error deleting virtual machine: %s", err))
	}
}


func (s *stepLxcCreate) SudoCommand(args ...string) error {
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