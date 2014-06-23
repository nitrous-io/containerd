// Exposes a directory as root filesystem
package dir

import (
	"io"
	"os"
	"path"

	"github.com/nitrous-io/containerd/containerfs"
	"github.com/nitrous-io/containerd/pkg/rootfs"
)

const (
	root = "/var/lib/docker/aufs/mnt"
)

func getRoot(id string) string {
	// for development with normal lxc container
	if os.Getenv("LXC") == "true" {
		return path.Join("/var/lib/lxc", id, "rootfs")
	}
	return path.Join(root, id)
}

func init() {
	containerfs.RegisterFS("dir", newDir)
}

func newDir(id string) (containerfs.ContainerFS, error) {
	root := getRoot(id)
	return &Dir{
		root: root,
		fs:   rootfs.NewRootfs(root),
	}, nil
}

type Dir struct {
	root string
	fs   *rootfs.Rootfs
}

func (d *Dir) Create(p string) (io.WriteCloser, error) {
	return d.fs.Create(p)
}

func (d *Dir) CreateNew(p string) (io.WriteCloser, error) {
	return d.fs.CreateNew(p)
}

func (d *Dir) Open(p string) (io.ReadCloser, error) {
	return d.fs.Open(p)
}

func (d *Dir) Mkdir(p string) error {
	return d.fs.Mkdir(p)
}

func (d *Dir) Rename(oldName, newName string) error {
	return d.fs.Rename(oldName, newName)
}

func (d *Dir) Remove(p string) error {
	return d.fs.Remove(p)
}

func (d *Dir) ReadDir(p string) ([]containerfs.FileInfo, error) {
	entries, err := d.fs.ReadDir(p)
	if err != nil {
		return nil, err
	}
	return toFileInfos(entries), nil
}

func (d *Dir) Stat(p string) (containerfs.FileInfo, error) {
	stat, err := d.fs.Stat(p)
	if err != nil {
		return containerfs.FileInfo{}, err
	}
	return toFileInfo(stat), nil
}

func toFileInfo(e os.FileInfo) containerfs.FileInfo {
	return containerfs.FileInfo{
		Name:    e.Name(),
		Size:    e.Size(),
		IsDir:   e.IsDir(),
		Mode:    e.Mode(),
		ModTime: e.ModTime(),
	}
}

func toFileInfos(entries []os.FileInfo) []containerfs.FileInfo {
	fis := []containerfs.FileInfo{}
	for _, e := range entries {
		fis = append(fis, toFileInfo(e))
	}
	return fis
}
