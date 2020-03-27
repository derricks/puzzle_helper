package cmd

import (
	"testing"
)

func TestIsUppercaseAscii(test *testing.T) {
	expectedResults := map[byte]bool{
		'A': true,
		'Z': true,
		'M': true,
		'a': false,
		'[': false,
		'@': false,
	}

	for toTest, expected := range expectedResults {
		if isUppercaseAscii(toTest) != expected {
			test.Errorf("Expected isUppercaseAscii(%c) to return %v but returned %v instead", toTest, expected, isUppercaseAscii(toTest))
		}
	}
}

func TestIsLowercaseAscii(test *testing.T) {
	expectedResults := map[byte]bool{
		'a': true,
		'A': false,
		'z': true,
		'{': false,
		'`': false,
	}

	for toTest, expected := range expectedResults {
		if isLowercaseAscii(toTest) != expected {
			test.Errorf("Expected isLowercaseAscii(%c) to return %v but it returned %v", toTest, expected, isLowercaseAscii(toTest))
		}
	}
}

func TestFrequencyCountsInString(test *testing.T) {
	tests := map[string]map[byte]int{
		"D'M D'LL": map[byte]int{'D': 2, 'M': 1, 'L': 2, byte(39): 0, ' ': 0},
	}

	for testString, expectedResults := range tests {
		freqs := frequencyCountInString(testString)

		for curByte, expectedCount := range expectedResults {
			actualCount, found := freqs[curByte]
			if found && (actualCount != expectedCount) {
				test.Errorf("Expected %v for %c but got %v", expectedCount, curByte, actualCount)
			}

			if !found && expectedCount != 0 {
				test.Errorf("For %c expected count of %v but got count of %v", curByte, expectedCount, actualCount)
			}
		}
	}
}

func TestCountTotalCharacters(test *testing.T) {
	tests := map[string]int{
		"D'M D'LL": 5,
		"%$a'":     0,
	}

	for curString, expectedCount := range tests {
		actualCount := countTotalCharacters(curString)
		if expectedCount != actualCount {
			test.Errorf("Expected %v for total characters in %s but got %v", expectedCount, curString, actualCount)
		}
	}
}
