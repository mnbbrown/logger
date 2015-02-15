package logger

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// https://github.com/gin-gonic/gin/blob/develop/logger.go
var (
	green   = string([]byte{27, 91, 57, 55, 59, 52, 50, 109})
	white   = string([]byte{27, 91, 57, 48, 59, 52, 55, 109})
	yellow  = string([]byte{27, 91, 57, 55, 59, 52, 51, 109})
	red     = string([]byte{27, 91, 57, 55, 59, 52, 49, 109})
	blue    = string([]byte{27, 91, 57, 55, 59, 52, 52, 109})
	magenta = string([]byte{27, 91, 57, 55, 59, 52, 53, 109})
	cyan    = string([]byte{27, 91, 57, 55, 59, 52, 54, 109})
	Reset   = string([]byte("\033[0m"))
)

type LoggerSink interface {
	io.Writer
}

// This is a list of logger sinks (io.Writer).
// When Write is called on a Logger it loops through the list writing to the individual sinks.
type Logger struct {
	sinks  []LoggerSink
	prefix func() string
	tags   []string
	mu     sync.Mutex
}

func NewLogger() *Logger {
	return &Logger{
		prefix: PrefixGen,
	}
}

func (l *Logger) AddSink(sink LoggerSink) {
	l.mu.Lock()
	l.sinks = append(l.sinks, sink)
	l.mu.Unlock()
}

func (l *Logger) SetPrefix(prefix func() string) {
	l.prefix = prefix
}

func (l *Logger) SetTags(tags ...string) {
	l.tags = tags
}

func (l *Logger) AddLogentriesSink(token, url string, port int) error {
	les := &LogentriesSink{
		token: token,
		url:   url,
		port:  port,
	}
	if err := les.Open(); err != nil {
		return err
	}
	l.AddSink(les)
	return nil
}

func (l *Logger) AddLocalSink() {
	l.AddSink(os.Stdout)
}

func PrefixGen() string {
	return time.Now().Format(time.RFC3339)
}

func NilGen() string {
	return ""
}

func (l *Logger) NewRequestLogger(tags ...string) *Logger {
	rl := NewLogger()
	rl.AddSink(l)
	rl.SetTags(tags...)
	return rl
}

func (l *Logger) Printf(format string, v ...interface{}) {
	go l.Write([]byte(fmt.Sprintf(format, v...)))
}

func (l *Logger) Print(str string) {
	go l.Write([]byte(str))
}

func (l *Logger) Fatalln(v ...interface{}) {
	l.Write([]byte(fmt.Sprintln(v...)))
	os.Exit(1)
}

func (l *Logger) Println(v ...interface{}) {
	go l.Write([]byte(fmt.Sprintln(v...)))
}

func (l *Logger) Write(p []byte) (n int, err error) {
	var buf []byte
	buf = append(buf, []byte(l.prefix()+" ")...)
	buf = append(buf, p...)
	for _, w := range l.sinks {
		n, err = w.Write(buf)
	}
	return n, err
}

// https://github.com/gin-gonic/gin/blob/develop/logger.go
func StatusColour(code int) string {
	switch {
	case code >= 200 && code <= 299:
		return green
	case code >= 300 && code <= 399:
		return white
	case code >= 400 && code <= 499:
		return yellow
	default:
		return red
	}
}

func MethodColour(method string) string {
	switch {
	case method == "GET":
		return blue
	case method == "POST":
		return cyan
	case method == "PUT":
		return yellow
	case method == "DELETE":
		return red
	case method == "PATCH":
		return green
	case method == "HEAD":
		return magenta
	case method == "OPTIONS":
		return white
	default:
		return Reset
	}
}
