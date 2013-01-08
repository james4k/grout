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

	Read(data M) error
	Write(dir, cachedir string, data M) error
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

type ContentSlice []Content

func (c ContentSlice) Len() int {
	return len(c)
}

func (c ContentSlice) Less(i, j int) bool {
	a, ok := c[i].(Collectable)
	if !ok {
		return c[i].Path() < c[j].Path()
	}
	b, ok := c[j].(Collectable)
	if !ok {
		return c[i].Path() < c[j].Path()
	}
	return a.Less(b)
}

func (c ContentSlice) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

type Dir struct {
	ContentInfo
}

func (d Dir) Read(data M) error {
	return nil
}

func (d Dir) Write(dir, cachedir string, data M) error {
	return os.MkdirAll(filepath.Join(dir, d.path), 0700)
}

type File struct {
	ContentInfo
}

func (f File) Read(data M) error {
	return nil
}

func (f File) Write(dir, cachedir string, data M) error {
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
