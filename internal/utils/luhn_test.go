package utils

import "testing"

func TestIsValidLuhn(t *testing.T) {
	testCases := []struct {
		number   string
		expected bool
		desc     string
	}{
		{"4532015112830366", true, "Valid Visa card number"},
		{"6011514433546201", true, "Valid Discover card number"},
		{"6011514433546202", false, "Invalid card number (changed last digit)"},
		{"4532015112830367", false, "Invalid Visa card number"},
		{"12345678903", true, "Valid order number from spec"},
		{"123456789", false, "Invalid order number"},
		{"", false, "Empty string"},
		{"12345abcde", false, "Contains non-digits"},
		{"79927398713", true, "Another valid number"},
		{"79927398714", false, "Invalid version of above"},
		{"0", true, "Single digit 0"},
		{"1", false, "Single digit 1"},
		{"18", true, "Valid two-digit number"},
		{"19", false, "Invalid two-digit number"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result := IsValidLuhn(tc.number)
			if result != tc.expected {
				t.Errorf("IsValidLuhn(%s) = %v; expected %v", tc.number, result, tc.expected)
			}
		})
	}
}

func TestLuhnChecksum(t *testing.T) {
	testCases := []struct {
		number   string
		expected int
		desc     string
	}{
		{"4532015112830366", 0, "Valid number should have checksum 0"},
		{"4532015112830367", 1, "Invalid number should have non-zero checksum"},
		{"12345678903", 0, "Valid order number should have checksum 0"},
		{"123456789", 7, "Invalid number should have non-zero checksum"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result := luhnChecksum(tc.number)
			if result != tc.expected {
				t.Errorf("luhnChecksum(%s) = %d; expected %d", tc.number, result, tc.expected)
			}
		})
	}
}
