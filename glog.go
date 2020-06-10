package glog

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"
)

type Config struct {
	format          string
	datefmt         string
	level           int
	handlers        []io.Writer // general handlers for NOTSET to INFO and default for WARN, ERROR and CRITICAL
	warningHandlers []io.Writer
	errorHandlers   []io.Writer // handles error and critical writers
}

type Log struct {
	Name     string
	Silenced bool
	Config   *Config
}

// LOGGING LEVELS
const NOTSET = 0
const DEBUG = 10
const INFO = 20
const WARNING = 30
const ERROR = 40
const CRITICAL = 50

var config = Config{
	format:          "%(t)s | %(n)20s | %(f)30s() | %(l)8s | %(m)s",
	datefmt:         "%b %d %H:%M:%S",
	level:           NOTSET,
	handlers:        []io.Writer{os.Stdout},
	warningHandlers: []io.Writer{},
	errorHandlers:   []io.Writer{},
}

var logs map[string]*Log

// format regex
var fmtRe = regexp.MustCompile(`%\((t|n|f|l|m)\)`)

func getFrame(skipFrames int) runtime.Frame {
	// Source: https://stackoverflow.com/questions/35212985/is-it-possible-get-information-about-caller-function-in-golang
	// We need the frame at index skipFrames+2, since we never want runtime.Callers and getFrame
	targetFrameIndex := skipFrames + 2

	// Set size to targetFrameIndex+2 to ensure we have room for one more caller than we need
	programCounters := make([]uintptr, targetFrameIndex+2)
	n := runtime.Callers(0, programCounters)

	frame := runtime.Frame{Function: "unknown"}
	if n > 0 {
		frames := runtime.CallersFrames(programCounters[:n])
		for more, frameIndex := true, 0; more && frameIndex <= targetFrameIndex; frameIndex++ {
			var frameCandidate runtime.Frame
			frameCandidate, more = frames.Next()
			if frameIndex == targetFrameIndex {
				frame = frameCandidate
			}
		}
	}

	return frame
}

// myCallerFunc returns the function name of the caller that called it
func myCallerFunc() string {
	return getFrame(2).Function
}

// myCallerFile returns the file name of the caller that called it
func myCallerFile() string {
	return getFrame(2).File
}

func Configure(cnf Config) { config = cnf }

// returns a logger object to the caller
func GetLog() Log {
	// remove leading file path and trailing file extension to retrieve the module name only.
	name := myCallerFile()
	parts := strings.Split(name, "/")
	name = strings.Split(parts[len(parts)-1], ".")[0]
	l := Log{
		name,
		false,
		&config,
	}
	logs[l.Name] = &l // store reference to log in logs map
	return l
}

// attempts to retrieve an active logger and return it
// returns nil if no log was found
func FetchLog(name string) *Log {
	l, ok := logs[name]
	if ok {
		return l
	}

	return nil
}

func writeToHandlers(msg string, handlers []io.Writer) error {
	// logs data to handlers
	for _, handler := range handlers {
		_, err := handler.Write([]byte(msg))
		if err != nil {
			return err
		}
	}

	return nil
}

func getLevelStr(level int) string {
	switch level {
	case NOTSET:
		return "NOTSET"

	case DEBUG:
		return "DEBUG"

	case INFO:
		return "INFO"

	case WARNING:
		return "WARN"

	case ERROR:
		return "ERROR"

	case CRITICAL:
		return "CRITICAL"

	default:
		return "UNKNOWN"
	}
}

func formatMsg(l *Log, time time.Time, msg string, level int) string {
	/*
		key:
		X = number for padding

		format codes:
		%(f)Xs               | function name                     e.g.: %(func)20s
		%(n)Xs               | log name                          e.g.: %(name)20s
		%(l)Xs               | level name                        e.g.: %(level)20s
		%(t)Xs               | time (formatted using dateFmt)    e.g.: %(time)20s
		%(m)Xs               | message                           e.g.: %(msg)20s
	*/
	var data []interface{}

	// create data array with the data corresponding to the data codes in the correct order
	chars := len(config.format)
	for i := 0; i <= chars-4; i++ {
		set := config.format[i : i+4]
		switch set {
		case "%(f)":
			data = append(data, myCallerFunc())
			break

		case "%(n)":
			data = append(data, l.Name)
			break

		case "%(l)":
			data = append(data, getLevelStr(level))
			break

		case "%(t)":
			data = append(data, getDateTime(time, config.datefmt))
			break

		case "%(m)":
			data = append(data, msg)
			break
		}
	}

	// remove data codes and leave formatting codes for sprintf
	f := fmtRe.ReplaceAllString(config.format, "%")

	// apply formatting and return the formatted string
	return fmt.Sprintf(f, data...)
}

func log(l *Log, level int, msg string) {
	// check if log is silenced
	if l.Silenced {
		return
	}

	// compare config level against the requested log level
	if !(level >= config.level) {
		return
	}

	// append newline
	if msg[len(msg)-1] != '\n' {
		msg += "\n"
	}

	// format msg
	msg = formatMsg(l, time.Now(), msg, level) + "\n"

	handlers := config.handlers
	if level == 30 && len(config.warningHandlers) > 0 {
		handlers = config.warningHandlers
	} else if level > 30 && len(config.errorHandlers) > 0 {
		handlers = config.errorHandlers
	}

	err := writeToHandlers(msg, handlers)
	if err != nil {
		panic(err)
	}
}

// sets the silence state of the log
func (l *Log) Silence(b bool) {
	l.Silenced = b
}

func (l *Log) Debug(msg string) {
	log(l, DEBUG, msg)
}

func (l *Log) Info(msg string) {
	log(l, INFO, msg)
}

func (l *Log) Warn(msg string) {
	log(l, WARNING, msg)
}

func (l *Log) Error(msg string) {
	log(l, ERROR, msg)
}

func (l *Log) Critical(msg string) {
	log(l, CRITICAL, msg)
}
