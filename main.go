package main

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
)

func main() {
	for {
		r, err := initRepository("data.json")
		if err != nil {
			panic(err)
		}
		err = os.MkdirAll(r.data.OriginalFilesDir, os.ModeDir)
		if err != nil {
			panic(err)
		}
		err = os.MkdirAll(r.data.EditedFilesDir, os.ModeDir)
		if err != nil {
			panic(err)
		}
		videos, err := initVideos(r)
		if err != nil {
			panic(err)
		}
		ctx := commandContext{
			repo:        r,
			videos:      videos,
			editOptions: editOptions{start: r.data.EditStart, end: r.data.EditEnd},
		}
		printVideos(videos)
		printEditOptions(ctx.editOptions)
		sel := &survey.Select{
			Options: commandIds,
			VimMode: true,
		}
		var cmdId string
		survey.AskOne(sel, &cmdId)
		err = runCommand(&ctx, cmdId)
		if err != nil {
			panic(err)
		}
	}
}

func printEditOptions(eo editOptions) {
	fmt.Printf("%d-%d\n", eo.start, eo.end)
}
