package main

import (
	"flag"
	"log"
)

var (
	dirArg string
)

const galleryURL = "https://tmtheme-editor.herokuapp.com/gallery.json"

func init() {
	flag.StringVar(&dirArg, "to", ".", "Diretório destino")
	flag.Parse()
}

func main() {
	log.Println(dirArg, galleryURL)
}
