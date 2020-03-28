package cmd

import (
	"testing"
)

type shiftTest struct {
	start       byte
	shiftAmount int
	expected    byte
}

func TestShiftByte(test *testing.T) {
	tests := []shiftTest{
		shiftTest{'A', 1, 'B'},
		shiftTest{'Z', 2, 'B'},
		shiftTest{' ', 10, ' '},
		shiftTest{'y', 5, 'd'},
	}

	for _, curTest := range tests {
		shiftedByte := shiftByte(curTest.start, curTest.shiftAmount)
		if shiftedByte != curTest.expected {
			test.Errorf("Expected %c from shiftByte(%c, %d) but got %c",
				curTest.expected, curTest.start, curTest.shiftAmount, shiftedByte)
		}
	}
}
