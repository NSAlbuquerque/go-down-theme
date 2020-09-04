package common

import "strings"

// ToFilemane retorna um nome amigável para arquivo a partir de um nome.
func ToFilemane(name string) (filename string) {
	filename = strings.NewReplacer("(", "", ")", "", "_", "-", " - ", "-", " ", "-").Replace(name)
	filename = strings.Title(filename)
	return
}
