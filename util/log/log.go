// Copyright 2015 Eleme Inc. All rights reserved.

// Package log implements leveled logging.
package log

import (
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"time"
)

// Level
const (
	DEBUG int = iota
	INFO
	WARN
	ERROR
)

// Level name
var levelNames = [4]string{"DEBUG", "INFO", "WARN", "ERROR"}

// Logging runtime
var (
	name    string
	level             = INFO
	w       io.Writer = os.Stderr
	colored           = true
	enabled           = true
)

// colors to ansi code map
var colors = map[string]int{
	"black":   0,
	"red":     1,
	"green":   2,
	"yellow":  3,
	"blue":    4,
	"magenta": 5,
	"cyan":    6,
	"white":   7,
}

// levelColors
var levelColors = map[int]string{
	DEBUG: "blue",
	INFO:  "green",
	WARN:  "yellow",
	ERROR: "red",
}

// SetColored sets the color enability.
func SetColored(b bool) {
	colored = b
}

// SetName sets the logging name.
func SetName(s string) {
	name = s
}

// SetLevel sets the logging level.
func SetLevel(l int) {
	level = l % len(levelNames)
}

// SetWriter sets the writer.
func SetWriter(writer io.Writer) {
	w = writer
}

// Disable the logging.
func Disable() {
	enabled = false
}

// Enable the logging.
func Enable() {
	enabled = true
}

// Debug logs message with level DEBUG.
func Debug(a ...interface{}) error {
	return log(DEBUG, fmt.Sprint(a...))
}

// Info logs message with level INFO.
func Info(a ...interface{}) error {
	return log(INFO, fmt.Sprint(a...))
}

// Warn logs message with level WARN.
func Warn(a ...interface{}) error {
	return log(WARN, fmt.Sprint(a...))
}

// Error logs message with level ERROR.
func Error(a ...interface{}) error {
	return log(ERROR, fmt.Sprint(a...))
}

// Fatal and logs message with level FATAL.
func Fatal(a ...interface{}) {
	log(ERROR, fmt.Sprint(a...))
	os.Exit(1)
}

// Debugf formats and logs message with level DEBUG.
func Debugf(format string, a ...interface{}) error {
	return log(DEBUG, fmt.Sprintf(format, a...))
}

// Infof formats and logs message with level INFO.
func Infof(format string, a ...interface{}) error {
	return log(INFO, fmt.Sprintf(format, a...))
}

// Warnf formats and logs message with level WARN.
func Warnf(format string, a ...interface{}) error {
	return log(WARN, fmt.Sprintf(format, a...))
}

// Errorf formats and logs message with level ERROR.
func Errorf(format string, a ...interface{}) error {
	return log(ERROR, fmt.Sprintf(format, a...))
}

// Fatalf formats and logs message with level FATAL.
func Fatalf(format string, a ...interface{}) {
	log(ERROR, fmt.Sprintf(format, a...))
	os.Exit(1)
}

// Colored returns text in color.
func Colored(color string, text string) string {
	return fmt.Sprintf("\033[3%dm%s\033[0m", colors[color], text)
}

// log dose logging.
func log(l int, msg string) error {
	if enabled && l >= level {
		// Caller pkg.
		_, fileName, line, _ := runtime.Caller(2)
		pkgName := path.Base(path.Dir(fileName))
		filepath := path.Join(pkgName, path.Base(fileName))
		// Datetime and pid.
		now := time.Now().String()[:19]
		// Message
		level := levelNames[l]
		header := Colored(levelColors[l], fmt.Sprintf("%s %s %s:%d", level, now, filepath, line))
		_, err := fmt.Fprintf(w, "%s %s\n", header, msg)
		return err
	}
	return nil
}
