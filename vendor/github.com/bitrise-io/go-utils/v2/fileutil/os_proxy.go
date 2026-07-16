package fileutil

import (
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// SysStat holds file system stat information.
type SysStat struct {
	Uid int
	Gid int
}

// OsProxy defines the subset of os package functions we want to proxy.
// Add more methods as you need them.
type OsProxy interface {
	Stat(name string) (os.FileInfo, error)
	Lstat(name string) (os.FileInfo, error)
	Readlink(name string) (string, error)
	Symlink(oldname, newname string) error
	Mkdir(name string, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
	Open(name string) (*os.File, error)
	Create(name string) (*os.File, error)
	Remove(name string) error
	RemoveAll(path string) error
	Rename(oldpath, newpath string) error
	Chmod(name string, mode os.FileMode) error
	Chown(name string, uid, gid int) error
	Chtimes(name string, atime, mtime time.Time) error
	Getwd() (string, error)
	Abs(path string) (string, error)
	DirFS(dir string) fs.FS
	OpenFile(name string, flag int, perm os.FileMode) (*os.File, error)
	Lchown(name string, uid, gid int) error
}

// RealOS is the default implementation that delegates to the real os package.
type RealOS struct{}

func (RealOS) Stat(name string) (os.FileInfo, error)        { return os.Stat(name) }                //nolint:revive
func (RealOS) Lstat(name string) (os.FileInfo, error)       { return os.Lstat(name) }               //nolint:revive
func (RealOS) Readlink(name string) (string, error)         { return os.Readlink(name) }            //nolint:revive
func (RealOS) Symlink(oldname, newname string) error        { return os.Symlink(oldname, newname) } //nolint:revive
func (RealOS) Mkdir(name string, perm os.FileMode) error    { return os.Mkdir(name, perm) }         //nolint:revive
func (RealOS) MkdirAll(path string, perm os.FileMode) error { return os.MkdirAll(path, perm) }      //nolint:revive
func (RealOS) Open(name string) (*os.File, error)           { return os.Open(name) }                //nolint:revive
func (RealOS) Create(name string) (*os.File, error)         { return os.Create(name) }              //nolint:revive
func (RealOS) Remove(name string) error                     { return os.Remove(name) }              //nolint:revive
func (RealOS) RemoveAll(path string) error                  { return os.RemoveAll(path) }           //nolint:revive
func (RealOS) Rename(oldpath, newpath string) error         { return os.Rename(oldpath, newpath) }  //nolint:revive
func (RealOS) Chmod(name string, mode os.FileMode) error    { return os.Chmod(name, mode) }         //nolint:revive
func (RealOS) Chown(name string, uid, gid int) error        { return os.Chown(name, uid, gid) }     //nolint:revive
func (RealOS) Getwd() (string, error)                       { return os.Getwd() }                   //nolint:revive
func (RealOS) Abs(path string) (string, error)              { return filepath.Abs(path) }           //nolint:revive
func (RealOS) DirFS(dir string) fs.FS                       { return os.DirFS(dir) }                //nolint:revive
func (RealOS) Lchown(name string, uid, gid int) error       { return os.Lchown(name, uid, gid) }    //nolint:revive

//nolint:revive
func (RealOS) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(name, flag, perm)
}

//nolint:revive
func (RealOS) Chtimes(name string, atime, mtime time.Time) error {
	return os.Chtimes(name, atime, mtime)
}
