package main

import (
	"errors"
	"path"
)

type format struct {
	extensions []string
}

var mp4 = format{
	extensions: []string{"mp4"},
}

var formats = []*format{&mp4}
var errUnknownFormat = errors.New("unknown format")

func getFormat(fileName string) (*format, error) {
	for _, f := range formats {
		for _, e := range f.extensions {
			if path.Ext(fileName) == e {
				return f, nil
			}
		}
	}
	return nil, errUnknownFormat
}
