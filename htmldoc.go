package grout

import (
	"github.com/james4k/fmatter"
	"github.com/james4k/layouts"
	"html/template"
	"os"
	"path/filepath"
)

type HTMLDocument struct {
	ContentInfo
	FrontMatter M
	Template    *template.Template
}

func (d *HTMLDocument) Read(data M) error {
	var err error
	d.FrontMatter = make(M, 8)
	content, err := fmatter.ReadFile(d.FullPath(), d.FrontMatter)
	if err != nil {
		return err
	}

	d.Template, err = template.New(d.Path()).Parse(string(content))
	return err
}

func (d *HTMLDocument) Write(dir, cachedir string, data M) error {
	newf, err := os.Create(filepath.Join(dir, d.Path()))
	if err != nil {
		return err
	}
	defer newf.Close()

	data["page"] = d.FrontMatter
	if layout, ok := d.FrontMatter["layout"]; ok && layout != "nil" {
		err = layouts.Execute(newf, layout.(string), d.Template, data)
	} else {
		err = d.Template.Execute(newf, data)
	}
	delete(data, "page")
	return err
}
