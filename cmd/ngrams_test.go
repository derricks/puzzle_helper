package cmd

import (
	"strings"
	"testing"
)

// tests to write for scanner
// Hello, you!
// ...Hello
// ...He for trigrams (should be an error)
type ngramTest struct {
	input          string
	ngramSize      int
	expectedTokens []string
	errorExpected  bool
}

func TestNgramScanner(test *testing.T) {
	tests := []ngramTest{
		ngramTest{"Hello, you!", 4, []string{"HELL", "ELLO", "LLOY", "LOYO", "OYOU"}, false},
		ngramTest{"...Hello", 4, []string{"HELL", "ELLO"}, false},
		ngramTest{"...He", 3, nil, true},
		ngramTest{"he", 1, []string{"H", "E"}, false},
	}

	for index, testCase := range tests {
		reader := strings.NewReader(testCase.input)
		actuals := make([]string, 0, len(testCase.expectedTokens))
		scanner := NewNgramScanner(reader, testCase.ngramSize, false)

		for scanner.Scan() {
			actuals = append(actuals, scanner.Text())
		}

		if len(actuals) != len(testCase.expectedTokens) {
			test.Logf("Actuals: %v", actuals)
			test.Errorf("Test case %d: Expected %d tokens but got %d", index, len(testCase.expectedTokens), len(actuals))
		}

	}
}

func TestReadNgramsIntoTrie(test *testing.T) {
	input := "attack a Tacky Norse horse"
	expectedCounts := map[string]int{
		"ATTA": 1,
		"TTAC": 1,
		"TACK": 2,
		"ACKA": 1,
		"CKAT": 1,
		"KATA": 1,
		"ATAC": 1,
		"ACKY": 1,
		"CKYN": 1,
		"KYNO": 1,
		"YNOR": 1,
		"NORS": 1,
		"ORSE": 2,
		"RSEH": 1,
		"SEHO": 1,
		"EHOR": 1,
		"HORS": 1,
	}

	trie, totalCount := readNgramsIntoTrie(strings.NewReader(input), 4)

	if totalCount != 19 {
		test.Errorf("Expected 19 total count, got %d", totalCount)
	}

	for ngram, expectedCount := range expectedCounts {
		actualCount, wasPresent := trie.getValueForString(ngram)
		if !wasPresent {
			test.Errorf("Expected ngram %s in trie, but it was absent", ngram)
		}

		if actualCount != expectedCount {
			test.Errorf("Expected count of %d for %s but got %d", expectedCount, ngram, actualCount)
		}
	}

}
