package cmd

import (
	"bufio"
	"strings"
	"testing"
	"time"
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
	dictChannel := make(chan string)
	go func() {
		feedDictionaryReaders(dictChannel, bufio.NewReader(strings.NewReader(dictionary)))
	}()

	findMatchesFromDictionary(matchesData, dictChannel)

	if len(matchesData[0].patternMatches) != 2 {
		test.Errorf("Expected HELLO to have 2 matches, but it had %v", len(matchesData[0].patternMatches))
	}

	if len(matchesData[1].patternMatches) != 1 {
		test.Errorf("Expected CHEESES to have 1 match, but it had %v", len(matchesData[1].patternMatches))
	}
}

type partitionTest struct {
	count                 int
	expectedMatchesAtZero []string
}

func stringInSlice(seek string, stringSlice []string) bool {
	for _, curString := range stringSlice {
		if seek == curString {
			return true
		}
	}
	return false
}

func TestPartitionMatches(test *testing.T) {
	data := &substitutionWordMatches{"HELLO", "ABCCD", []string{"YUCCA", "WATTS", "VROOM", "VILLA", "SWEET"}}

	tests := []partitionTest{
		partitionTest{5, []string{"YUCCA"}},
		partitionTest{1, []string{"YUCCA", "WATTS", "VROOM", "VILLA", "SWEET"}},
		partitionTest{2, []string{"YUCCA", "VROOM", "SWEET"}},
	}

	for index, curTest := range tests {
		partitioned := partitionMatches(curTest.count, data)

		if len(partitioned) != curTest.count {
			test.Errorf("Test case %v: expected %v partitions but got %v", index, curTest.count, len(partitioned))
		}

		// verify matches are what's expected
		if len(partitioned[0].patternMatches) != len(curTest.expectedMatchesAtZero) {
			test.Errorf("Test case %d: Expected partition 0 to have %d items but had %d",
				index, len(curTest.expectedMatchesAtZero), len(partitioned[0].patternMatches))
		}

		for _, expectToFind := range curTest.expectedMatchesAtZero {
			if !stringInSlice(expectToFind, partitioned[0].patternMatches) {
				test.Errorf("Test case %d: Expected to find %s in %v but did not",
					index, expectToFind, partitioned[0].patternMatches)
			}
		}

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

	resultsChannel := make(chan map[byte]byte, 1)
	dictionary := "WILLING\nPEOPLE\nSOME"
	dictChannel := make(chan string)
	go func() {
		feedDictionaryReaders(dictChannel, bufio.NewReader(strings.NewReader(dictionary)))
	}()
	findMatchesFromDictionary(matchesData, dictChannel)
	collectValidMaps(matchesData, make(map[byte]byte), resultsChannel)
	byteMap := <-resultsChannel
	close(resultsChannel)

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

	// test length from different dictionaries

	testCases := map[string]int{
		"BOSOMY\nHELLFIRE\nCRUTCH":                     0,
		"WILLING\nPEOPLE\nSOME\nSUCCUMB\nTHATCH\nGASH": 2,
	}

	for testDict, expectedLength := range testCases {
		for _, match := range matchesData {
			match.patternMatches = make([]string, 0, 2)
		}
		resultsChannel := make(chan map[byte]byte)
		dictChannel := make(chan string)
		go func() {
			feedDictionaryReaders(dictChannel, bufio.NewReader(strings.NewReader(testDict)))
		}()
		findMatchesFromDictionary(matchesData, dictChannel)
		go collectValidMaps(matchesData, make(map[byte]byte), resultsChannel)

		byteMaps := make([]map[byte]byte, 0, expectedLength)
	Loop:
		for {
			select {
			case validMap := <-resultsChannel:
				if len(validMap) > 0 {
					byteMaps = append(byteMaps, validMap)
				}
			case <-time.After(2 * time.Second):
				break Loop
			}
		}

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
