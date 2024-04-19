package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"time"
	"unicode/utf8"

	"github.com/fatih/color"
	"github.com/rodaine/table"
	"github.com/wirekang/p0418/ytb"
)

type ffmpegMiddleware interface {
	Args(repo *repository, v *video, eo editOptions) ([]string, error)
}

type video struct {
	id                  int
	sourceFileName      string
	sourceFileCreatedAt int64
	category            *category
	format              *format
	createdAt           int64
	editedAt            int64
	url                 string
	uploadedAt          int64
}

func initVideos(repo *repository) ([]*video, error) {
	oldVideos, err := initOldVideos(repo)
	if err != nil {
		return nil, err
	}
	newVideos, err := initNewVideos(repo)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Videos(New): %d(%d)\n", len(oldVideos), len(newVideos))
	result := make([]*video, 0, len(oldVideos)+len(newVideos))
	result = append(result, oldVideos...)
	result = append(result, newVideos...)
	slices.SortStableFunc(result, func(a *video, b *video) int {
		return int(a.editedAt-b.editedAt)*100000 + int(a.uploadedAt-b.uploadedAt)
	})
	return result, nil
}

func initOldVideos(repo *repository) ([]*video, error) {
	videos := make([]*video, 0, len(repo.data.Videos))
	for _, videoData := range repo.data.Videos {
		category, err := getCategoryById(videoData.CategoryId)
		if err != nil {
			return nil, err
		}
		format, err := getFormatById(videoData.FormatId)
		if err != nil {
			return nil, err
		}
		videos = append(videos, &video{
			id:                  videoData.Id,
			sourceFileName:      videoData.SourceFileName,
			sourceFileCreatedAt: videoData.SourceFileCreatedAt,
			category:            category,
			format:              format,
			createdAt:           videoData.CreatedAt,
			editedAt:            videoData.EditedAt,
			url:                 videoData.Url,
			uploadedAt:          videoData.UploadedAt,
		})
	}
	return videos, nil
}

var ignoreFiles = []string{"desktop.ini"}

func initNewVideos(repo *repository) ([]*video, error) {
	nameId := make(map[string]int, len(repo.data.Videos))
	for _, v := range repo.data.Videos {
		nameId[v.SourceFileName] = v.Id
	}
	entries, err := os.ReadDir(repo.data.SourceFilesDir)
	if err != nil {
		return nil, err
	}
	result := []*video{}
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
		if slices.Contains(ignoreFiles, name) {
			continue
		}
		_, ok := nameId[name]
		if ok {
			continue
		}
		format, err := getFormatByFileName(name)
		if err != nil {
			if errors.Is(err, errUnknownFormat) {
				fmt.Println("Ignore unknwon format", name)
				continue
			}
		}
		category, err := getCategoryByFileName(name)
		if err != nil {
			if errors.Is(err, errUnknownCategory) {
				fmt.Println("Ignore unknwon category", name)
				continue
			}
			return nil, err
		}
		nextId := repo.data.NextId
		repo.data.NextId += 1
		fmt.Printf("New video %d: %s\n", nextId, name)
		v := &video{
			id:                  nextId,
			sourceFileName:      name,
			sourceFileCreatedAt: info.ModTime().Unix(),
			category:            category,
			format:              format,
			createdAt:           time.Now().Unix(),
			editedAt:            0,
			url:                 "",
			uploadedAt:          0,
		}
		result = append(result, v)
		repo.data.Videos = append(repo.data.Videos, videoData{
			Id:                  v.id,
			SourceFileName:      v.sourceFileName,
			SourceFileCreatedAt: v.sourceFileCreatedAt,
			CategoryId:          v.category.id,
			FormatId:            v.format.id,
			CreatedAt:           v.createdAt,
			EditedAt:            v.editedAt,
			Url:                 v.url,
			UploadedAt:          v.uploadedAt,
		})
		err = repo.save()
		if err != nil {
			return nil, err
		}
		err = copyFile(v.getSourceFileName(repo), v.getOriginalFileName(repo))
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func printVideos(videos []*video) {
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()
	tbl := table.New("id", "cat.", "sourceFileName", "edited", "uploaded")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt).WithWidthFunc(widthFunc)
	for _, v := range videos {
		tbl.AddRow(v.id, v.category.id, v.sourceFileName, colorBool(v.editedAt > 0), colorBool(v.uploadedAt > 0))
	}
	tbl.Print()
}

func colorBool(v bool) string {
	if !v {
		return color.New(color.FgRed).Sprint("☐")
	}
	return color.New(color.FgYellow).Sprint("☑")
}

var colorReg = regexp.MustCompile("\x1b[[0-9;]*m")

func widthFunc(s string) int {
	return utf8.RuneCountInString(colorReg.ReplaceAllString(s, ""))
}

func (v *video) getSourceFileName(repo *repository) string {
	return joinPath(repo.data.SourceFilesDir, v.sourceFileName)
}

func (v *video) getOriginalFileName(repo *repository) string {
	return joinPath(repo.data.OriginalFilesDir, fmt.Sprintf("%d.%s", v.id, v.format.extensions[0]))
}

func (v *video) getEditedFileName(repo *repository) string {
	return joinPath(repo.data.EditedFilesDir, fmt.Sprintf("%d.%s", v.id, v.format.extensions[0]))
}

func (v *video) edit(repo *repository, eo editOptions) error {
	start := time.Now()
	fmt.Println("Edit", v.id)
	args := []string{"-i", v.getOriginalFileName(repo), "-y", "-ss", fmt.Sprintf("00:00:%d", eo.start), "-to", fmt.Sprintf("00:00:%d", eo.end)}
	fargs, err := v.format.Args(repo, v, eo)
	if err != nil {
		return err
	}
	args = append(args, fargs...)
	cargs, err := v.category.Args(repo, v, eo)
	if err != nil {
		return err
	}
	args = append(args, cargs...)
	args = append(args, v.getEditedFileName(repo))
	c := exec.Command("ffmpeg", args...)
	b, err := c.CombinedOutput()
	if err != nil {
		logStrong("FFMPEG ERROR")
		fmt.Print(string(b))
		return err
	}
	duration := time.Since(start)
	fmt.Println("Success", duration)
	i, err := v.getRepoDataIndex(repo)
	if err != nil {
		return err
	}
	repo.data.Videos[i].EditedAt = time.Now().Unix()
	return repo.save()
}

func (v *video) getRepoDataIndex(repo *repository) (int, error) {
	for i, d := range repo.data.Videos {
		if d.Id == v.id {
			return i, nil
		}
	}
	return -1, fmt.Errorf("not in repo data")
}

func (v *video) upload(repo *repository) error {
	fmt.Println("Upload", v.id)
	url, err := ytb.Upload(ytb.UploadProps{
		Title:            v.category.youtubeTitle(v),
		Category:         v.category.youtubeCategory,
		Tags:             v.category.youtubeTags,
		Public:           true,
		File:             v.getEditedFileName(repo),
		ClientSecretFile: repo.data.YoutubeClientSecretFile,
	})
	if err != nil {
		return err
	}
	fmt.Println(url)
	i, err := v.getRepoDataIndex(repo)
	if err != nil {
		return err
	}
	repo.data.Videos[i].UploadedAt = time.Now().Unix()
	repo.data.Videos[i].Url = url
	return repo.save()
}
