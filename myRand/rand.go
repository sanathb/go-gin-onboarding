package myRand

import (
	"crypto/rand"
	"fmt"
)

func RandToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
