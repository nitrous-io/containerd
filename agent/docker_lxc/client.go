package docker_lxc

import (
	"github.com/nitrous-io/containerd/agent"
	"github.com/nitrous-io/containerd/containerfs"
)

func init() {
	agent.RegisterClient("docker_lxc", newClient)
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

func newClient(id, username string) (agent.Client, error) {
	fs, err := containerfs.GetContainerFS("aufs", id)
	if err != nil {
		return nil, err
	}
	return &Client{
		id:          id,
		username:    username,
		ContainerFS: fs,
	}, nil
}
