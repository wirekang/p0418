package ctl

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"slices"
	"time"
	"unicode/utf8"

	"github.com/fatih/color"
	"github.com/rodaine/table"
	"github.com/wirekang/p0418/cat"
	"github.com/wirekang/p0418/cfg"
)

func Start(cmds []string, runCmd func(int) error) error {
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("")
		printVideos()
		printCmd(cmds)
		fmt.Print(">> ")
		c, _ := r.ReadString('\n')
		c = string([]byte(c)[0])
		i := slices.Index(mappings, c)
		if i == -1 || i >= len(cmds) {
			fmt.Println("unknown", c)
			continue
		}
		err := runCmd(i)
		if err != nil {
			fmt.Println("Error", err)
		}
	}
}

var mappings = []string{"q", "w", "e", "r", "t", "a", "s", "d", "f", "g", "z", "x", "c", "v", "b"}

func printCmd(cmds []string) {
	for i := range cmds {
		fmt.Printf("[%s] %s\n", mappings[i], cmds[i])
	}
}

var colorReg = regexp.MustCompile("\x1b[[0-9;]*m")

func widthFunc(s string) int {
	return utf8.RuneCountInString(colorReg.ReplaceAllString(s, ""))
}

func printVideos() {
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()
	tbl := table.New("id", "cat.", "sourceFileName", "createdAt", "editedAt", "uploadedAt", "range")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt).WithWidthFunc(widthFunc)
	for _, v := range cfg.Data.Videos {
		c, _ := cat.GetCategoryById(v.CategoryId)
		r := c.DefaultRange
		if v.Range != nil {
			r = *v.Range
		}
		tbl.AddRow(v.Id, v.CategoryId, v.SourceFileName, formatTime(&v.CreatedAt), formatTime(v.EditedAt), formatTime(v.UploadedAt), r)
	}
	tbl.Print()
}

func formatTime(t *int64) string {
	if t == nil {
		return "-"
	}
	tt := time.Unix(*t, 0)
	return tt.Format("0102:1504")
}

func formatBool(v bool) string {
	if !v {
		return color.New(color.FgRed).Sprint("☐")
	}
	return color.New(color.FgYellow).Sprint("☑")
}
