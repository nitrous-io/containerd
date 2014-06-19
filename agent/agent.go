package agent

import (
	"fmt"
	"io"
	"log"

	"github.com/nitrous-io/containerd/containerfs"
)

type Shell interface {
	Open(term string, w, h float64) error
	Attach(input io.Reader, output io.Writer) chan error
	Resize(w, h float64) error
}

type Command interface {
	AllPipe() (stdin io.WriteCloser, stdout io.Reader, stderr io.Reader, err error)
	Start() error
	Wait() error
}

type Client interface {
	// Command
	NewCommand(cmd string) Command

	// Filesystem proxy to containerfs
	Create(path string) (io.WriteCloser, error)
	CreateNew(path string) (io.WriteCloser, error)
	Open(path string) (containerfs.File, error)
	Mkdir(path string) error
	Rename(oldname, newname string) error
	Remove(path string) error
	ReadDir(p string) ([]containerfs.FileInfo, error)

	// Shell
	NewShell() (Shell, error)
}

type NewClientFunc func(containerId, username string) (Client, error)

var agents map[string]NewClientFunc

func init() {
	agents = map[string]NewClientFunc{}
}

func RegisterClient(name string, f NewClientFunc) {
	agents[name] = f
}

func GetClient(name, containerId, username string) (Client, error) {
	log.Println(agents)
	f, ok := agents[name]
	if !ok {
		return nil, fmt.Errorf("Invalid name %s", name)
	}
	return f(containerId, username)
}
