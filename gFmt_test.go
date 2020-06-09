package glog

import (
	"testing"
	"time"
)

func TestGetNumDigits(t *testing.T) {
	c1 := getNumDigits(8020)
	c2 := getNumDigits(80219)

	if c1 != 4 {
		t.Errorf("for number: 8020, expected: %d, got: %d", 4, c1)
	}

	if c2 != 5 {
		t.Errorf("for number: 80219, expected: %d, got: %d", 5, c2)
	}
}

func TestGetDateTime(t *testing.T) {
	// November 17th, 2009 :: 20:34:58::651387237 UTC
	then := time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)

	fmt := "Week(Short): %a, Week(Long): %A, Week(Num): %w\n" +
		"Day of month: %d, Month(Short): %b, Month(Long): %B, Month(Num): %m\n" +
		"Year(Short): %y, Year(Long): %Y\n" +
		"Hour(24h): %H, Hour(12h): %I, AM/PM: %p\n" +
		"Minute: %M, Second: %S, Nanosecond: %f\n" +
		"UTC Offset: %z, Timezone: %Z\n" +
		"Day of the year: %j, Week of the year (Monday): %W\n"

	expected := "Week(Short): Tue, Week(Long): Tuesday, Week(Num): 2\n" +
		"Day of month: 17, Month(Short): Nov, Month(Long): November, Month(Num): 11\n" +
		"Year(Short): 09, Year(Long): 2009\n" +
		"Hour(24h): 20, Hour(12h): 08, AM/PM: PM\n" +
		"Minute: 34, Second: 58, Nanosecond: 651387237\n" +
		"UTC Offset: +0000, Timezone: UTC\n" +
		"Day of the year: 321, Week of the year (Monday): 47\n"

	got := getDateTime(then, fmt)
	if got != expected {
		t.Errorf("Date Time was incorrect.\nGot:\n%s\n\nExpected:\n%s\n", got, expected)
	}
}
