package main

import (
	"log"
	"os"

	"github.com/integralist/integralist.co.uk/internal/builder"
)

func main() {
	b := builder.New("content", "assets", "public", "https://www.integralist.co.uk")
	if err := b.Build(); err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}
