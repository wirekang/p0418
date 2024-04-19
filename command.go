package main

import (
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

type commandContext struct {
	repo        *repository
	videos      []*video
	editOptions editOptions
}

type command func(ctx *commandContext) error

var commands = map[string]command{
	"0:editAll":      cmdEditAll,
	"1:editSelect":   cmdEditSelect,
	"2:setRange":     cmdSetRange,
	"3:openDir":      cmdOpenDir,
	"5:uploadSelect": cmdUploadSelect,
	"6:purgeSelect":  cmdPurgeSelect,
	"9:exit":         cmdExit,
}

func cmdIdOrderFunc(id string) int {
	i, _ := strconv.ParseInt(strings.Split(id, ":")[0], 10, 32)
	return int(i)
}

var commandIds = (func() []string {
	r := make([]string, 0, len(commands))
	for k := range commands {
		r = append(r, k)
	}
	slices.SortFunc(r, func(a, b string) int {
		return cmdIdOrderFunc(a) - cmdIdOrderFunc(b)
	})
	return r
})()

var cmdExit command = func(ctx *commandContext) error {
	os.Exit(0)
	return nil
}

var cmdOpenDir command = func(ctx *commandContext) error {
	exec.Command("explorer", strings.ReplaceAll(ctx.repo.data.EditedFilesDir, "/", "\\")).Run()
	return nil
}

var cmdEditAll command = func(ctx *commandContext) error {
	v := []*video{}
	for _, vv := range ctx.videos {
		if vv.editedAt == 0 {
			v = append(v, vv)
		}
	}
	return edit(ctx, v)
}

var cmdEditSelect command = func(ctx *commandContext) error {
	v, err := selectVideos(ctx.videos)
	if err != nil {
		return err
	}
	err = edit(ctx, v)
	return err
}

var cmdPurgeSelect command = func(ctx *commandContext) error {
	logStrong("PURGE SELECT")
	v, err := selectVideos(ctx.videos)
	if err != nil {
		return err
	}
	err = purge(ctx, v)
	return err
}

var cmdSetRange command = func(ctx *commandContext) error {
	fmt.Print("[Start] [End]:")
	var start, end int
	fmt.Scanf("%d %d\n", &start, &end)
	ctx.editOptions = editOptions{
		start: start,
		end:   end,
	}
	ctx.repo.data.EditStart = start
	ctx.repo.data.EditEnd = end
	return ctx.repo.save()
}

var cmdUploadSelect command = func(ctx *commandContext) error {
	v, err := selectVideos(ctx.videos)
	if err != nil {
		return err
	}
	return upload(ctx, v)
}

func runCommand(ctx *commandContext, id string) error {
	return commands[id](ctx)
}

func edit(ctx *commandContext, videos []*video) error {
	for _, v := range videos {
		err := v.edit(ctx.repo, ctx.editOptions)
		if err != nil {
			return err
		}
	}
	return nil
}

func purge(ctx *commandContext, videos []*video) error {
	for _, v := range videos {
		err := os.Remove(v.getSourceFileName(ctx.repo))
		if err != nil {
			logStrong("Failed to remove source file:", v.id, err)
		}
		err = os.Remove(v.getOriginalFileName(ctx.repo))
		if err != nil {
			logStrong("Failed to remove original file:", v.id, err)
		}
		if v.editedAt > 0 {
			err = os.Remove(v.getEditedFileName(ctx.repo))
			if err != nil {
				logStrong("Failed to remove edited file:", v.id, err)
			}
		}
		i := slices.IndexFunc(ctx.repo.data.Videos, func(vd videoData) bool {
			return vd.Id == v.id
		})
		if i == -1 {
			return fmt.Errorf("video not in repository: %d", v.id)
		}
		ctx.repo.data.Videos = slices.Delete(ctx.repo.data.Videos, i, i+1)
		err = ctx.repo.save()
		if err != nil {
			return err
		}
		logStrong(v.id, "PURGED")
	}
	return fmt.Errorf("PLEASE RESTART")
}

func selectVideos(videos []*video) ([]*video, error) {
	videoBySid := map[string]*video{}
	options := []string{}
	descriptions := []string{}
	for _, v := range videos {
		sid := fmt.Sprintf("%d", v.id)
		options = append(options, sid)
		descriptions = append(descriptions, fmt.Sprintf("%s - %s", v.category.id, v.sourceFileName))
		videoBySid[sid] = v
	}
	p := &survey.MultiSelect{
		Message: "Select Videos",
		VimMode: true,
		Options: options,
		Description: func(value string, index int) string {
			return descriptions[index]
		},
	}
	sids := []string{}
	survey.AskOne(p, &sids)
	r := []*video{}
	for _, sid := range sids {
		r = append(r, videoBySid[sid])
	}
	return r, nil
}

func upload(ctx *commandContext, videos []*video) error {
	for _, v := range videos {
		err := v.upload(ctx.repo)
		if err != nil {
			return err
		}
	}
	return nil
}
