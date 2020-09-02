package main

import (
	"flag"
	"log"
)

var (
	dirArg string
)

func init() {
	flag.StringVar(&dirArg, "to", ".", "Diretório destino")
	flag.Parse()
}

func main() {
	log.Println(dirArg)
}
