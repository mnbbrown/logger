package logger

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

var AlreadyConnected = errors.New("Logger already connected")

type LogentriesSink struct {
	conn      net.Conn
	connError error
	lock      sync.Mutex
	prefix    string
	token     string
	url       string
	port      int
}

func (l *LogentriesSink) SetToken(token string) {
	l.token = token
}

func (l *LogentriesSink) Open() error {
	if l.token == "" {
		return errors.New("Logger token not defined. Use NewLogger(token).")
	}

	if l.url == "" {
		return errors.New("Logger url is not defined.")
	}

	if l.port == 0 {
		return errors.New("Logger port is not defined.")
	}

	l.lock.Lock()

	if l.conn != nil {
		l.lock.Unlock()
		return AlreadyConnected
	}

	defer l.lock.Unlock()
	l.conn, l.connError = net.DialTimeout("tcp", fmt.Sprintf("%s:%d", l.url, l.port), 2*time.Second)
	if l.connError != nil {
		return l.connError
	}

	return nil
}

func (l *LogentriesSink) IsConnected() bool {
	l.lock.Lock()
	defer l.lock.Unlock()

	if l.conn == nil {
		return false
	}

	buf := make([]byte, 1)

	l.conn.SetReadDeadline(time.Now())

	_, err := l.conn.Read(buf)

	switch err.(type) {
	case net.Error:
		if err.(net.Error).Timeout() == true {
			l.conn.SetReadDeadline(time.Time{})
			return true
		}
	}

	return false

}

func (l *LogentriesSink) EnsureOpenConnection() error {
	if ok := l.IsConnected(); !ok {
		if err := l.Open(); err != nil {
			return err
		}
	}
	return nil
}

func (l *LogentriesSink) Write(p []byte) (n int, err error) {
	if err := l.EnsureOpenConnection(); err != nil {
		return 0, err
	}

	l.lock.Lock()
	defer l.lock.Unlock()

	count := strings.Count(string(p), "\n")
	p = []byte(strings.Replace(string(p), "\n", "\u2028", count-1))

	var buf []byte
	buf = append(buf, (l.token + " ")...)
	buf = append(buf, ("[" + l.prefix + "] ")...)
	buf = append(buf, p...)

	return l.conn.Write(buf)
}
