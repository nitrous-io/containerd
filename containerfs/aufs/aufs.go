package aufs

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"syscall"

	"github.com/nitrous-io/containerd/containerfs"
)

const (
	//root = "/var/lib/docker/aufs/mnt"
	root = "/var/lib/lxc"
)

type Aufs struct {
	root string
}

// create a new file with the same owner id as the parent
func (a *Aufs) Create(name string) (io.WriteCloser, error) {
	name = a.path(name)
	uid, gid, err := a.owner(filepath.Dir(name))
	if err != nil {
		return nil, err
	}
	f, err := os.Create(name)
	if err != nil {
		return nil, err
	}
	if err := f.Chown(uid, gid); err != nil {
		return nil, err
	}
	return f, nil
}

// create a new file with the same owner id as the parent. This failed if the
// file exists
func (a *Aufs) CreateNew(name string) (io.WriteCloser, error) {
	name = a.path(name)
	uid, gid, err := a.owner(filepath.Dir(name))
	if err != nil {
		return nil, err
	}
	f, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0755)
	if err != nil {
		return nil, err
	}
	if err := f.Chown(uid, gid); err != nil {
		return nil, err
	}
	return f, nil
}

func (a *Aufs) Open(name string) (io.ReadCloser, error) {
	return os.Open(a.path(name))
}

func (a *Aufs) Mkdir(name string) error {
	name = a.path(name)
	err := os.MkdirAll(name, 0755)
	if err != nil {
		return err
	}
	uid, gid, err := a.owner(filepath.Base(name))
	if err != nil {
		return err
	}
	return os.Chown(name, uid, gid)
}

func (a *Aufs) Rename(oldName, newName string) error {
	oldName = a.path(oldName)
	newName = a.path(newName)
	return os.Rename(oldName, newName)
}

func (a *Aufs) Remove(name string) error {
	name = a.path(name)
	return os.Remove(name)
}

func (a *Aufs) ReadDir(name string) ([]containerfs.FileInfo, error) {
	name = a.path(name)
	entries, err := ioutil.ReadDir(name)
	if err != nil {
		return nil, err
	}
	return toFileInfos(entries), nil
}

func (a *Aufs) Stat(name string) (containerfs.FileInfo, error) {
	name = a.path(name)
	stat, err := os.Stat(name)
	if err != nil {
		return containerfs.FileInfo{}, err
	}
	return toFileInfo(stat), nil
}

func init() {
	containerfs.RegisterFS("aufs", newAufs)
}

func newAufs(id string) (containerfs.ContainerFS, error) {
	return &Aufs{
		root: path.Join(root, id, "rootfs"),
	}, nil
}

func (a *Aufs) path(name string) string {
	return path.Join(a.root, name)
}

func (a *Aufs) owner(name string) (int, int, error) {
	stat, err := os.Stat(name)
	if err != nil {
		return 0, 0, err
	}
	s, ok := stat.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, 0, fmt.Errorf("failed to get system file stat")
	}
	return int(s.Uid), int(s.Gid), nil
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
