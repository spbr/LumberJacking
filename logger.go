/*
TheLumberjack is a network based logging facility which provides a dynamic range of log files
Copyright (C) 2016  Shane P. Brady

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */
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
