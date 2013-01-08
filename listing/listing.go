package listing

import (
	"bytes"
	"fmt"
	. "github.com/james4k/grout"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

type Listing struct {
	*HTMLDocument
	id          int
	url         string
	img         string
	thumb       string
	thumbWidth  int
	thumbHeight int
	content     string
	metadata    M
}

var listingNameRE = regexp.MustCompile(`^([0-9]{1,10})-([0-9A-z\-]+)$`)

func (l *Listing) Read(data M) error {
	err := l.HTMLDocument.Read(data)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(make([]byte, 0, 256))
	data["page"] = l.FrontMatter
	err = l.Template.Execute(buf, data)
	delete(data, "page")
	if err != nil {
		return err
	}
	l.content = string(buf.Bytes())
	l.metadata = make(M, 8)
	for k, v := range l.FrontMatter {
		l.metadata[k] = v
	}
	l.metadata["id"] = l.id
	l.metadata["url"] = l.url
	l.metadata["img"] = l.img
	l.metadata["thumb"] = l.thumb
	l.metadata["content"] = l.content
	return nil
}

func (l *Listing) PostRead(data M, collection []Content, i int) error {
	if i > 0 {
		l.metadata["prev"] = collection[i-1].Path()
	}
	if i+1 < len(collection) {
		l.metadata["next"] = collection[i+1].Path()
	}
	return nil
}

func (l *Listing) Write(dir, cachedir string, data M) error {
	// TODO: ideally we wouldn't Mkdir for every post...but the
	// OS/filesystem caches probably work well enough.  We'll measure
	// everything later in any case.
	newdir := filepath.Dir(filepath.Join(dir, l.Path()))
	err := os.MkdirAll(newdir, 0700)
	if err != nil {
		return err
	}

	for k, v := range l.metadata {
		l.FrontMatter[k] = v
	}

	fmt.Print(l.Path())
	err = l.HTMLDocument.Write(dir, cachedir, data)
	if err != nil {
		return err
	}

	err = l.writeImages(dir, cachedir, data)
	if err != nil {
		return err
	}

	return err
}

func (l *Listing) Metadata() M {
	return l.metadata
}

func (l *Listing) Less(other Collectable) bool {
	// reverse order
	otherListing := other.(*Listing)
	return l.id > otherListing.id
}

func (l *Listing) writeImages(dir, cachedir string, data M) error {
	path := l.FullPath()
	ext := filepath.Ext(path)
	path = path[:len(path)-len(ext)]
	file, err := os.Open(path + ".png")
	if err != nil {
		file, err = os.Open(path + ".jpg")
	}
	if err != nil {
		return err
	}
	defer file.Close()

	outpath := l.Path()
	ext = filepath.Ext(outpath)
	outpath = outpath[:len(outpath)-len(ext)]
	cachepath := filepath.Join(cachedir, outpath)
	outpath = filepath.Join(dir, outpath)
	cacheinfo, err := os.Stat(cachepath + ".jpg")
	if err == nil {
		info, err := file.Stat()
		if err != nil {
			return err
		}
		if info.ModTime().Before(cacheinfo.ModTime()) {
			err = l.copyCachedImages(outpath, cachepath)
			if err == nil {
				fmt.Println(" [images cached]")
				return nil
			}
			fmt.Println(err)
		}
	}

	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}

	opt := &jpeg.Options{Quality: 85}

	fullC := make(chan error, 1)
	go func() {
		imgfile, err := os.Create(outpath + ".jpg")
		if err != nil {
			fullC <- err
			return
		}
		defer imgfile.Close()

		// FIXME: NearestNeighbor seems to be the only filter that doens't give crazy distortion on a few specific images.
		fullimg := resize.Resize(500, 0, img, resize.NearestNeighbor)
		err = jpeg.Encode(imgfile, fullimg, opt)
		if err != nil {
			fullC <- err
			return
		}
		fullC <- nil
	}()

	thumbC := make(chan error, 1)
	go func() {
		thumbfile, err := os.Create(outpath + "_thumb.jpg")
		if err != nil {
			thumbC <- err
			return
		}
		defer thumbfile.Close()

		thumbimg := resize.Resize(uint(l.thumbWidth), uint(l.thumbHeight), img, resize.Bicubic)
		err = jpeg.Encode(thumbfile, thumbimg, opt)
		if err != nil {
			thumbC <- err
			return
		}
		thumbC <- nil
	}()

	for i := 0; i < 2; i++ {
		select {
		case err = <-fullC:
			if err != nil {
				return err
			}
			fmt.Print(" +large")
		case err = <-thumbC:
			if err != nil {
				return err
			}
			fmt.Print(" +thumb")
		}
	}
	fmt.Println("")
	return nil
}

func (l *Listing) copyCachedImages(path, cachepath string) error {
	paths := []struct {
		in  string
		out string
	}{
		{cachepath + ".jpg", path + ".jpg"},
		{cachepath + "_thumb.jpg", path + "_thumb.jpg"},
	}
	for i := 0; i < len(paths); i++ {
		in, err := os.Open(paths[i].in)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.Create(paths[i].out)
		if err != nil {
			return err
		}
		defer out.Close()
		_, err = io.Copy(out, in)
		if err != nil {
			return err
		}
	}
	return nil
}

func GenerateListing(sitecfg, cfg M, info ContentInfo) (Content, error) {
	path := info.Path()
	ext := filepath.Ext(path)
	switch ext {
	case ".html", ".htm":
		withoutExt := path[:len(path)-len(ext)]
		matches := listingNameRE.FindStringSubmatch(withoutExt)
		if len(matches) < 3 {
			return nil, fmt.Errorf("bad post name: %s", path)
		}

		id, err := strconv.Atoi(matches[1])
		if err != nil {
			return nil, err
		}

		// TODO: support custom permalinks, and date format for
		// metadata
		info.SetPath(fmt.Sprintf("/%s/%d/%s.html",
			cfg.String("path", "listing"), id, matches[2]))

		//baseurl := sitecfg.String("url", "")
		baseurl := ""
		url, err := BuildURL(baseurl, info.Path())
		if err != nil {
			return nil, err
		}

		img, err := BuildURL(baseurl,
			fmt.Sprintf("/%s/%d/%s.jpg",
				cfg.String("path", "listing"),
				id, matches[2]))
		if err != nil {
			return nil, err
		}

		thumb, err := BuildURL(baseurl,
			fmt.Sprintf("/%s/%d/%s_thumb.jpg",
				cfg.String("path", "listing"),
				id, matches[2]))
		if err != nil {
			return nil, err
		}

		return &Listing{
			HTMLDocument: &HTMLDocument{ContentInfo: info},
			id:           id,
			url:          url,
			img:          img,
			thumb:        thumb,
			thumbWidth:   cfg.Int("thumb_width", 0),
			thumbHeight:  cfg.Int("thumb_height", 0),
		}, nil
	default:
		return nil, ErrIgnore
	}
	panic("unreachable")
}

func init() {
	RegisterGenerator("listing", GenerateListing)
}
