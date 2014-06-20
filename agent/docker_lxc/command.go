package docker_lxc

import (
	"io"
	"os/exec"
)

type command struct {
	*exec.Cmd
}

func (c *command) AllPipe() (stdin io.WriteCloser, stdout io.Reader, stderr io.Reader, err error) {
	if stdin, err = c.StdinPipe(); err != nil {
		return
	}
	if stdout, err = c.StdoutPipe(); err != nil {
		return
	}
	if stderr, err = c.StderrPipe(); err != nil {
		return
	}
	return
}

func newCommand(id, user, cmd string) *command {
	args := []string{
		"--name", id,
		"--clear-env",
		"--",
		"sudo", "-u", user,
		"-i", "sh", "-c", cmd,
	}
	return &command{
		Cmd: exec.Command("lxc-attach", args...),
	}
}
