package logger

import (
	"fmt"
	"github.com/apudiu/server-backup/internal/util"
	"io"
	"os"
	"sync"
	"time"
)

type Logger struct {
	locker  sync.Mutex
	verbose bool
	data    []byte
}

// ToggleStdOut toggles printing the log to stdOut
func (l *Logger) ToggleStdOut(enable bool) {
	l.verbose = enable
}

// Add adds to the log
func (l *Logger) Add(b []byte) {
	l.locker.Lock()

	if l.verbose {
		fmt.Print(string(b))
	}

	l.data = append(l.data, b...)
	l.locker.Unlock()
}

// AddLn adds EOL as suffix
func (l *Logger) AddLn(b []byte) {
	l.Add(append(b, []byte(util.Eol)...))
}

// AddHeader adds a line like "[2024-12-28 15:16:17] content (@b)"
func (l *Logger) AddHeader(s string) {
	ts := fmt.Sprintf("[%s] %s", time.Now().Format(time.DateTime), s)
	l.AddLn([]byte(ts))
}

// ReadStream reads from a stream & adds to the log
func (l *Logger) ReadStream(stream *io.Reader) {
	util.ReadLinesFromStream(stream, func(b []byte) {
		l.AddLn(b)
	})
}

// Print prints collected longs
func (l *Logger) Print() {
	fmt.Println(string(l.data))
}

// WriteToFile dumps to specified file
func (l *Logger) WriteToFile(filePath string) error {
	// create path if not exist
	err := util.CreatePath(filePath, 0755, true)
	if err != nil {
		return err
	}

	l.locker.Lock()

	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(l.data)
	if err != nil {
		return err
	}

	l.locker.Unlock()
	return nil
}

func New() *Logger {
	return &Logger{}
}
