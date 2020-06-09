package glog

import (
	"io"
	"os"
	"testing"
	"time"
)

func TestGetLog(t *testing.T) {
	l := GetLog()

	target := "glog_test"
	if l.Name != target {
		t.Errorf("Log name was incorrect, got: %s, expected: %s", l.Name, target)
	}
}

func TestFormatMsg(t *testing.T) {
	// data setup
	l := GetLog()
	fmt := "%(t)s | %(n)20s | %(f)30s() | %(l)8s | %(m)s"
	datefmt := "%b %d %H:%M:%S"

	cnf := Config{
		fmt,
		datefmt,
		DEBUG,
		[]io.Writer{os.Stderr},
		[]io.Writer{},
		[]io.Writer{},
	}

	Configure(cnf)

	// test
	expected := "Nov 17 20:34:58 |            glog_test |             glog.TestFormatMsg() |    DEBUG | " +
		"This is a test"
	got := formatMsg(&l,
		time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC),
		"This is a test", DEBUG)

	if got != expected {
		t.Errorf("Log name was incorrect:\ngot:\n%s\n\nexpected:\n%s\n", got, expected)
	}
}

// TODO increase test coverage
