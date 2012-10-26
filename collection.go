package grout

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Collectable interface {
	Content
	Metadata(data M) M
}

type collection struct {
	name     string
	generate Generator
	config   M
	content  []Content
}

// ErrIgnore is specially handled to allow generation to proceed
var ErrIgnore = errors.New("grout: ignore content")

func (c *collection) Dir() string {
	return c.config.String("dir",
		"_"+strings.ToLower(c.name))
}

func (c *collection) Read(dir string, sitecfg, tmplData M) error {
	matches, err := filepath.Glob(filepath.Join(dir, c.Dir(), "*"))
	if err != nil {
		return err
	}

	dir = filepath.Join(dir, c.Dir())
	content := make([]Content, 0, 8)
	for _, m := range matches {
		fileinfo, err := os.Stat(m)
		if err != nil {
			return err
		}
		if fileinfo.IsDir() {
			continue
		}

		path, err := filepath.Rel(dir, m)
		if err != nil {
			return err
		}
		fmt.Println(path)

		con, err := c.generate(sitecfg, c.config, ContentInfo{fileinfo, m, path})
		if err != nil {
			if err == ErrIgnore {
				continue
			}
			return err
		}
		err = con.Read()
		if err != nil {
			return err
		}

		content = append(content, con)
	}
	// TODO: content should be sorted
	var metadata []M
	for _, con := range content {
		col, ok := con.(Collectable)
		if !ok {
			continue
		}
		metadata = append(metadata, col.Metadata(tmplData))
	}
	if metadata != nil {
		tmplData[c.name] = metadata
	}
	c.content = content
	return nil
}

func (c *collection) Write(dir string, tmplData M) error {
	var err error
	for _, con := range c.content {
		fmt.Println(con)
		err = con.Write(dir, tmplData)
		if err != nil {
			return err
		}
	}
	return nil
}
