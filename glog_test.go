package glog

import (
	"bytes"
	"io"
	"os"
	"testing"
	"time"
)

// need to emulate these two calls to get correct caller function when calling formatMsg directly
type formatCall func(*Log, time.Time, string, int) string

func emulateFormatCall(fn formatCall, lptr *Log, t time.Time, msg string, level int) string {
	return fn(lptr, t, msg, level)
}

func emulateLogCall(fn formatCall, lptr *Log, t time.Time, msg string, level int) string {
	return emulateFormatCall(fn, lptr, t, msg, level)
}

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
	expected := "Nov 17 20:34:58 |            glog_test | github.com/HOWZ1T/glog.TestFormatMsg() |    DEBUG | " +
		"This is a test"
	got := emulateLogCall(formatMsg, l,
		time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC),
		"This is a test", DEBUG)

	if got != expected {
		t.Errorf("Formatting was incorrect:\ngot:\n%s\n\nexpected:\n%s\n", got, expected)
	}
}

func TestLogging(t *testing.T) {
	buf := bytes.NewBufferString("")

	// data setup
	l := GetLog()
	fmt := "%(n)20s | %(f)30s() | %(l)8s | %(m)s"
	datefmt := "%b %d %H:%M:%S"

	cnf := Config{
		fmt,
		datefmt,
		DEBUG,
		[]io.Writer{buf},
		[]io.Writer{},
		[]io.Writer{},
	}

	Configure(cnf)
	FetchLog("")
	l.Debug("Debug Test")
	l.Info("Info Test")
	l.Warn("Warning Test")
	l.Error("Error Test")
	l.Critical("Critical Test")

	expected :=
		`           glog_test | github.com/HOWZ1T/glog.TestLogging() |    DEBUG | Debug Test
           glog_test | github.com/HOWZ1T/glog.TestLogging() |     INFO | Info Test
           glog_test | github.com/HOWZ1T/glog.TestLogging() |     WARN | Warning Test
           glog_test | github.com/HOWZ1T/glog.TestLogging() |    ERROR | Error Test
           glog_test | github.com/HOWZ1T/glog.TestLogging() | CRITICAL | Critical Test` + "\n"

	got := buf.String()

	if got != expected {
		t.Errorf("Result was incorrect:\ngot:\n%s\n\nexpected:\n%s\n", got, expected)
	}

	buf.Reset()
	l.Debugf("Debug %s", "Test")
	l.Infof("Info %s", "Test")
	l.Warnf("Warning %s", "Test")
	l.Errorf("Error %s", "Test")
	l.Criticalf("Critical %s", "Test")
	got = buf.String()

	if got != expected {
		t.Errorf("Result was incorrect:\ngot:\n%s\n\nexpected:\n%s\n", got, expected)
	}
}
