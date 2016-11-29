package main

import "sync"

// ServerConfig holds the server configuration values
type ServerConfig struct {
	LogHome string
	Port    string
	MaxLogs int
}

// LogConfig contains logger base configuration
type LogConfig struct {
	logPath     string
	pathPrefix  string
	logflags    uint32
	maxfiles    int   // limit the number of log files under `logPath`
	curfiles    int   // number of files under `logPath` currently
	nfilesToDel int   // number of files deleted when reaching the limit of the number of log files
	maxsize     int64 // limit size of a log file
	purgeLock   sync.Mutex
}


