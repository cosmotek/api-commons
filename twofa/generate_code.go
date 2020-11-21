package twofa

import (
	"math/rand"
	"strconv"
	"time"
)

const alphanumericKeyBank = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890"

// init is run when the package is imported (prior to app's main function)
func init() {
	rand.Seed(time.Now().UnixNano())
}

// GenerateNumericPincode generates a numeric pincode of
// the provided length for use in SMS verification.
func GenerateNumericPincode(length int) string {
	output := ""
	for i := 0; i < length; i++ {
		char := strconv.Itoa(rand.Intn(10))
		output += char
	}
	return output
}

// GenerateAlphaNumericCode generates a alpha-numeric pincode of
// the provided length for use in Email verification.
func GenerateAlphaNumericCode(length int) string {
	output := ""
	for i := 0; i < length; i++ {
		keyindex := rand.Intn(len(alphanumericKeyBank))
		output += string(alphanumericKeyBank[keyindex])
	}
	return output
}
