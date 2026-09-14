package ziputil

import (
	"os"
	"path/filepath"
)

// OsProxy defines the subset of OS operations used by ZipManager.
type OsProxy interface {
	Lstat(name string) (os.FileInfo, error)
	Readlink(name string) (string, error)
	Open(name string) (*os.File, error)
	Create(name string) (*os.File, error)
	OpenFile(name string, flag int, perm os.FileMode) (*os.File, error)
	Remove(name string) error
	Rename(oldpath, newpath string) error
	MkdirAll(path string, perm os.FileMode) error
	Symlink(oldname, newname string) error
	EvalSymlinks(path string) (string, error)
}

// RealOS is the default implementation that delegates to the real os package.
type RealOS struct{}

func (RealOS) Lstat(name string) (os.FileInfo, error)       { return os.Lstat(name) }               //nolint:revive
func (RealOS) Readlink(name string) (string, error)         { return os.Readlink(name) }            //nolint:revive
func (RealOS) Open(name string) (*os.File, error)           { return os.Open(name) }                //nolint:revive
func (RealOS) Create(name string) (*os.File, error)         { return os.Create(name) }              //nolint:revive
func (RealOS) MkdirAll(path string, perm os.FileMode) error { return os.MkdirAll(path, perm) }      //nolint:revive
func (RealOS) Symlink(oldname, newname string) error        { return os.Symlink(oldname, newname) } //nolint:revive
func (RealOS) EvalSymlinks(path string) (string, error)     { return filepath.EvalSymlinks(path) }  //nolint:revive
func (RealOS) Remove(name string) error                     { return os.Remove(name) }              //nolint:revive
func (RealOS) Rename(oldpath, newpath string) error         { return os.Rename(oldpath, newpath) }  //nolint:revive
func (RealOS) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) { //nolint:revive
	return os.OpenFile(name, flag, perm)
}
