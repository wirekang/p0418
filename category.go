package main

import (
	"errors"
	"fmt"
	"strings"
)

type filterFunc func(repo *repository, video *video, eo editOptions) ([]string, error)

type category struct {
	id              string
	prefixes        []string
	filterFunc      filterFunc
	youtubeTags     []string
	youtubeCategory string
	youtubeTitle    func(v *video) string
}

var lol = category{
	id:              "lol",
	prefixes:        []string{"League of Legends"},
	youtubeTags:     []string{"league of legends", "fizz", "qiyana"},
	filterFunc:      lolFilterFunc,
	youtubeCategory: "20",
	youtubeTitle: func(v *video) string {
		return fmt.Sprintf("%d #leagueoflegends #qiyana #fizz", v.id)
	},
}

var categories = []*category{&lol}

var errUnknownCategory = errors.New("unknown category")

func getCategoryById(id string) (*category, error) {
	for _, c := range categories {
		if c.id == id {
			return c, nil
		}
	}
	return nil, fmt.Errorf("invalid category id: %s", id)
}

func getCategoryByFileName(fileName string) (*category, error) {
	for _, c := range categories {
		for _, p := range c.prefixes {
			if strings.HasPrefix(fileName, p) {
				return c, nil
			}
		}
	}
	return nil, errUnknownCategory
}

const inputWidth = 1920
const inputHeight = 1080
const youtubeShortsRatio = 9.0 / 16.0
const youtubeShortsHeight = 1920
const youtubeShortsWidth = int(youtubeShortsHeight * youtubeShortsRatio)
const lolPaddingWidth = max(inputWidth, youtubeShortsWidth)
const lolPaddingHeight = max(inputHeight, youtubeShortsHeight)
const lolPaddingX = 0
const lolPaddingY = 200
const lolPaddingColor = "black"
const lolCropWidth = youtubeShortsWidth
const lolCropHeight = youtubeShortsHeight
const lolCropX = (lolPaddingWidth - youtubeShortsWidth) / 2
const lolCropY = 0
const lolFontFile = "C\\:/Windows/Fonts/arial.ttf"
const lolFontColor = "white"
const lolFontSize = 48
const lolFontX = lolCropX + 16
const lolFontY = lolPaddingY + inputHeight + 8

func lolFilterFunc(repo *repository, v *video, eo editOptions) ([]string, error) {
	text := fmt.Sprintf("KR Server/Unranked\n%d", v.id)
	r := []string{}
	r = append(r, fmt.Sprintf("pad=%d:%d:%d:%d:%s", lolPaddingWidth, lolPaddingHeight, lolPaddingX, lolPaddingY, lolPaddingColor))
	r = append(r, fmt.Sprintf("drawtext=fontfile='%s':text='%s':fontcolor='%s':fontsize=%d:x=%d:y=%d:line_spacing=-10", lolFontFile, text, lolFontColor, lolFontSize, lolFontX, lolFontY))
	r = append(r, fmt.Sprintf("crop=%d:%d:%d:%d", lolCropWidth, lolCropHeight, lolCropX, lolCropY))
	return r, nil
}

func (c *category) Args(repo *repository, v *video, eo editOptions) ([]string, error) {
	filters, err := c.filterFunc(repo, v, eo)
	if err != nil {
		return nil, err
	}
	r := []string{"-vf"}
	r = append(r, strings.Join(filters, ","))
	return r, nil
}
