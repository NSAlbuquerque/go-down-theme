package common

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
)

// ToFilemane retorna um nome amigável para arquivo a partir de um nome.
func ToFilemane(name string) (filename string) {
	filename = strings.NewReplacer("(", "", ")", "", "_", "-", " - ", "-", " ", "-").Replace(name)
	filename = strings.Title(filename)
	return
}

// Hash retorna o hash MD5 de uma string.
func Hash(s string) string {
	sum := md5.Sum([]byte(s))
	return string(hex.EncodeToString(sum[:]))
}
