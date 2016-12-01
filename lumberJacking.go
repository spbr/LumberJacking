/* Author: Shane P. Brady

LumberJacking is a network based logging facility which provides a dynamic range of log files

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
	"path"
	"time"

	"errors"
	"github.com/gorilla/mux"
)

/*
** Global vars
 */
var gLoggers map[string]*Logger

var configFile string
var serverConfig ServerConfig

// initialization function
func init() {
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

	log.Println("Starting server on 127.0.0.1:" + serverConfig.Port)
	log.Fatal(http.ListenAndServe("127.0.0.1:"+serverConfig.Port, router))
}

/*func test() {
	var wg sync.WaitGroup
	tSaved := time.Now()
	for i := 0; i != 200000; i++ {
		wg.Add(1)
		go func() {
			TWLog("test", "{\"hit\"}");
			wg.Add(-1)
		}()
	}
	wg.Wait()
	fmt.Println(time.Now().Sub(tSaved))
} */

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

func getMinuteBlock(gConf *LogConfig, minutes int) int {
	return (minutes / gConf.maxMinutes) * gConf.maxMinutes
}

func init() {
	gConf.setLogPath("./log")
}

func genLogPrefix(t *time.Time, logname string) string {
	layout := "%04d-%02d-%02d %02d:%02d:%02d - %s: "
	myt := time.Now()
	prefix := fmt.Sprintf(layout, myt.Year(), myt.Month(), myt.Day(), myt.Hour(),
		myt.Minute(), myt.Second(), logname)
	return prefix
}

func TWLog(logname string, logMessage string) error {
	return twlog(gLoggers, logname, logMessage)
}
func twlog(gLoggers map[string]*Logger, logname string, logMessage string) error {
	t := time.Now()
	prefix := genLogPrefix(&t, logname)
	logEntry := fmt.Sprintf("%s%s\n", prefix, logMessage)
	err := gLoggers[logname].log(&t, logEntry)
	return err
}

var gProgname = path.Base(os.Args[0])

var gConf = LogConfig{
}

// CheckForLogEntry checks to see if the log name exists, if so, nothing happens.  If not, a new logger is created.
// If there are too many log entries, an error is returned
func FindOrCreateLogEntry(logname string) error {
	if _, ok := gLoggers[logname]; ok {
		return nil
	}
	if len(gLoggers) == serverConfig.MaxLogs {
		return errors.New("Maximum Log Entries created")
	} else {
		newLogger := Logger{logname: logname}
		gLoggers[logname] = &newLogger
		return nil
	}
}
