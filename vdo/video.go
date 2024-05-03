package vdo

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"slices"
	"time"

	"github.com/wirekang/p0418/cat"
	"github.com/wirekang/p0418/cfg"
	"github.com/wirekang/p0418/utils"
	"github.com/wirekang/p0418/ytb"
)

var ignoredFiles = []string{"desktop.ini"}

func Load() (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("loading videos: %w", err)
		}
	}()
	nameId := makeNameId()
	err = walkSoureFiles(func(i fs.FileInfo) error {
		name := i.Name()
		_, ok := nameId[name]
		if ok {
			return nil
		}
		c, err := cat.GetCategoryBySourceFileName(name)
		if err != nil {
			if errors.Is(err, cat.ErrUnknownCategory) {
				return nil
			}
			return err
		}
		id := cfg.Data.NextId
		ext := path.Ext(name)
		fmt.Printf("New video (%d): %s\n", id, name)
		cfg.Data.Videos = append(cfg.Data.Videos, cfg.Video{
			Id:                  id,
			Extension:           ext,
			SourceFileName:      name,
			SourceFileCreatedAt: i.ModTime().Unix(),
			CategoryId:          c.Id,
			CreatedAt:           time.Now().Unix(),
		})
		cfg.Data.NextId += 1
		err = cfg.Save()
		if err != nil {
			return err
		}
		err = utils.Copy(path.Join(cfg.Data.SourceFilesDir, name), path.Join(cfg.Data.OriginalFilesDir, fmt.Sprintf("%d%s", id, ext)))
		return err
	})
	if err != nil {
		return err
	}
	return nil
}

func makeNameId() map[string]int {
	nameId := make(map[string]int, len(cfg.Data.Videos))
	for _, v := range cfg.Data.Videos {
		nameId[v.SourceFileName] = v.Id
	}
	return nameId
}

func walkSoureFiles(f func(i fs.FileInfo) error) error {
	entries, err := os.ReadDir(cfg.Data.SourceFilesDir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		i, err := e.Info()
		if err != nil {
			return err
		}
		if i.IsDir() {
			continue
		}
		if slices.Contains(ignoredFiles, i.Name()) {
			continue
		}
		err = f(i)
		if err != nil {
			return err
		}
	}
	return nil
}

func ffmpegArgs(v cfg.Video) (args []string, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("making ffmpeg args from video: %w", err)
		}
	}()

	c, err := cat.GetCategoryById(v.CategoryId)
	if err != nil {
		return nil, err
	}
	cargs, err := cat.FfmpegArgs(c, v, v.Range)
	if err != nil {
		return nil, err
	}

	args = []string{"-i", path.Join(cfg.Data.OriginalFilesDir, fmt.Sprintf("%d%s", v.Id, v.Extension)), "-y", "-vcodec", "libx264", "-acodec", "copy"}
	args = append(args, cargs...)
	args = append(args, path.Join(cfg.Data.OutputFilesDir, fmt.Sprintf("%d%s", v.Id, v.Extension)))
	return args, nil
}

func Edit(v cfg.Video) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("editting video: %w", err)
		}
	}()
	start := time.Now()
	fmt.Println("Edit", v.Id)
	args, err := ffmpegArgs(v)
	if err != nil {
		return err
	}
	c := exec.Command("ffmpeg", args...)
	b, err := c.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s\n: %w", string(b), err)
	}
	duration := time.Since(start)
	fmt.Println("Success", duration)
	for i := range cfg.Data.Videos {
		if cfg.Data.Videos[i].Id == v.Id {
			now := time.Now().Unix()
			cfg.Data.Videos[i].EditedAt = &now
		}
	}
	err = cfg.Save()
	if err != nil {
		return err
	}
	return nil
}

func Upload(v cfg.Video) (err error) {
	fmt.Println("Upload", v.Id)
	defer func() {
		if err != nil {
			err = fmt.Errorf("uploading video: %w", err)
		}
	}()
	var confirm int
	fmt.Print("type video id to confirm: ")
	fmt.Scanf("%d\n", &confirm)
	if confirm != v.Id {
		return fmt.Errorf("confirm failed")
	}
	c, err := cat.GetCategoryById(v.CategoryId)
	if err != nil {
		return err
	}
	title, err := utils.TemplateString(c.YoutubeTitle, v)
	if err != nil {
		return err
	}
	url, err := ytb.Upload(ytb.UploadProps{
		Title:            title,
		Description:      "",
		Category:         c.YoutubeCategory,
		Tags:             c.YoutubeTags,
		Public:           true,
		File:             path.Join(cfg.Data.OutputFilesDir, fmt.Sprintf("%d%s", v.Id, v.Extension)),
		ClientSecretFile: cfg.Data.YoutubeClientSecretFile,
	})
	if err != nil {
		return err
	}
	fmt.Println("Success")
	for i := range cfg.Data.Videos {
		if cfg.Data.Videos[i].Id == v.Id {
			now := time.Now().Unix()
			cfg.Data.Videos[i].UploadedAt = &now
			cfg.Data.Videos[i].Url = &url
		}
	}
	return cfg.Save()
}

func Purge(v cfg.Video) error {
	fmt.Println("Purge", v.Id)
	videos := make([]cfg.Video, 0)
	for i := range cfg.Data.Videos {
		if cfg.Data.Videos[i].Id == v.Id {
			os.Remove(path.Join(cfg.Data.SourceFilesDir, v.SourceFileName))
			os.Remove(path.Join(cfg.Data.OriginalFilesDir, fmt.Sprintf("%d%s", v.Id, v.Extension)))
			os.Remove(path.Join(cfg.Data.OutputFilesDir, fmt.Sprintf("%d%s", v.Id, v.Extension)))
			continue
		}
		videos = append(videos, cfg.Data.Videos[i])
	}
	cfg.Data.Videos = videos
	return cfg.Save()
}
