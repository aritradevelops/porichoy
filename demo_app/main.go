package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
)

//go:embed public/*
var publicFS embed.FS

func main() {
	subFS, err := fs.Sub(publicFS, "public")
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/", http.FileServer(http.FS(subFS)))
	http.ListenAndServe(":5000", nil)
}
