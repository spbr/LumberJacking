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

// consts
const (
	maxInt64          = int64(^uint64(0) >> 1)
	logCreatedTimeLen = 21
	logFilenameMinLen = 36
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
		fmt.Printf("%v", err)
		os.Exit(1)
	}

	err = json.Unmarshal(file, &serverConfig)
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}
	gLoggers = make(map[string]*Logger, serverConfig.MaxLogs)

	Init(serverConfig.LogHome, 2000, 2, 20000)

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
	tSaved := time.Now()
	for i := 0; i != 200000; i++ {
		wg.Add(1)
		go func() {
			Server("{\"hit\":{\"_index\":\"establishments\",\"_type\":\"establishment\",\"_id\":\"estab_2641\",\"_score\":null,\"_source\":{\"establishment_id\":2641,\"establishment_name\":\"Somtum Der\",\"neighborhood\":\"east village\",\"cuisines\":[],\"last_message_updated\":0,\"geo\":{\"lon\":-73.9843516,\"lat\":40.7252272}},\"sort\":[0,1658.1558661470444]}}")
			wg.Add(-1)
		}()
	}
	wg.Wait()
	fmt.Println(time.Now().Sub(tSaved))
} */

// Init must be called first, to create the logging server
func Init(logpath string, maxfiles, nfilesToDel int, maxsize uint32) error {
	err := os.MkdirAll(logpath, 0755)
	if err != nil {
		return err
	}

	if maxfiles <= 0 || maxfiles > 100000 {
		return fmt.Errorf("maxfiles must be greater than 0 and less than or equal to 100000: %d", maxfiles)
	}

	if nfilesToDel <= 0 || nfilesToDel > maxfiles {
		return fmt.Errorf("nfilesToDel must be greater than 0 and less than or equal to maxfiles! toDel=%d maxfiles=%d",
			nfilesToDel, maxfiles)
	}

	// get names form the directory `logpath`
	files, err := getDirnames(logpath)
	if err != nil {
		return err
	}

	gConf.setLogPath(logpath)
	gConf.maxfiles = maxfiles
	gConf.curfiles = calcLogfileNum(files)
	gConf.nfilesToDel = nfilesToDel
	gConf.setMaxSize(maxsize)
	return nil
}

func (conf *LogConfig) setFlags(flag uint32, on bool) {
	if on {
		conf.logflags = conf.logflags | flag
	} else {
		conf.logflags = conf.logflags & ^flag
	}
}

func (conf *LogConfig) setMaxSize(maxsize uint32) {
	if maxsize > 0 {
		conf.maxsize = int64(maxsize) * 1024 * 1024
	} else {
		conf.maxsize = maxInt64 - (1024 * 1024 * 1024 * 1024 * 1024)
	}
}

func (conf *LogConfig) setLogPath(logpath string) {
	conf.logPath = logpath + "/"
	conf.pathPrefix = conf.logPath

}

func getMinuteBlock(minutes int) int {
	return (minutes / 15) * 15
}

func init() {
	gConf.setLogPath("./log")
}

// helpers
func getDirnames(dir string) ([]string, error) {
	f, err := os.Open(dir)
	if err == nil {
		defer f.Close()
		return f.Readdirnames(0)
	}
	return nil, err
}

func calcLogfileNum(files []string) int {
	curfiles := 0
	for _ = range files {
		curfiles++
	}
	return curfiles
}

func genLogPrefix(t *time.Time, logname string) string {
	layout := "%04d-%02d-%02d %02d:%02d:%02d - %s: "
	myt := time.Now()
	prefix := fmt.Sprintf(layout, myt.Year(), myt.Month(), myt.Day(), myt.Hour(),
		myt.Minute(), myt.Second(), logname)
	return prefix
}

func TWLog(logname string, logMessage string) {
	t := time.Now()
	prefix := genLogPrefix(&t, logname)
	logEntry := fmt.Sprintf("%s%s\n", prefix, logMessage)
	gLoggers[logname].log(&t, logEntry)
}

var gProgname = path.Base(os.Args[0])

var gConf = LogConfig{
	maxfiles:    400,
	nfilesToDel: 10,
	maxsize:     100 * 1024 * 1024,
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
