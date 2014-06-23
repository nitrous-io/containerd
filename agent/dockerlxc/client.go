package dockerlxc

import (
	"os/exec"

	"github.com/nitrous-io/containerd/agent"
	"github.com/nitrous-io/containerd/containerfs"
	_ "github.com/nitrous-io/containerd/containerfs/dir"
)

func init() {
	agent.RegisterClient("dockerlxc", newClient)
}

type Client struct {
	containerfs.ContainerFS
	username string
	id       string
}

func (c *Client) NewCommand(cmd string) agent.Command {
	return newCommand(c.id, c.username, cmd)
}

func (c *Client) NewShell() (agent.Shell, error) {
	return newShell(c.id, c.username)
}

func (c *Client) Wait() error {
	cmd := exec.Command("docker", "wait", c.id)
	return cmd.Run()
}

func newClient(id, username string) (agent.Client, error) {
	fs, err := containerfs.GetContainerFS("dir", id)
	if err != nil {
		return nil, err
	}
	return &Client{
		id:          id,
		username:    username,
		ContainerFS: fs,
	}, nil
}
