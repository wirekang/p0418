package main

import (
	"strings"
)

func joinPath(p ...string) string {
	return strings.Join(p, "/")
}
