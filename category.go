package main

import (
	"errors"
	"strings"
)

type category struct {
	prefixes []string
}

var lol = category{
	prefixes: []string{"League of Legends"},
}

var categories = []*category{&lol}

var errUnknownCategory = errors.New("unknown category")

func getCategory(fileName string) (*category, error) {
	for _, c := range categories {
		for _, p := range c.prefixes {
			if strings.HasPrefix(fileName, p) {
				return c, nil
			}
		}
	}
	return nil, errUnknownCategory
}
