package main

import (
	"fmt"
	"os"
	"log"
	"time"
	"sync"
	"errors"
)


// Logger is the actual logging structure
type Logger struct {
	file    *os.File
	tag     string
	logname string
	year    int
	day     int
	month   int
	hour    int
	lock    sync.Mutex
}

func (l *Logger) log(t *time.Time, data string) error {
	mins := gConf.getMinuteBlock(t.Minute())
	tag := fmt.Sprintf("%04d%02d%02d%02d%02d", t.Year(), t.Month(), t.Day(), t.Hour(), mins)
	l.lock.Lock()
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}
	defer l.lock.Unlock()
	if l.tag == "" || l.tag != tag || l.file == nil {
		gConf.purgeLock.Lock()
		hasLocked := true
		defer func() {
			if hasLocked {
				gConf.purgeLock.Unlock()
			}
		}()
		filename := fmt.Sprintf("%s%s.%s.log.%04d-%02d-%02d-%02d-%02d", gConf.pathPrefix, l.logname,
			hostname, t.Year(), t.Month(), t.Day(), t.Hour(), mins)
		newfile, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Printf("Error opening log file: %s - %s", filename, err)
			return err
		}
		gConf.purgeLock.Unlock()
		hasLocked = false

		l.file.Close()
		l.file = newfile
		l.tag = tag

	}

	written, err := l.file.WriteString(data)
	if written != len(data) {
		return errors.New("Unable to write full data")
	}
	return err
}
