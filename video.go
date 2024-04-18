package main

import (
	"errors"
	"fmt"
	"os"
	"time"
)

type cut struct {
	start int
	end   int
}

type editOptions struct {
	cut cut
}

type video struct {
	id              int
	sourceFileName  string
	sourceCreatedAt int64
	category        *category
	format          *format
	createdAt       int64
	editOptions     *editOptions
	editedAt        int64
	url             string
	uploadedAt      int64
	purged          bool
}

func initVideos(repo repository) ([]video, error) {
	entries, err := os.ReadDir(repo.data.SourceFilesDir)
	if err != nil {
		return nil, err
	}
	newVideos := []video{}
	for _, e := range entries {
		if e.IsDir() {
			fmt.Println("Ignore dir", e.Name())
			continue
		}
		info, err := e.Info()
		if err != nil {
			return nil, err
		}
		name := info.Name()
		format, err := getFormat(name)
		if err != nil {
			if errors.Is(err, errUnknownFormat) {
				fmt.Println("Ignore unknwon format", name)
				continue
			}
		}

		category, err := getCategory(name)
		if err != nil {
			if errors.Is(err, errUnknownCategory) {
				fmt.Println("Ignore unknwon category", name)
				continue
			}
			return nil, err
		}
		newVideos = append(newVideos, video{
			id:              0,
			sourceFileName:  name,
			sourceCreatedAt: info.ModTime().Unix(),
			category:        category,
			format:          format,
			createdAt:       time.Now().Unix(),
			editOptions:     nil,
			editedAt:        0,
			url:             "",
			uploadedAt:      0,
			purged:          false,
		})
	}
}
