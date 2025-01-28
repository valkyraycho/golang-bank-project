package utils

import (
	"fmt"
	"math/rand"
	"strings"
)

const alphabets = "abcdefghijklmnopqrstuvwxyz"

func RandomString(length int) string {
	builder := strings.Builder{}

	k := len(alphabets)
	for i := 0; i < length; i++ {
		builder.WriteByte(alphabets[rand.Intn(k)])
	}

	return builder.String()
}

func RandomName() string {
	return RandomString(6)
}

func RandomEmail() string {
	return fmt.Sprintf("%s@gmail.com", RandomString(6))
}

func RandomCurrency() string {
	return SupportedCurrencies[rand.Intn(len(SupportedCurrencies))]
}

func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

func RandomMoney() int64 {
	return RandomInt(0, 1000)
}
