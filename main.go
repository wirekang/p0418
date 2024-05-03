package main

import (
	"fmt"

	"github.com/wirekang/p0418/cfg"
	"github.com/wirekang/p0418/cmd"
	"github.com/wirekang/p0418/ctl"
	"github.com/wirekang/p0418/utils"
	"github.com/wirekang/p0418/vdo"
)

func m() error {
	err := cfg.Load()
	if err != nil {
		return err
	}
	err = utils.MkdirAll(cfg.Data.OriginalFilesDir, cfg.Data.OutputFilesDir)
	if err != nil {
		return err
	}
	err = vdo.Load()
	if err != nil {
		return err
	}
	return ctl.Start(cmd.List(), cmd.Run)
}

func main() {
	err := m()
	if err != nil {
		fmt.Println("ERROR\n", err)
	}
}
