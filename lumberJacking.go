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
	"sync"
	"time"

	"github.com/gorilla/mux"
	"errors"
)

// consts
const (
	maxInt64 = int64(^uint64(0) >> 1)
	logCreatedTimeLen = 21
	logFilenameMinLen = 36
)

// Config holds the server configuration values
type ServerConfig struct {
	LogHome string
	Port    string
    LogLevelMax int
}

// logger configuration
type config struct {
	logPath     string
	pathPrefix  string
	logflags    uint32
	maxfiles    int   // limit the number of log files under `logPath`
	curfiles    int   // number of files under `logPath` currently
	nfilesToDel int   // number of files deleted when reaching the limit of the number of log files
	maxsize     int64 // limit size of a log file
	purgeLock   sync.Mutex
}

// logger
type logger struct {
	file    *os.File
	tag     string
	logname string
	year    int
	day     int
	month   int
	hour    int
	size    int64
	lock    sync.Mutex
}
/*
** Global vars
 */
var gLoggers map[string]*logger

var wg sync.WaitGroup
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
    gLoggers = make(map[string]*logger, serverConfig.LogLevelMax)

	Init(serverConfig.LogHome, 2000, 2, 20000)

	/*
	 ** Let's start the server
	 */

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/log/{logname}", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, Log(req, w, "info"))
	}).Methods("POST")

	log.Println("Starting server on 127.0.0.1:" + serverConfig.Port)
	log.Fatal(http.ListenAndServe("127.0.0.1:" + serverConfig.Port, router))
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

func (conf *config) setFlags(flag uint32, on bool) {
	if on {
		conf.logflags = conf.logflags | flag
	} else {
		conf.logflags = conf.logflags & ^flag
	}
}

func (conf *config) setMaxSize(maxsize uint32) {
	if maxsize > 0 {
		conf.maxsize = int64(maxsize) * 1024 * 1024
	} else {
		conf.maxsize = maxInt64 - (1024 * 1024 * 1024 * 1024 * 1024)
	}
}

func (conf *config) setLogPath(logpath string) {
	conf.logPath = logpath + "/"
	conf.pathPrefix = conf.logPath

}

func (l *logger) log(t *time.Time, data string) {
	mins := getMinuteBlock(t.Minute())
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
		// reaches limit of number of log files
		filename := fmt.Sprintf("%s%s.%s.log.%04d-%02d-%02d-%02d-%02d", gConf.pathPrefix, l.logname,
			hostname, t.Year(), t.Month(), t.Day(), t.Hour(), mins)
		newfile, err := os.OpenFile(filename, os.O_CREATE | os.O_WRONLY | os.O_APPEND, 0644)
		if err != nil {
			log.Printf("Error opening log file: %s - %s", filename, err)
			return
		}
		gConf.curfiles++
		gConf.purgeLock.Unlock()
		hasLocked = false

		l.file.Close()
		l.file = newfile
		l.tag = tag
		l.size = 0

	}

	n, _ := l.file.WriteString(data)
	l.size += int64(n)
}

func getMinuteBlock(minutes int) int {
	return (minutes / 15) * 15
}

// sort files by created time embedded in the filename
type byCreatedTime []string

func (a byCreatedTime) Len() int {
	return len(a)
}

func (a byCreatedTime) Less(i, j int) bool {
	s1, s2 := a[i], a[j]
	if len(s1) < logFilenameMinLen {
		return true
	} else if len(s2) < logFilenameMinLen {
		return false
	} else {
		return s1[len(s1) - logCreatedTimeLen:] < s2[len(s2) - logCreatedTimeLen:]
	}
}

func (a byCreatedTime) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

// init is called after all the variable declarations in the package have evaluated their initializers,
// and those are evaluated only after all the imported packages have been initialized.
// Besides initializations that cannot be expressed as declarations, a common use of init functions is to verify
// or repair correctness of the program state before real execution begins.
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

var gConf = config{
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
	if len(gLoggers) == serverConfig.LogLevelMax {
		return errors.New("Maximum Log Entries created")
	} else {
		newLogger := logger{logname: logname}
		gLoggers[logname] = &newLogger
		return nil
	}
}
