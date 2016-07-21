package main

import (
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/NSAlbuquerque/go-down-themes/models"
)

var (
	dirArg string

	GALERY_URL = "https://tmtheme-editor.herokuapp.com/gallery.json"
)

func init() {
	flag.StringVar(&dirArg, "to", ".", "Diretório destino")
	flag.Parse()
}

func main() {
	g := donwloadGalery()
	list := models.Galery{}

	for _, metatheme := range g {
		filepath, err := downloadTheme(metatheme, dirArg)
		if err != nil {
			log.Println(err)
			continue
		}

		m := models.ThemeMetaData{
			Name:  metatheme.Name,
			Url:   filepath,
			Light: metatheme.Light,
		}

		list = append(list, m)
	}
	replicarGalery(list)
}

func donwloadGalery() models.Galery {
	log.Printf("Baixando galeria de temas de %s...", GALERY_URL)
	resp, err := http.Get(GALERY_URL)
	if err != nil {
		log.Fatalln(err)
	}

	meta := models.Galery{}
	err = json.NewDecoder(resp.Body).Decode(&meta)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(" Feito")

	return meta
}

func downloadTheme(meta models.ThemeMetaData, dirDestino string) (string, error) {
	var folder = "light"
	if !meta.Light {
		folder = "dark"
	}

	err := os.MkdirAll(path.Join(dirDestino, folder), os.ModePerm|os.ModeType)
	if err != nil {
		log.Println("Falhou")
		return "", err
	}

	filepath := path.Join(dirDestino, folder, toFileName(meta.Name))

	log.Printf("Baixando tema %s para %s ", meta.Name, filepath)

	req, err := http.Get(meta.Url)
	if err != nil {
		log.Println("Falhou")
		return "", err
	}

	if req.StatusCode != http.StatusOK {
		log.Println("Falhou")
		return "", errors.New(req.Status)
	}

	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Println("Falhou")
		return "", err
	}

	ioutil.WriteFile(filepath, data, os.ModePerm|os.ModeType)
	log.Println("Feito")
	return filepath, nil
}

func toFileName(name string) (filename string) {
	filename = strings.NewReplacer("(", "", ")", "", "_", "-", " - ", "-", " ", "-").Replace(name)
	filename = strings.Title(filename) + ".tmTheme"
	return
}

func replicarGalery(g models.Galery) {
	log.Printf("Replicando... ")
	data, err := json.Marshal(g)

	err = ioutil.WriteFile("themes_meta.json", data, os.ModePerm|os.ModeType)
	if err != nil {
		log.Fatalln("Flaha de replicação")
	}
	log.Println("Feito")
}
