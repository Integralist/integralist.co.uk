package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	addr := ":8080"
	fmt.Printf("Serving at http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, http.FileServer(http.Dir("public"))))
}
