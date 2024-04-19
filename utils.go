package main

import (
	"os"
	"strings"
)

func joinPath(p ...string) string {
	return strings.Join(p, "/")
}

func copyFile(src string, dst string) error {
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
