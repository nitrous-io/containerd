package containerfs

import (
	"fmt"
	"io"
	"os"
	"time"
)

type File struct {
	Content  []byte
	FileInfo FileInfo
}

type FileInfo struct {
	Name    string
	Size    int64
	IsDir   bool
	Mode    os.FileMode
	ModTime time.Time
}

type ContainerFS interface {
	// Filesystem
	Create(path string) (io.WriteCloser, error)
	CreateNew(path string) (io.WriteCloser, error)
	Open(path string) (File, error)
	Mkdir(path string) error
	Rename(oldname, newname string) error
	Remove(path string) error
	ReadDir(p string) ([]FileInfo, error)
}

type NewFS func(id string) (ContainerFS, error)

var fs map[string]NewFS

func init() {
	fs = map[string]NewFS{}
}

func RegisterFS(name string, f NewFS) {
	fs[name] = f
}

func GetContainerFS(name string, containerId string) (ContainerFS, error) {
	f, ok := fs[name]
	if !ok {
		return nil, fmt.Errorf("Invalid name %s", name)
	}
	return f(containerId)
}
