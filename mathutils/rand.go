package mathutils

import (
	// "fmt"
	"math/rand"
	"time"
)

var (
	letters    = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	lowerChars = []rune("abcdefghijklmnopqrstuvwxyz")
	upperChars = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	num        = []rune("0123456789")
)

func RandomNumString(n int) string {
	rand.Seed(time.Now().UnixNano())

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(num))]
	}
	return string(b)
}

func RandomNumber(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

func RandomLowerCharString(n int) string {
	rand.Seed(time.Now().UnixNano())

	b := make([]rune, n)
	for i := range b {
		b[i] = lowerChars[rand.Intn(len(num))]
	}
	return string(b)
}

func RandomUpperCharString(n int) string {
	rand.Seed(time.Now().UnixNano())

	b := make([]rune, n)
	for i := range b {
		b[i] = upperChars[rand.Intn(len(num))]
	}
	return string(b)
}
