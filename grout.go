package grout

import (
	"fmt"
	"github.com/james4k/layouts"
	"io/ioutil"
	"launchpad.net/goyaml"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func Build(input, output string, opt *Options) {
	if input == "" {
		input = "."
	}
	if output == "" {
		output = filepath.Join(input, "_site")
	}

	b := &builder{Options: opt}
	err := b.readConfig(input)
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	content := b.walkFiles(input)
	layouts.Clear()
	err = layouts.Glob(filepath.Join(input, "_layouts", "*"))
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	tmplData := b.makeTemplateData()
	err = b.readContent(content, tmplData)
	if err != nil {
		log.Fatalf("read error: %v\n", err)
	}

	collections := b.makeCollections()
	err = b.readCollections(input, collections, tmplData)
	if err != nil {
		log.Fatalf("read collections error: %v\n", err)
	}

	tempdir, err := ioutil.TempDir(input, "_tmpsite_")
	if err != nil {
		log.Fatalf("failed to create temp dir: %v\n", err)
	}

	err = b.writeContent(tempdir, output, content, tmplData)
	if err != nil {
		log.Fatalf("write error: %v\n", err)
	}

	err = b.writeCollections(tempdir, output, collections, tmplData)
	if err != nil {
		log.Fatalf("write collections error: %v\n", err)
	}

	os.Rename(output, tempdir+"_old")
	err = os.Rename(tempdir, output)
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	temppattern := filepath.Join(input, "_tmpsite_*")
	tempmatches, err := filepath.Glob(temppattern)
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	for _, tmp := range tempmatches {
		// Just to be safe, make sure tmp contains _tmpsite_.
		// If this ever took the wrong input.. eek.
		if !strings.Contains(tmp, "_tmpsite_") {
			panic("tried to remove unrecognized temp folder!")
		}
		os.RemoveAll(tmp)
	}

	if opt.HttpHost != "" {
		/*watcher, err := fsnotify.NewWatcher()
		if err != nil {
			halt("%v\n", err)
		}*/
		fmt.Printf("HTTP server listening on %s\n", opt.HttpHost)
		http.Handle("/", http.FileServer(http.Dir(filepath.Join(output, ""))))
		err = http.ListenAndServe(opt.HttpHost, nil)
		if err != nil {
			fmt.Println(err)
		}
	}
}

type builder struct {
	*Options
	cfg M
}

func (b *builder) readConfig(dir string) error {
	m := defaultConfig
	raw, err := ioutil.ReadFile(filepath.Join(dir, "_config.yml"))
	if err != nil {
		// if file doesn't exist, just return the default config
		b.cfg = m
		return nil
	}

	err = goyaml.Unmarshal(raw, m)
	if err != nil {
		return err
	}

	m.sanitize()
	b.cfg = m
	return nil
}

func (b *builder) makeTemplateData() M {
	m := make(M, 16)
	for k, v := range b.cfg {
		if k == "collections" {
			continue
		}
		m[k] = v
	}
	return m
}

func (b *builder) walkFiles(basepath string) []Content {
	content := make([]Content, 0, 32)
	filepath.Walk(basepath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("error walking path on \"%s\" (%v)\n", path, err)
			return nil
		}
		name := info.Name()
		if name[0] == '.' || name[0] == '_' {
			if info.IsDir() && name != "." {
				return filepath.SkipDir
			}
			return nil
		}

		relpath, err := filepath.Rel(basepath, path)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		if relpath == "." {
			return nil
		}

		ci := ContentInfo{info, path, relpath}
		if info.IsDir() {
			content = append(content, Dir{ci})
			return nil
		}

		ext := filepath.Ext(name)
		switch ext {
		case ".go":
		case ".html", ".htm":
			content = append(content, &HTMLDocument{
				ContentInfo: ci,
			})
		case ".xml", ".css":
			content = append(content, &TextDocument{
				ContentInfo: ci,
			})
		default:
			content = append(content, File{
				ContentInfo: ci,
			})
		}
		return nil
	})
	return content
}

func (b *builder) readContent(content []Content, tmplData M) error {
	var err error
	for _, c := range content {
		err = c.Read(tmplData)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *builder) writeContent(dir, cachedir string, content []Content, data M) error {
	var err error
	for _, c := range content {
		err = c.Write(dir, cachedir, data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *builder) makeCollections() []collection {
	cfg := b.cfg.Map("collections")
	if cfg == nil {
		return []collection{}
	}

	collections := make([]collection, 0, len(cfg))
	for name, iprops := range cfg {
		props, ok := iprops.(M)
		if !ok {
			continue
		}
		c := collection{name: name, config: props}
		c.generate = generators[props.String("generator", "post")]
		if c.generate == nil {
			continue
		}
		collections = append(collections, c)
	}
	return collections
}

func (b *builder) readCollections(dir string, collections []collection, tmplData M) error {
	var err error
	for i := range collections {
		c := &collections[i]
		err = c.Read(dir, b.cfg, tmplData)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *builder) writeCollections(dir, cachedir string, collections []collection, data M) error {
	var err error
	for i := range collections {
		c := &collections[i]
		err = c.Write(dir, cachedir, data)
		if err != nil {
			return err
		}
	}
	return nil
}
