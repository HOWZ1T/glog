package glog

import (
	"strconv"
	"strings"
	"time"
)

// returns the day of the week as a number 0-6 where 0 is Sunday
func getDayAsNum(day string) int {
	switch strings.ToLower(day) {
	case "monday":
		return 1

	case "tuesday":
		return 2

	case "wednesday":
		return 3

	case "thursday":
		return 4

	case "friday":
		return 5

	case "saturday":
		return 6

	case "sunday":
		return 0

	default:
		return -1
	}
}

// returns the month as a string number 01-12
func getMonthAsNumStr(month string) string {
	switch strings.ToLower(month) {
	case "january":
		return "01"

	case "february":
		return "02"

	case "march":
		return "03"

	case "april":
		return "04"

	case "may":
		return "05"

	case "june":
		return "06"

	case "july":
		return "07"

	case "august":
		return "08"

	case "september":
		return "09"

	case "october":
		return "10"

	case "november":
		return "11"

	case "december":
		return "12"

	default:
		return ""
	}
}

// returns the number of digits in a number. e.g.: an input of 123 will return 3
func getNumDigits(number int) int {
	count := 0
	for {
		count += 1
		number /= 10
		if number == 0 {
			return count
		}
	}
}

var dateCodes = map[rune]func(time time.Time) string{
	'a': func(time time.Time) string { return time.Weekday().String()[0:3] },                       // weekday, short (Wed)
	'A': func(time time.Time) string { return time.Weekday().String() },                            // weekday, full (Wednesday)
	'w': func(time time.Time) string { return strconv.Itoa(getDayAsNum(time.Weekday().String())) }, // weekday, as a number 0 - 6, 0 is Sunday  (3)
	'd': func(time time.Time) string { return strconv.Itoa(time.Day()) },                           // day of the month 01 - 31
	'b': func(time time.Time) string { return time.Month().String()[0:3] },                         // month name, short, Dec
	'B': func(time time.Time) string { return time.Month().String() },                              // month name, full, December
	'm': func(time time.Time) string { return getMonthAsNumStr(time.Month().String()) },            // month as number 01 - 12, 12
	'y': func(time time.Time) string { s := strconv.Itoa(time.Year()); return s[len(s)-2:] },       // year, short, without century, 18
	'Y': func(time time.Time) string { return strconv.Itoa(time.Year()) },                          // year, full, 2018

	'H': func(time time.Time) string { // hour 00 - 23, 17
		h := time.Hour()
		if h < 10 {
			return "0" + strconv.Itoa(h)
		}
		return strconv.Itoa(h)
	},

	'I': func(time time.Time) string { // hour 00 - 12, 05
		h := time.Hour()
		// convert 24 hour to 12 hour
		if h > 12 {
			h -= 12
		} else if h == 0 {
			h = 12
		}

		if h < 10 {
			return "0" + strconv.Itoa(h)
		}
		return strconv.Itoa(h)
	},

	'p': func(time time.Time) string { // AM/PM
		h := time.Hour()
		if h > 12 || h == 12 {
			return "PM"
		}
		return "AM"
	},

	'M': func(time time.Time) string { // minute, 00-59
		m := time.Minute()
		if m < 10 {
			return "0" + strconv.Itoa(m)
		}
		return strconv.Itoa(m)
	},

	'S': func(time time.Time) string { // second, 00-59
		s := time.Second()
		if s < 10 {
			return "0" + strconv.Itoa(s)
		}
		return strconv.Itoa(s)
	},

	'f': func(time time.Time) string { // nanosecond 000000-999999
		// guarantees a digit count (with padding) of at least 6, actual count may be greater than 6
		ns := time.Nanosecond()
		if ns == 0 {
			return "000000"
		}
		dc := getNumDigits(ns)

		if dc >= 6 {
			return strconv.Itoa(ns)
		}

		// pad zeros
		out := ""
		for i := 0; i < 6-dc; i++ {
			out += "0"
		}
		return out + strconv.Itoa(ns)
	},

	'z': func(time time.Time) string {
		// offset is in seconds east of UTC
		_, offset := time.Zone()

		// convert offset from seconds to hours and minutes
		// Format HHMM where H is hour and M is minute

		sign := "+"
		if offset < 0 {
			sign = "-"
			offset *= -1
		}

		h := offset / 3600
		remainder := offset - (h * 3600)
		mins := remainder / 60

		out := sign
		// pad hours and minutes and combine with sign to form utc offset in form: +0100
		if h < 10 {
			out += "0" + strconv.Itoa(h)
		} else {
			out += strconv.Itoa(h)
		}

		if mins < 10 {
			out += "0" + strconv.Itoa(mins)
		} else {
			out += strconv.Itoa(mins)
		}

		return out
	},

	'Z': func(time time.Time) string { // timezone CST
		name, _ := time.Zone()
		return name
	},

	'j': func(time time.Time) string { // day number of year 001-366
		d := time.YearDay()
		c := getNumDigits(d)
		if c == 3 {
			return strconv.Itoa(d)
		}

		// pad zeros
		out := ""
		for i := 0; i < 3-c; i++ {
			out += "0"
		}
		return out + strconv.Itoa(d)
	},

	'W': func(time time.Time) string { // week number of the year, Monday as the first day of the week, 00-53
		_, w := time.ISOWeek()
		if w < 10 {
			return "0" + strconv.Itoa(w)
		}
		return strconv.Itoa(w)
	},
}

func getDateTime(time time.Time, format string) string {
	out := ""
	chars := len(format)
	for i := 0; i < chars; i++ {
		// handles the last char
		if i == chars-1 {
			out += string(format[i])
			continue // skip over to the end
		}

		c1 := format[i]
		c2 := format[i+1]

		// test for date codes and apply the relevant function
		if c1 == '%' {
			fn, ok := dateCodes[rune(c2)]
			if ok {
				i += 1
				out += fn(time)
			} else {
				out += string(c1)
			}
		} else {
			out += string(c1)
		}
	}
	return out
}
