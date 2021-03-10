package glog

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Config struct {
	Format          string
	DateFMT         string
	Level           int
	Handlers        []io.Writer // general Handlers for NOTSET, DEBUG and INFO. Default for WARN, ERROR and CRITICAL if warning and error Handlers are not given.
	WarningHandlers []io.Writer
	ErrorHandlers   []io.Writer // handles error and critical writers
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
	Format:          "%(t)s | %(n)20s | %(f)30s() | %(l)8s | %(m)s",
	DateFMT:         "%b %d %H:%M:%S",
	Level:           NOTSET,
	Handlers:        []io.Writer{os.Stdout},
	WarningHandlers: []io.Writer{},
	ErrorHandlers:   []io.Writer{},
}

var logs map[string]*Log
var lock = &sync.Mutex{}

// Format regex
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

func myCallerFuncWithSkip(skip int) string {
	return getFrame(2 + skip).Function
}

// myCallerFile returns the file name of the caller that called it
func myCallerFile() string {
	return getFrame(2).File
}

func Configure(cnf Config) { config = cnf }

// instantiates a singleton of logs slice if necessary
func instantiate() {
	lock.Lock()
	defer lock.Unlock()

	if logs == nil {
		// instantiate singleton logs
		logs = make(map[string]*Log)
	}
}

// returns a logger object to the caller
func GetLog() *Log {
	instantiate() // makes sure singleton is instantiated if necessary

	// remove leading file path and trailing file extension to retrieve the module name only.
	name := myCallerFile()
	parts := strings.Split(name, "/")
	name = strings.Split(parts[len(parts)-1], ".")[0]

	// check if logger is already defined, if so return it
	l, ok := logs[name]
	if ok {
		return l
	}

	l = &Log{
		name,
		false,
		&config,
	}

	logs[l.Name] = l // store reference to log in logs map
	return l
}

// attempts to retrieve an active logger and return it
// returns nil if no log was found
func FetchLog(name string) *Log {
	instantiate() // makes sure singleton is instantiated if necessary
	l, ok := logs[name]
	if ok {
		return l
	}

	return nil
}

func writeToHandlers(msg string, handlers []io.Writer) error {
	// logs data to Handlers
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

		Format codes:
		%(f)Xs               | function name                     e.g.: %(func)20s
		%(n)Xs               | log name                          e.g.: %(name)20s
		%(l)Xs               | Level name                        e.g.: %(Level)20s
		%(t)Xs               | time (formatted using dateFmt)    e.g.: %(time)20s
		%(m)Xs               | message                           e.g.: %(msg)20s
	*/
	var data []interface{}

	// create data array with the data corresponding to the data codes in the correct order
	chars := len(config.Format)
	for i := 0; i <= chars-4; i++ {
		set := config.Format[i : i+4]
		switch set {
		case "%(f)":
			data = append(data, myCallerFuncWithSkip(2)) // skipping 2 frames to get the actual caller function
			break

		case "%(n)":
			data = append(data, l.Name)
			break

		case "%(l)":
			data = append(data, getLevelStr(level))
			break

		case "%(t)":
			data = append(data, getDateTime(time, config.DateFMT))
			break

		case "%(m)":
			data = append(data, msg)
			break
		}
	}

	// remove data codes and leave formatting codes for sprintf
	f := fmtRe.ReplaceAllString(config.Format, "%")

	// apply formatting and return the formatted string
	return fmt.Sprintf(f, data...)
}

func log(l *Log, level int, msg string) {
	// check if log is silenced
	if l.Silenced {
		return
	}

	// compare config Level against the requested log Level
	if !(level >= config.Level) {
		return
	}

	// append newline
	if msg[len(msg)-1] != '\n' {
		msg += "\n"
	}

	// Format msg
	msg = formatMsg(l, time.Now(), msg, level)

	handlers := config.Handlers
	if level == WARNING && len(config.WarningHandlers) > 0 {
		handlers = config.WarningHandlers
	} else if level > WARNING && len(config.ErrorHandlers) > 0 {
		handlers = config.ErrorHandlers
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

func (l *Log) Debugf(format string, args ...interface{}) {
	log(l, DEBUG, fmt.Sprintf(format, args...))
}

func (l *Log) Infof(format string, args ...interface{}) {
	log(l, INFO, fmt.Sprintf(format, args...))
}

func (l *Log) Warnf(format string, args ...interface{}) {
	log(l, WARNING, fmt.Sprintf(format, args...))
}

func (l *Log) Errorf(format string, args ...interface{}) {
	log(l, ERROR, fmt.Sprintf(format, args...))
}

func (l *Log) Criticalf(format string, args ...interface{}) {
	log(l, CRITICAL, fmt.Sprintf(format, args...))
}
