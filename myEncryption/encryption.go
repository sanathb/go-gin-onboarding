package myEncryption

import (
	"crypto/sha1"
	"fmt"
)

func GetSHA1(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}
