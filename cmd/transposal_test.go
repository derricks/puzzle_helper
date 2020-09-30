package cmd

import (
	"testing"
)

func TestCreateLetterCountsMap(test *testing.T) {
	input := "THIS IS A STRING WITH REPEATED LETTERS"
	expectedCounts := map[string]int{
		"T": 6,
		"H": 2,
		"I": 4,
		"S": 4,
		"A": 2,
		"R": 3,
		"N": 1,
		"G": 1,
		"W": 1,
		"E": 5,
		"P": 1,
		"D": 1,
		"L": 1,
	}

	actual := createLetterCountsMap(input)

	// verify that every letter in expected is in actual with the correct count
	for testLetter, expectedCount := range expectedCounts {
		if _, present := actual[testLetter]; !present {
			test.Errorf("Letter %s is not in response", testLetter)
		}

		if actual[testLetter] != expectedCount {
			test.Errorf("Letter %s has incorrect count. Expected %d but got %d", testLetter, expectedCount, actual[testLetter])
		}
	}

	// verify that there are no spaces in actual
	if _, present := actual[" "]; present {
		test.Errorf("Found space in actual count but there should be none")
	}

	// verify that every letter in actual is also in expected
	for actualLetter, _ := range actual {
		if _, present := expectedCounts[actualLetter]; !present {
			test.Errorf("Actual contains %s which is not expected", actualLetter)
		}
	}
}

func TestDecrementLetterCounts(test *testing.T) {
	input := map[string]int{
		"T": 6,
		"N": 1,
	}

	newMap := decrementLetterCounts("T", input)
	tCount := newMap["T"]
	if tCount != 5 {
		test.Errorf("T should have been 5 was %d", tCount)
	}

	nCount := newMap["N"]
	if nCount != 1 {
		test.Errorf("N should have been unchanged after the decrement. Expected 1 got %d", nCount)
	}

	newMap = decrementLetterCounts("N", input)
	if _, nPresent := newMap["N"]; nPresent {
		test.Errorf("N should not have been present in map but is")
	}

}
