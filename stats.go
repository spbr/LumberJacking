package main

import (
	"sync"
	"time"
	"fmt"
	"github.com/bitly/go-simplejson"
)

// Stats contains the basic stats the server keeps track of
type Stats struct {
	startupTime int64
	requests    int64
	logsWritten int64
	errors      int64
	lock        sync.Mutex
}

// Init initializes the stats structure, and sets the startup time
func (s *Stats) Init() {
	s.startupTime = time.Now().Unix()
	s.requests = 0
	s.logsWritten = 0
	s.errors = 0
}

// IncRequests increments the number of total requests
func (s *Stats) IncRequests() {
	s.lock.Lock()
	s.requests++
	s.lock.Unlock()
}

// IncLogsWritten increments the number of actual log entries written
func (s *Stats) IncLogsWritten() {
	s.lock.Lock()
	s.logsWritten++
	s.lock.Unlock()
}

// IncErrors increments the number of errors encountered
func (s *Stats) IncErrors() {
	s.lock.Lock()
	s.errors++
	s.lock.Unlock()
}

// ToString returns the stats in a string representation
func (s *Stats) ToString() string {
	return fmt.Sprintf("Startup Time: %12d\nRequests: %12d\nLogs Written: %12d\nErrors: %12d",
	s.startupTime, s.requests, s.logsWritten, s.errors)
}

// ToJSONString returns the stats in a JSON string representation
func (s *Stats) ToJSONString() (string, error) {
	stats := simplejson.New()
	t := time.Unix(s.startupTime, 0)

	stats.Set("startup_time", t.Format(time.RFC1123))
	stats.Set("requests", s.requests)
	stats.Set("logsWritten", s.logsWritten)
	stats.Set("errors", s.errors)

	ret, err := stats.MarshalJSON()
	if err != nil {
		return "", err
	}
	return string(ret), nil
}

