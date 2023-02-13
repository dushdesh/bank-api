package util

import (
	"math/rand"
	"strings"
	"time"
)

var alphabets = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ!@#$%^&*()"

func init() {
	rand.Seed(time.Now().UnixNano())
}

// RandomInt generates a random number between min and max
func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

// RandomString generartes a random string of n chars
func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabets)

	for i := 0; i < n; i++ {
		c := alphabets[rand.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}

// RandomOwner genrates a random owner id
func RandomOwner() string {
	return RandomString(6)
}

// RandomAmount genrated random amounts
func RandomAmount() int64 {
	return RandomInt(0, 1000)
}

// RandomCurrency generates randomm currency codes
func RandomCurrency() string {
	currencies := []string{"CAD", "EUR", "USD", "INR"}
	n := len(currencies)
	return currencies[rand.Intn(n)]
}
