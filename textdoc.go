package grout

import (
	"github.com/james4k/fmatter"
	"os"
	"path/filepath"
	"text/template"
)

type TextDocument struct {
	ContentInfo
	FrontMatter M
	Template    *template.Template
}

func (d *TextDocument) Read() error {
	var err error
	d.FrontMatter = make(M, 8)
	content, err := fmatter.ReadFile(d.FullPath(), d.FrontMatter)
	if err != nil {
		return err
	}

	d.Template, err = template.New(d.Path()).Parse(string(content))
	return err
}

func (d *TextDocument) Write(dir string, data M) error {
	newf, err := os.Create(filepath.Join(dir, d.Path()))
	if err != nil {
		return err
	}
	defer newf.Close()

	data["page"] = d.FrontMatter
	err = d.Template.Execute(newf, data)
	delete(data, "page")

	return err
}
