package grout

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

type Post struct {
	*HTMLDocument
	date    string
	xmldate string
	url     string
	atomid  string
}

var postNameRE = regexp.MustCompile(`^([0-9]{4})-([0-9]{2})-([0-9]{2})-([0-9A-z\-]+)$`)

func (p *Post) Read() error {
	err := p.HTMLDocument.Read()
	if err != nil {
		return err
	}
	return nil
}

func (p *Post) Write(dir string, data M) error {
	// TODO: ideally we wouldn't Mkdir for every post...but the
	// OS/filesystem caches probably work well enough.  We'll measure
	// everything later in any case.
	newdir := filepath.Dir(filepath.Join(dir, p.Path()))
	err := os.MkdirAll(newdir, 0700)
	if err != nil {
		return err
	}

	err = p.HTMLDocument.Write(dir, data)
	if err != nil {
		return err
	}

	return err
}

func (p *Post) Metadata(data M) M {
	buf := bytes.NewBuffer(make([]byte, 0, 512))
	data["page"] = p.FrontMatter
	err := p.Template.Execute(buf, data)
	delete(data, "page")
	if err != nil {
		return M{}
	}

	return M{
		"title":   p.FrontMatter["title"],
		"date":    p.date,
		"xmldate": p.xmldate,
		"url":     p.url,
		"atomid":  p.atomid,
		"content": string(buf.Bytes()),
	}
}

func GeneratePost(sitecfg, cfg M, info ContentInfo) (Content, error) {
	path := info.Path()
	ext := filepath.Ext(path)
	switch ext {
	case ".html", ".htm":
		withoutExt := path[:len(path)-len(ext)]
		matches := postNameRE.FindStringSubmatch(withoutExt)
		if len(matches) < 5 {
			return nil, fmt.Errorf("bad post name: %s", path)
		}

		// TODO: support custom permalinks, and date format for
		// metadata
		info.SetPath(fmt.Sprintf("%s/%s/%s/%s.html",
			matches[1], matches[2], matches[3], matches[4]))

		baseurl := sitecfg.String("url", "")
		url, err := BuildURL(baseurl, info.Path())
		if err != nil {
			return nil, err
		}
		atomid, err := BuildURL(baseurl, fmt.Sprintf("%s-%s-%s-%s",
			matches[1], matches[2], matches[3], matches[4]))
		if err != nil {
			return nil, err
		}

		datetime, err := time.Parse("2006 01 02",
			fmt.Sprintf("%s %s %s", matches[1], matches[2],
				matches[3]))
		if err != nil {
			return nil, err
		}

		return &Post{
			HTMLDocument: &HTMLDocument{ContentInfo: info},
			date: fmt.Sprintf("%s-%s-%s",
				matches[1], matches[2], matches[3]),
			xmldate: XMLDate(datetime),
			url:     url,
			atomid:  atomid,
		}, nil
	default:
		return nil, ErrIgnore
	}
	panic("unreachable")
}

func init() {
	RegisterGenerator("post", GeneratePost)
}
