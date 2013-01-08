package grout

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Collectable interface {
	Content
	PostRead(data M, collection []Content, i int) error
	Metadata() M
	Less(other Collectable) bool
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
	// FIXME: Probably all be much cleaner if we could work
	// with a []Collectable instead of a []Content.
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

		con, err := c.generate(sitecfg, c.config, ContentInfo{fileinfo, m, path})
		if err != nil {
			if err == ErrIgnore {
				continue
			}
			return err
		}
		content = append(content, con)
	}

	for _, con := range content {
		err = con.Read(tmplData)
		if err != nil {
			return err
		}
	}
	sort.Sort(ContentSlice(content))
	for i, con := range content {
		col, ok := con.(Collectable)
		if !ok {
			continue
		}
		err = col.PostRead(tmplData, content, i)
		if err != nil {
			return err
		}
	}

	var metadata []M
	for _, con := range content {
		col, ok := con.(Collectable)
		if !ok {
			continue
		}
		metadata = append(metadata, col.Metadata())
	}
	if metadata != nil {
		tmplData[c.name] = metadata
	}
	c.content = content
	return nil
}

func (c *collection) Write(dir, cachedir string, tmplData M) error {
	var err error
	fmt.Printf("Writing %s...\n", c.name)
	for _, con := range c.content {
		err = con.Write(dir, cachedir, tmplData)
		if err != nil {
			return err
		}
	}
	fmt.Println("")
	return nil
}
