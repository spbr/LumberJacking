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

import "sync"
import "errors"

// LogConfig contains logger base configuration
type LogConfig struct {
	logPath     string
	pathPrefix  string
	maxMinutes int
	purgeLock   sync.Mutex
}

func (conf *LogConfig) setLogPath(logpath string) {
	conf.logPath = logpath + "/"
	conf.pathPrefix = conf.logPath
}
func (conf *LogConfig) setMaxMinutes(maxMinutes int) error {
	if maxMinutes == 0 {
		return errors.New("max minutes cannot be set to zero")
	}
	if maxMinutes > 60 {
		return errors.New("max minutes cannot be greater than zero")
	}
	if 60 % maxMinutes > 0 {
		return errors.New("max minutes must be a divisor of 60")
	}
	conf.maxMinutes = maxMinutes
	return nil
}

func (conf *LogConfig) getMinuteBlock(minutes int) int {
	return (minutes / conf.maxMinutes) * conf.maxMinutes
}
