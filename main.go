package main

func main() {
	r, err := initRepository("data.json")
	if err != nil {
		panic(err)
	}
	_ = r
}
