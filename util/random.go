package util

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

const alphabet = "abcdefghijklmnnopqrstuvwxyz"

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

// RandomInt generates a random integer between min and max
func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

// RandomString generates a random string of length n
func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabet)

	for i := 0; i < n; i++ {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}

// RandomName generates a random owner name
func RandomName() string {
	return fmt.Sprintf("%s %s", RandomString(5), RandomString(5))
}

// Random Phone generates a random phone number
func RandomPhone() string {
	return fmt.Sprintf("+%d %d", RandomInt(1, 999), RandomInt(10000000, 99999999))
}

// RandomEmail generates a random email
func RandomEmail() string {
	return fmt.Sprintf("%s@email.com", RandomString(8))
}

// RandomAddress generates a random address
func RandomAddress() string {
	return fmt.Sprintf("%d %s St", RandomInt(1, 999), RandomString(5))
}

// RandomStatus generates a random status
func RandomStatus() string {
	status := []string{DRAFT, PENDING_PAYMENT, PAID, OVERDUE}
	return status[rand.Intn(len(status))]
}

// RandomFloatN generates a random float64 value in the interval [0.0, n)
func RandomFloatN(n float64) float64 {
	return rand.Float64() * n
}
