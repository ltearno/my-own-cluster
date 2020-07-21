package tools

import (
	"crypto/sha256"
	"fmt"
	"regexp"
)

var IsLetter = regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString

func Min(a int, b int) int {
	if a <= b {
		return a
	}

	return b
}

func Sha256Sum(bytes []byte) string {
	crc := sha256.New()
	crc.Write(bytes)
	sha256Bytes := crc.Sum(nil)
	return fmt.Sprintf("%x", sha256Bytes)
}

func SimplifyHeaders(in map[string][]string) map[string]string {
	o := make(map[string]string)
	for k, v := range in {
		o[k] = v[0]
	}
	return o
}
