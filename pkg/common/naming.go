package common

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
)

// ToFilemane returns the friendly name for the theme from the name of its file.
func ToFilemane(name string) (filename string) {
	filename = strings.NewReplacer("(", "", ")", "", "_", "-", " - ", "-", " ", "-").Replace(name)
	filename = strings.Title(filename)
	return
}

// Hash returns the MD5 hash of a string.
func Hash(s string) string {
	sum := md5.Sum([]byte(s))
	return string(hex.EncodeToString(sum[:]))
}
