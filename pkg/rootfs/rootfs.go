package rootfs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"syscall"
)

type Rootfs struct {
	root string
}

func (r *Rootfs) Create(p string) (*os.File, error) {
	return r.openFile(p, os.O_RDWR|os.O_CREATE|os.O_TRUNC)
}

func (r *Rootfs) CreateNew(p string) (*os.File, error) {
	return r.openFile(p, os.O_RDWR|os.O_CREATE|os.O_EXCL)
}

func (r *Rootfs) Open(p string) (*os.File, error) {
	return os.Open(r.path(p))
}

func (r *Rootfs) Mkdir(p string) error {
	p = r.path(p)
	stat, err := os.Stat(filepath.Dir(p))
	if err != nil {
		return err
	}
	uid, gid, err := owner(stat)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(p, 0755); err != nil {
		return err
	}
	return os.Chown(p, uid, gid)
}

func (r *Rootfs) Rename(oldName, newName string) error {
	oldName = r.path(oldName)
	newName = r.path(newName)
	return os.Rename(oldName, newName)
}

func (r *Rootfs) Remove(name string) error {
	return os.RemoveAll(r.path(name))
}

func (r *Rootfs) ReadDir(name string) ([]os.FileInfo, error) {
	return ioutil.ReadDir(r.path(name))
}

func (r *Rootfs) Stat(name string) (os.FileInfo, error) {
	return os.Stat(r.path(name))
}

func (r *Rootfs) openFile(p string, flag int) (*os.File, error) {
	p = r.path(p)
	stat, err := os.Stat(filepath.Dir(p))
	if err != nil {
		return nil, err
	}
	uid, gid, err := owner(stat)
	if err != nil {
		return nil, err
	}
	f, err := os.OpenFile(p, flag, 0755)
	if err != nil {
		return nil, err
	}
	if err := f.Chown(uid, gid); err != nil {
		return nil, err
	}
	return f, nil
}

func (r *Rootfs) path(p string) string {
	return filepath.Clean(path.Join(r.root, p))
}

func NewRootfs(root string) *Rootfs {
	return &Rootfs{root: root}
}

func owner(stat os.FileInfo) (int, int, error) {
	s, ok := stat.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, 0, fmt.Errorf("failed to get system file stat")
	}
	return int(s.Uid), int(s.Gid), nil
}
