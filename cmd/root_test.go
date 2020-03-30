package cmd

import (
	"bufio"
	"regexp"
	"strings"
	"testing"
)

func TestFeedDictionaryReaders(test *testing.T) {
	dictionaries := []*bufio.Reader{
		bufio.NewReader(strings.NewReader("STRINGONE\nSTRINGTWO\nSTRINGTHREE")),
		bufio.NewReader(strings.NewReader("lcstringone\nlcstringtwo\nlcstringthree")),
	}

	isuppercase := regexp.MustCompile("^[A-Z]+$")

	entryChannel := make(chan string)
	go func() {
		feedDictionaryReaders(entryChannel, dictionaries...)
	}()

	entries := make([]string, 0)
	for entry := range entryChannel {
		if !isuppercase.MatchString(entry) {
			test.Errorf("String %v should have been in uppercase.", entry)
		}
		entries = append(entries, entry)
	}
	if len(entries) != 6 {
		test.Errorf("Should have received %d entries but received %d", 6, len(entries))
	}

}
