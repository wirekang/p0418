package cat

import (
	"fmt"
	"strings"

	"github.com/wirekang/p0418/cfg"
	"github.com/wirekang/p0418/utils"
)

var ErrUnknownCategory = fmt.Errorf("unknown video category")

func GetCategoryBySourceFileName(f string) (cfg.Category, error) {
	for _, c := range cfg.Data.Categories {
		for _, p := range c.OriginalFilePrefixes {
			if strings.HasPrefix(f, p) {
				return c, nil
			}
		}
	}
	return cfg.Category{}, ErrUnknownCategory
}

func GetCategoryById(id string) (cfg.Category, error) {
	for _, c := range cfg.Data.Categories {
		if c.Id == id {
			return c, nil
		}
	}
	return cfg.Category{}, ErrUnknownCategory
}

func makeFilters(c cfg.Category, v any) ([]string, error) {
	var outputH = c.EditOptions.OutputHeight
	var outputW = int(float32(outputH) / c.EditOptions.OutputRatio)
	var padW = max(c.EditOptions.OriginalWidth, outputW)
	var padH = max(c.EditOptions.OriginalWidth, outputH)
	var padX = c.EditOptions.PaddingX
	var padY = c.EditOptions.PaddingY
	var padColor = c.EditOptions.PaddingColor
	var cropW = outputW
	var cropH = outputH
	var cropX = (padW - outputW) / 2
	var cropY = 0
	var fontFile = c.EditOptions.FontFile
	var fontColor = c.EditOptions.FontColor
	var fontSize = c.EditOptions.FontSize
	var textX = cropX + 16
	var textY = padY + c.EditOptions.OriginalHeight + 8

	text, err := utils.TemplateString(c.Text, v)
	if err != nil {
		return nil, err
	}
	r := []string{}
	r = append(r, fmt.Sprintf("pad=%d:%d:%d:%d:%s", padW, padH, padX, padY, padColor))
	r = append(r, fmt.Sprintf("drawtext=fontfile='%s':text='%s':fontcolor='%s':fontsize=%d:x=%d:y=%d:line_spacing=-10", fontFile, text, fontColor, fontSize, textX, textY))
	r = append(r, fmt.Sprintf("crop=%d:%d:%d:%d", cropW, cropH, cropX, cropY))
	return r, nil
}

func FfmpegArgs(c cfg.Category, v any, overrideRange *cfg.Range) (args []string, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("making ffmpeg args from category: %w", err)
		}
	}()
	filters, err := makeFilters(c, v)
	if err != nil {
		return nil, err
	}
	r := c.DefaultRange
	if overrideRange != nil {
		r = *overrideRange
	}
	args = []string{"-ss", fmt.Sprintf("00:00:%d", r.Start), "-to", fmt.Sprintf("00:00:%d", r.End), "-vf", strings.Join(filters, ",")}
	return args, nil
}
