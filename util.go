package grout

import (
	"net/url"
	"time"
)

func BuildURL(root, path string) (string, error) {
	if root == "" {
		root = "/"
	}
	u, err := url.Parse(root)
	if err != nil {
		return "", err
	}
	u, err = u.Parse(path)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func XMLDate(t time.Time) string {
	return t.Format("2006-01-02T15:04:05-07:00")
}
