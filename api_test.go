package main

import (
	"testing"
	"time"
)

//Test uptime function
func Test_uptimeFunc(t *testing.T) {
	//timeStruct contains the two times to calculate the difference between
	type timeStruct struct {
		start time.Time
		now   time.Time
	}

	//Define and declare test cases
	var t1, t2, t3, t4, t5 timeStruct
	t1.start = time.Date(2015, 5, 1, 0, 0, 0, 0, time.UTC)
	t1.now = time.Date(2016, 6, 2, 1, 1, 1, 1, time.UTC)

	t2.start = time.Date(2016, 1, 2, 0, 0, 0, 0, time.UTC)
	t2.now = time.Date(2016, 2, 1, 0, 0, 0, 0, time.UTC)

	t3.start = time.Date(2016, 2, 2, 0, 0, 0, 0, time.UTC)
	t3.now = time.Date(2016, 3, 1, 0, 0, 0, 0, time.UTC)

	t4.start = time.Date(2015, 2, 11, 0, 0, 0, 0, time.UTC)
	t4.now = time.Date(2016, 1, 12, 0, 0, 0, 0, time.UTC)

	t5.start = time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC)
	t5.now = time.Date(2009, 11, 10, 23, 0, 0, 0, time.Local)

	//Insert test cases into a map with expected returns
	testCases := map[timeStruct]string{
		t1: "P1Y1M1DT1H1M1S",
		t2: "P30D",
		t3: "P28D",
		t4: "P11M1D",
		t5: "PT1H",
	}

	//For each case, check to see if uptimeFunc returns the expected result.
	for input, expected := range testCases {
		actual := uptimeFunc(input.start, input.now)

		if actual != expected {
			t.Error("Failure: Expected ", expected, " got ", actual)
		}
	}
}
