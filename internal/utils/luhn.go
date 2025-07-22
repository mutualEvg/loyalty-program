package utils

import (
	"strconv"
	"strings"
)

// IsValidLuhn validates a number using the Luhn algorithm
func IsValidLuhn(number string) bool {
	// Remove any spaces and validate that it contains only digits
	number = strings.ReplaceAll(number, " ", "")
	if len(number) == 0 {
		return false
	}

	for _, char := range number {
		if char < '0' || char > '9' {
			return false
		}
	}

	return luhnChecksum(number) == 0
}

func luhnChecksum(number string) int {
	var sum int
	var alternate bool

	// Process digits from right to left
	for i := len(number) - 1; i >= 0; i-- {
		digit, _ := strconv.Atoi(string(number[i]))

		if alternate {
			digit *= 2
			if digit > 9 {
				digit = digit/10 + digit%10
			}
		}

		sum += digit
		alternate = !alternate
	}

	return sum % 10
}
