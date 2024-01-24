package logger

import (
	"github.com/apudiu/server-backup/internal/util"
	"os"
	"sync"
)

type Logger struct {
	locker sync.Mutex
	data   []byte
}

func (l *Logger) Add(b []byte) {
	l.locker.Lock()
	l.data = append(l.data, b...)
	l.locker.Unlock()
}

func (l *Logger) AddWithLn(b []byte) {
	l.Add(append(b, []byte(util.EolChar())...))
}

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
