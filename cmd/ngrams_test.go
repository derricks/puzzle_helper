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
		scanner := NewNgramScanner(reader, testCase.ngramSize)

		for scanner.Scan() {
			actuals = append(actuals, scanner.Text())
		}

		if len(actuals) != len(testCase.expectedTokens) {
			test.Logf("Actuals: %v", actuals)
			test.Errorf("Test case %d: Expected %d tokens but got %d", index, len(testCase.expectedTokens), len(actuals))
		}

	}
}
