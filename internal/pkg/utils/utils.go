package utils

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

func GenerateID(prefix string) string {
	raw := uuid.New().String()
	clean := ""
	for _, c := range raw {
		if c != '-' {
			clean += string(c)
		}
	}
	if len(clean) > 8 {
		clean = clean[:8]
	}
	return fmt.Sprintf("%s_%s", prefix, clean)
}

func GenerateNumericID(prefix string) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	num := r.Intn(900000) + 100000
	return fmt.Sprintf("%s_%d", prefix, num)
}
