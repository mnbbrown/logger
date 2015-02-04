package logger

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"os"
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
	reset   = string([]byte{27, 91, 48, 109})
)

type LoggerSink interface {
	io.Writer
}

// This is a list of logger sinks (io.Writer).
// When Write is called on a Logger it loops through the list writing to the individual sinks.
type Logger struct {
	sinks  []LoggerSink
	prefix string
}

func NewLogger() *Logger {
	return &Logger{}
}

func (l *Logger) AddSink(sink LoggerSink) {
	l.sinks = append(l.sinks, sink)
}

func (l *Logger) AddLogentriesSink(token, url string, port int) {
	les := &LogentriesSink{
		token: token,
		url:   url,
		port:  port,
	}
	l.AddSink(les)
}

func (l *Logger) AddLocalSink() {
	l.AddSink(os.Stdout)
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
	for _, w := range l.sinks {
		n, err = w.Write(p)
	}
	return n, err
}

// https://github.com/gin-gonic/gin/blob/develop/logger.go
func colorForStatus(code int) string {
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

func colorForMethod(method string) string {
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
		return reset
	}
}

type FormatBuffer struct {
	buffer bytes.Buffer
	items  interface{}
}

func (f *FormatBuffer) AddItems(format string, a ...interface{}) {
	f.buffer.WriteString(format)
	f.items = append(f.items, a...)
}

func (f *FormatBuffer) String() string {
	return fmt.Printf(f.buffer.String(), f.items)
}

func (l *Logger) LoggerMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		buffer := &FormatBuffer{}

		// If request has been assigned a unique ID print it.
		if id := ctx.Request.Header.Get("X-Request-Id"); id != "" {
			buffer.AddItem("| %s ", id)
		}

		start := time.Now()

		// Process request
		ctx.Next()

		end := time.Now()
		latency := end.Sub(start)

		method := ctx.Request.Method
		statusCode := ctx.Writer.Status()
		statusColor := colorForStatus(statusCode)
		methodColor := colorForMethod(method)

		buffer.AddItems("| %s %3d %s | %5v | %s | %s %s %s %s\n%s",
			statusColor, statusCode, reset,
			latency,
			ClientIP(ctx),
			methodColor, method, reset,
			ctx.Request.URL.Path,
			ctx.Errors.String(),
		)
		l.Print(buffer.String())
	}
}

func ClientIP(ctx *gin.Context) string {
	clientIP := ctx.Request.Header.Get("X-Real-IP")
	if len(clientIP) == 0 {
		clientIP = ctx.Request.Header.Get("X-Forwarded-For")
	}
	if len(clientIP) == 0 {
		clientIP = ctx.Request.RemoteAddr
	}
	return clientIP
}
