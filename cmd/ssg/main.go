package main

import (
	"log"
	"os"

	"lostsgnl.com/internal/builder"
)

func main() {
	b := builder.New("content", "assets", "public", "https://lostsgnl.com")
	if err := b.Build(); err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}
