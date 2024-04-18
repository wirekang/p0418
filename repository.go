package main

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
)

type repository struct {
	File string
	data repositoryData
}

type repositoryData struct {
	NextId                  int
	SourceFilesDir          string
	OriginalFilesDir        string
	EditedFilesDir          string
	YoutubeClientSecretFile string
}

func initRepository(file string) (*repository, error) {
	var d repositoryData
	b, err := os.ReadFile(file)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			logStrong("fresh start detected\nfill empty fields in", file)
			r := &repository{file, d}
			r.data.NextId = 10001
			err = r.save()
			if err != nil {
				return nil, err
			}
			return r, errors.New("launch again")
		}
		return nil, err
	}
	err = json.Unmarshal(b, &d)
	if err != nil {
		return nil, err
	}
	return &repository{file, d}, nil
}

func (r *repository) save() error {
	b, err := json.MarshalIndent(r.data, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.Create(r.File)
	if err != nil {
		return err
	}
	_, err = f.Write(b)
	if err != nil {
		return err
	}
	return f.Close()
}
