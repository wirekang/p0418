package utils

import (
	"bytes"
	"fmt"
	"os"
	"text/template"
)

func Copy(src string, dst string) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("copying file: %w", err)
		}
	}()
	r, err := os.Open(src)
	if err != nil {
		return err
	}
	defer r.Close()
	w, err := os.Create(dst)
	if err != nil {
		return err
	}
	_, err = w.ReadFrom(r)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return nil
}

func MkdirAll(dirs ...string) error {
	for _, dir := range dirs {
		err := os.MkdirAll(dir, os.ModeDir)
		if err != nil {
			return err
		}
	}
	return nil
}

func TemplateString(t string, d any) (string, error) {
	tt, err := template.New("t").Parse(t)
	if err != nil {
		return "", err
	}
	b := bytes.NewBufferString("")
	err = tt.Execute(b, d)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}
