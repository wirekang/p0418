package main

import (
	"errors"
	"fmt"
	"path"
)

type format struct {
	id         string
	extensions []string
	vcodec     string
	acodec     string
}

var mp4 = format{
	id:         "mp4",
	extensions: []string{"mp4"},
	vcodec:     "libx264",
	acodec:     "copy",
}

var formats = []*format{&mp4}
var errUnknownFormat = errors.New("unknown format")

func getFormatById(id string) (*format, error) {
	for _, f := range formats {
		if f.id == id {
			return f, nil
		}
	}
	return nil, fmt.Errorf("invalid format id: %s", id)
}

func getFormatByFileName(fileName string) (*format, error) {
	for _, f := range formats {
		for _, e := range f.extensions {
			if path.Ext(fileName) == "."+e {
				return f, nil
			}
		}
	}
	return nil, errUnknownFormat
}

func (f *format) Args(repo *repository, v *video, eo editOptions) ([]string, error) {
	r := []string{"-vcodec", f.vcodec, "-acodec", f.acodec}
	return r, nil
}
