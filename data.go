package main

import "sync"

// ServerConfig holds the server configuration values
type ServerConfig struct {
	LogHome string
	Port    string
	MaxLogs int
	MaxMinutes int
}

// LogConfig contains logger base configuration
type LogConfig struct {
	logPath     string
	pathPrefix  string
	maxMinutes int
	purgeLock   sync.Mutex
}


