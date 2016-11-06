package logger

import (
	"log"
	"io"
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

type Logger struct {
	log *log.Logger
	prefix string
}

var (
	err = "ERROR"
	info = "INFO"
	debug = "DEBUG"
	testing = "TESTING"
	lightBlue = "\x1b[36m"
	yellow = "\x1b[33m"
	green = "\x1b[32m"
	red = "\x1b[31m"
	reset = "\x1b[0m"
)

func (l *Logger) formatString(level string, str string, color string) string {
	return fmt.Sprintln(color + level + "[" + l.prefix + "]" + ": " + reset + str)
}

func (l *Logger) Init(output io.Writer, preFix string, flag int) {
	l.log = log.New(output, "", flag)
	l.prefix = preFix
}

func (l *Logger) Error(msg string) {
	_, file, line, _ := runtime.Caller(1)
	tmp := strings.Split(file, "/")
	f := tmp[len(tmp)- 1]

	errMsg := l.formatString(err, msg, red)
	str := errMsg + red + f + ":" + strconv.Itoa(line) + reset

	l.log.Println(str)
}

func (l *Logger) Info(msg string) {
	l.log.Println(l.formatString(info, msg, green))
}

func (l *Logger) Debug(msg string) {
	_, file, line, _ := runtime.Caller(1)
	tmp := strings.Split(file, "/")
	f := tmp[len(tmp)- 1]

	errMsg := l.formatString(err, msg, yellow)
	str := errMsg + yellow + f + ":" + strconv.Itoa(line) + reset

	l.log.Println(str)
}

func (l *Logger) Testing(msg string) {
	l.log.Println(l.formatString(testing, msg, lightBlue))
}


