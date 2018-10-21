package main

import (
	"testing"

	"github.com/marni/goigc"
)

func Test_lengthCalc(t *testing.T) {
	//Test cases to check the length to:
	t1 := "http://skypolaris.org/wp-content/uploads/IGS%20Files/Madrid%20to%20Jerez.igc"
	t2 := "https://raw.githubusercontent.com/marni/goigc/71f274dd21ef320ccbb6c6271c7df7fd8a908779/testdata/optimize-long-flight-1.igc"
	t3 := "https://raw.githubusercontent.com/marni/goigc/71f274dd21ef320ccbb6c6271c7df7fd8a908779/testdata/optimize-short-flight-1.igc"

	//Create map with test cases and expected return
	testCases := map[string]float64{
		t1: 443.2573603705269,
		t2: 968.516424687944,
		t3: 76.70910322623965,
	}

	//For each case, check to see if lengthCalc returns the expected result.
	for input, expected := range testCases {
		//First parse the url to extract the track
		track, err := igc.ParseLocation(input)
		if err != nil {
			t.Errorf("%s failed to parse: %s", input, err)
		}
		actual := lengthCalc(track)

		if actual != expected {
			t.Error("Failure: Expected ", expected, " got ", actual)
		}
	}
}
