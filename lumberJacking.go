/** Author: Shane P. Brady
 **    LumberJacking is a network based logging facility which provides a dynamic range of log files
 */
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"errors"
	"github.com/gorilla/mux"
)

/*
** Global vars
 */
var gLoggers map[string]*Logger
var gConf = LogConfig{}
var gStats = Stats{}
var configFile string
var serverConfig ServerConfig

// initialization function
func init() {
	gConf.setLogPath("./log")
	gStats.Init()
	flag.StringVar(&configFile, "config", "", "config file")
}

func main() {

	flag.Usage = func() {
		fmt.Println("Must supply a configuration file")
		os.Exit(64)
	}

	flag.Parse()

	if configFile == "" {
		flag.Usage()
	}
	file, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	err = json.Unmarshal(file, &serverConfig)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	gLoggers = make(map[string]*Logger, serverConfig.MaxLogs)

	err = Init(serverConfig.LogHome, serverConfig.MaxMinutes)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	/*
	 ** Let's start the server
	 */

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/log/{logname}", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, Log(req, w, "info"))
	}).Methods("POST")
	router.HandleFunc("/stats", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, FetchStats(req, w))
	}).Methods("GET")

	log.Println("Starting server on 127.0.0.1:" + serverConfig.Port)
	log.Fatal(http.ListenAndServe("127.0.0.1:"+serverConfig.Port, router))
}

// Init must be called first, to create the logging server
func Init(logpath string, maxMinutes int) error {
	err := os.MkdirAll(logpath, 0755)
	if err != nil {
		return err
	}

	gConf.setLogPath(logpath)
	err = gConf.setMaxMinutes(maxMinutes)
	if err != nil {
		return err
	}
	return nil
}

func genLogPrefix(t *time.Time, logname string) string {
	layout := "%04d-%02d-%02d %02d:%02d:%02d - %s: "
	myt := time.Now()
	prefix := fmt.Sprintf(layout, myt.Year(), myt.Month(), myt.Day(), myt.Hour(),
		myt.Minute(), myt.Second(), logname)
	return prefix
}

// SPiBLog writes out a log file using the global loggers structure
func SPiBLog(logname string, logMessage string) error {
	return WriteLog(gLoggers, logname, logMessage)
}
// WriteLog writs out the log entry, and takes a loggers array as an argument
func WriteLog(gLoggers map[string]*Logger, logname string, logMessage string) error {
	t := time.Now()
	prefix := genLogPrefix(&t, logname)
	logEntry := fmt.Sprintf("%s%s\n", prefix, logMessage)
	err := gLoggers[logname].log(&t, logEntry)
	return err
}

// FindOrCreateLogEntry checks to see if the log name exists, if so, nothing happens.  If not, a new logger is created.
// If there are too many log entries, an error is returned
func FindOrCreateLogEntry(logname string) error {
	if _, ok := gLoggers[logname]; ok {
		return nil
	}
	if len(gLoggers) == serverConfig.MaxLogs {
		return errors.New("Maximum Log Entries created")
	}
	newLogger := Logger{logname: logname}
	gLoggers[logname] = &newLogger
	return nil
}
