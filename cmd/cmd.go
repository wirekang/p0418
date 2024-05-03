package cmd

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"slices"
	"strings"

	"github.com/wirekang/p0418/cfg"
	"github.com/wirekang/p0418/vdo"
)

var commands = [](func() error){
	editOlderUnedited,
	editLatestEditedWithRange,
	editUnuploaded,
	uploadEditedAndUnuploaded,
	purgeOne,
	purgeUploaded,
	openOutputDir,
	exit,
}

func openOutputDir() error {
	exec.Command("explorer", strings.ReplaceAll(cfg.Data.OutputFilesDir, "/", "\\")).Run()
	return nil
}

func exit() error {
	os.Exit(0)
	return nil
}

func editOlderUnedited() error {
	video := sortVideos(func(v cfg.Video) int {
		if v.EditedAt == nil {
			return v.Id - 99999
		}
		return v.Id
	})[0]
	_ = video
	return vdo.Edit(video)
}

func editLatestEditedWithRange() error {
	video := sortVideos(func(v cfg.Video) int {
		if v.EditedAt == nil {
			return math.MaxInt
		}
		return int(*v.EditedAt)
	})[0]
	var start, end int
	fmt.Print("start end: ")
	fmt.Scanf("%d %d\n", &start, &end)
	video.Range = &cfg.Range{Start: start, End: end}
	for i := range cfg.Data.Videos {
		if cfg.Data.Videos[i].Id == video.Id {
			cfg.Data.Videos[i].Range = video.Range
		}
	}
	err := cfg.Save()
	if err != nil {
		return err
	}
	return vdo.Edit(video)
}

func editUnuploaded() error {
	for _, v := range cfg.Data.Videos {
		if v.UploadedAt != nil {
			continue
		}
		err := vdo.Edit(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func uploadEditedAndUnuploaded() error {
	for _, v := range cfg.Data.Videos {
		if v.UploadedAt != nil || v.EditedAt == nil {
			continue
		}
		err := vdo.Upload(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func purgeUploaded() error {
	for _, v := range cfg.Data.Videos {
		if v.UploadedAt == nil {
			continue
		}
		err := vdo.Purge(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func purgeOne() error {
	fmt.Print("id:")
	var id int
	fmt.Scanf("%d\n", &id)
	for i := range cfg.Data.Videos {
		if cfg.Data.Videos[i].Id == id {
			return vdo.Purge(cfg.Data.Videos[i])
		}
	}
	return fmt.Errorf("wrong id %d", id)
}

func sortVideos(f func(cfg.Video) int) []cfg.Video {
	videos := make([]cfg.Video, len(cfg.Data.Videos))
	copy(videos, cfg.Data.Videos)
	slices.SortFunc(videos, func(a, b cfg.Video) int {
		return f(a) - f(b)
	})
	return videos
}

func getFunctionName(temp interface{}) string {
	strs := strings.Split((runtime.FuncForPC(reflect.ValueOf(temp).Pointer()).Name()), ".")
	return strs[len(strs)-1]
}

func List() []string {
	r := []string{}
	for _, c := range commands {
		r = append(r, getFunctionName(c))
	}
	return r
}

func Run(i int) error {
	fmt.Println("Run command", getFunctionName(commands[i]))
	fmt.Println()
	return commands[i]()
}
