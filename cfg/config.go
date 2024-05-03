package cfg

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/wirekang/p0418/utils"
)

var FileName = "config.json"

const CategoryLol = "lol"

type Config struct {
	NextId                  int
	SourceFilesDir          string
	OriginalFilesDir        string
	OutputFilesDir          string
	YoutubeClientSecretFile string
	Categories              []Category
	Videos                  []Video
}

type Video struct {
	Id                  int
	SourceFileName      string
	SourceFileCreatedAt int64
	Extension           string
	CategoryId          string
	CreatedAt           int64
	EditedAt            *int64
	Url                 *string
	UploadedAt          *int64
	Range               *Range
}

type Category struct {
	Id                   string
	DefaultRange         Range
	EditOptions          EditOptions
	OriginalFilePrefixes []string
	YoutubeTags          []string
	YoutubeCategory      string
	YoutubeTitle         string
	Text                 string
}

type Range struct {
	Start int
	End   int
}

type EditOptions struct {
	OriginalWidth  int
	OriginalHeight int
	OutputHeight   int
	OutputRatio    float32
	FontFile       string
	FontColor      string
	FontSize       int
	PaddingX       int
	PaddingY       int
	PaddingColor   string
}

var Data = Config{
	NextId:                  1000,
	SourceFilesDir:          "FILLHERE",
	OriginalFilesDir:        "FILLHERE",
	OutputFilesDir:          "FILLHERE",
	YoutubeClientSecretFile: "FILLHERE",
	Videos:                  []Video{},
	Categories: []Category{
		{
			Id: CategoryLol,
			DefaultRange: Range{
				Start: 14,
				End:   29,
			},
			EditOptions: EditOptions{
				OriginalWidth:  1920,
				OriginalHeight: 1080,
				OutputHeight:   1920,
				OutputRatio:    1.7777777778,
				FontFile:       "C\\:/Windows/Fonts/arial.ttf",
				FontColor:      "white",
				FontSize:       48,
				PaddingX:       0,
				PaddingY:       200,
				PaddingColor:   "black",
			},
			OriginalFilePrefixes: []string{"League of Legends"},
			YoutubeTags:          []string{"league of legends"},
			YoutubeCategory:      "20",
			YoutubeTitle:         "{{.Id}} #leagueoflegends",
			Text:                 "{{.Id}}",
		},
	},
}

func Load() (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("loading config file: %w", err)
		}
	}()
	b, err := os.ReadFile(FileName)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			err2 := Save()
			if err2 != nil {
				return fmt.Errorf("inititializing config file: %w", err2)
			}
			fmt.Println("Config file created. Fill", FileName)
			return fmt.Errorf("launch again")
		}
		return err
	}
	return json.Unmarshal(b, &Data)
}

func Save() (err error) {
	temp, err := os.CreateTemp(os.TempDir(), "config")
	_ = temp.Close()
	_ = utils.Copy(FileName, temp.Name())
	defer func() {
		if err != nil {
			fmt.Println("CONFIG FILE BACKUPED:", temp.Name())
			err = fmt.Errorf("saving config file: %w", err)
		}
	}()
	b, err := json.MarshalIndent(Data, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.Create(FileName)
	if err != nil {
		return err
	}
	_, err = f.Write(b)
	if err != nil {
		return err
	}
	return f.Close()
}
