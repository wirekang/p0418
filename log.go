package main

import "fmt"

const strongBorder = "=============================="

func logStrong(e ...any) {
	e2 := []any{strongBorder + "\n\n"}
	e2 = append(e2, e...)
	e2 = append(e2, "\n\n"+strongBorder)
	fmt.Println(e2...)
}
