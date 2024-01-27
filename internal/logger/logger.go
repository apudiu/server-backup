package logger

import (
	"bufio"
	"fmt"
	"github.com/apudiu/server-backup/internal/util"
	"io"
	"os"
	"sync"
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
		fmt.Println(string(b))
	}

	l.data = append(l.data, b...)
	l.locker.Unlock()
}

// AddLn adds EOL as suffix
func (l *Logger) AddLn(b []byte) {
	l.Add(append(b, []byte(util.EolChar())...))
}

// ReadStream reads from a stream & adds to the log
func (l *Logger) ReadStream(stream *io.Reader) {
	scanner := bufio.NewScanner(*stream)
	for scanner.Scan() {
		l.Add(scanner.Bytes())
	}
}

// Print prints collected longs
func (l *Logger) Print() {
	fmt.Println(string(l.data))
}

// LogToFile dumps to specified file
func (l *Logger) LogToFile(filePath string) error {
	l.locker.Lock()

	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() {
		err = f.Close()
	}()

	_, err = f.Write(l.data)
	if err != nil {
		return err
	}

	l.locker.Unlock()
	return nil
}
