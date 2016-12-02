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
