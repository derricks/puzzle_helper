package cmd

import (
	"bufio"
	"strings"
	"testing"
)

func TestSubstitutionPattern(test *testing.T) {
	tests := map[string]string{
		"HELLO": "ABCCD",
		"A":     "A",
	}

	for input, expected := range tests {
		actual := substitutionPattern(input)
		if expected != actual {
			test.Errorf("Expected %v from input text %v but got %v", expected, input, actual)
		}
	}
}

func TestFindMatchesFromDictionary(test *testing.T) {
	matchesData := []*substitutionWordMatches{
		&substitutionWordMatches{"HELLO", "ABCCD", make([]string, 0, 2)},
		&substitutionWordMatches{"CHEESES", "ABCCDCD", make([]string, 0, 2)},
	}

	dictionary := "BLEED\nWHEESES\nBOLT\nBOSSY"
	findMatchesFromDictionary(matchesData, bufio.NewReader(strings.NewReader(dictionary)))

	if len(matchesData[0].patternMatches) != 2 {
		test.Errorf("Expected HELLO to have 2 matches, but it had %v", len(matchesData[0].patternMatches))
	}

	if len(matchesData[1].patternMatches) != 1 {
		test.Errorf("Expected CHEESES to have 1 match, but it had %v", len(matchesData[1].patternMatches))
	}
}

func TestCopyByteMap(test *testing.T) {
	source := map[byte]byte{
		'A': 'B',
		'C': 'D',
	}

	copy := copyByteMap(source)

	for key, value := range source {
		copyValue, found := copy[key]
		if !found {
			test.Errorf("Did not find %c in copy, but should have", key)
		}
		if copyValue != value {
			test.Errorf("Values did not match in copy. Expected %c but got %c", value, copyValue)
		}
	}
}

func TestCollectValidMaps(test *testing.T) {
	matchesData := []*substitutionWordMatches{
		// willing people some
		&substitutionWordMatches{"BUXXUDR", "ABCCBDE", make([]string, 0, 2)},
		&substitutionWordMatches{"CPICXP", "ABCADB", make([]string, 0, 2)},
		&substitutionWordMatches{"TIZP", "ABCD", make([]string, 0, 2)},
	}

	dictionary := "WILLING\nPEOPLE\nSOME"
	findMatchesFromDictionary(matchesData, bufio.NewReader(strings.NewReader(dictionary)))
	byteMap := collectValidMaps(matchesData, make(map[byte]byte))[0]

	expectedMap := map[byte]byte{
		'B': 'W',
		'U': 'I',
		'X': 'L',
		'D': 'N',
		'R': 'G',
		'C': 'P',
		'P': 'E',
		'I': 'O',
		'T': 'S',
		'Z': 'M',
	}

	for cipher, plain := range expectedMap {
		if plain != byteMap[cipher] {
			test.Errorf("Expected %c to map to %c but it maps to %c", cipher, plain, byteMap[cipher])
		}
	}

	testCases := map[string]int{
		"BOSOMY\nHELLFIRE\nCRUTCH":                     0,
		"WILLING\nPEOPLE\nSOME\nSUCCUMB\nTHATCH\nGASH": 2,
	}

	for testDict, expectedLength := range testCases {
		for _, match := range matchesData {
			match.patternMatches = make([]string, 0, 2)
		}

		findMatchesFromDictionary(matchesData, bufio.NewReader(strings.NewReader(testDict)))
		byteMaps := collectValidMaps(matchesData, make(map[byte]byte))
		if len(byteMaps) != expectedLength {
			test.Errorf("Expected %d item map from %s, but it was %d items", expectedLength, testDict, len(byteMaps))
			for _, curMap := range byteMaps {
				for cipherByte, plainByte := range curMap {
					test.Logf("Byte %c maps to %c", cipherByte, plainByte)
				}
				test.Logf("\n")
			}
		}
	}
}
