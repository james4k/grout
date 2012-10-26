package grout

import (
	"io"
	"os"
	"path/filepath"
)

type Content interface {
	IsDir() bool
	FullPath() string
	Path() string

	Read() error
	Write(dir string, data M) error
}

type ContentInfo struct {
	os.FileInfo
	fullpath string
	path     string
}

func (c ContentInfo) FullPath() string {
	return c.fullpath
}

func (c ContentInfo) Path() string {
	return c.path
}

func (c *ContentInfo) SetFullPath(p string) {
	c.fullpath = p
}

func (c *ContentInfo) SetPath(p string) {
	c.path = p
}

type Dir struct {
	ContentInfo
}

func (d Dir) Read() error {
	return nil
}

func (d Dir) Write(dir string, data M) error {
	return os.MkdirAll(filepath.Join(dir, d.path), 0700)
}

type File struct {
	ContentInfo
}

func (f File) Read() error {
	return nil
}

func (f File) Write(dir string, data M) error {
	oldf, err := os.Open(f.fullpath)
	if err != nil {
		return err
	}
	defer oldf.Close()

	newf, err := os.Create(filepath.Join(dir, f.path))
	if err != nil {
		return err
	}
	defer newf.Close()

	io.Copy(newf, oldf)
	return nil
}
