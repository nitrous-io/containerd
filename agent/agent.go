package agent

import (
	"fmt"
	"io"
	"os"
	"time"
)

type File struct {
	Content  []byte
	FileInfo *FileInfo
}

type FileInfo struct {
	Name    string
	Size    int64
	IsDir   bool
	Mode    os.FileMode
	ModTime time.Time
}

type Shell interface {
	io.Closer
	StdinPipe() (io.WriteCloser, error)
	StdoutPipe() (io.Reader, error)
	StderrPipe() (io.Reader, error)
	AllPipe() (stdin io.WriteCloser, stdout io.Reader, stderr io.Reader, err error)
	Resize(w, h float64) error
}

type Client interface {
	io.Closer

	// Command
	Output(cmd string) ([]byte, error)
	Run(cmd string) error
	Start(cmd string) error
	Wait() error

	// Filesystem
	Create(path string) (File, error)
	CreateNew(path string) (File, error)
	Open(path string) (File, error)
	Mkdir(path string) error
	Rename(oldname, newname string) error
	Remove(path string) error
	ReadDir(p string) ([]FileInfo, error)

	// Shell
	NewShell(term string, h, w float64) (Shell, error)
}

type NewClientFunc func() (Client, error)

var agents map[string]NewClientFunc

func init() {
	agents = map[string]NewClientFunc{}
}

func RegisterClient(name string, f NewClientFunc) {
	agents[name] = f
}

func GetClient(name string) (Client, error) {
	f, ok := agents[name]
	if !ok {
		return nil, fmt.Errorf("Invalid name %s", name)
	}
	return f()
}
