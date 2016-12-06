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
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"sync"
)

var tgConf = LogConfig{}
var tgLoggers map[string]*Logger

var tconfigFile string
var tserverConfig ServerConfig

func init() {
	flag.StringVar(&tconfigFile, "testconfig", "", "test config file")
}

func TestLog(t *testing.T) {

	flag.Usage = func() {
		fmt.Println("Must supply a configuration file")
		os.Exit(64)
	}

	flag.Parse()

	if tconfigFile == "" {
		flag.Usage()
	}
	file, err := ioutil.ReadFile(tconfigFile)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	err = json.Unmarshal(file, &tserverConfig)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	tgLoggers = make(map[string]*Logger, tserverConfig.MaxLogs)

	err = Init(tserverConfig.LogHome, tserverConfig.MaxMinutes)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	err = os.MkdirAll(tserverConfig.LogHome, 0755)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	tgConf.setLogPath(tserverConfig.LogHome)
	err = tgConf.setMaxMinutes(tserverConfig.MaxMinutes)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	newLogger := Logger{logname: "test"}
	tgLoggers["test"] = &newLogger

	err = WriteLog(tgLoggers, "test", "{\"test\": 1}")
	if err != nil {
		t.Error("Didn't expect error ", err)
	}
}

func TestPerf(t *testing.T) {
	flag.Usage = func() {
		fmt.Println("Must supply a configuration file")
		os.Exit(64)
	}

	flag.Parse()

	if tconfigFile == "" {
		flag.Usage()
	}
	file, err := ioutil.ReadFile(tconfigFile)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	err = json.Unmarshal(file, &tserverConfig)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	tgLoggers = make(map[string]*Logger, tserverConfig.MaxLogs)

	err = Init(tserverConfig.LogHome, tserverConfig.MaxMinutes)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	err = os.MkdirAll(tserverConfig.LogHome, 0755)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	tgConf.setLogPath(tserverConfig.LogHome)
	err = tgConf.setMaxMinutes(tserverConfig.MaxMinutes)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	newLogger := Logger{logname: "test"}
	tgLoggers["test"] = &newLogger

	var wg sync.WaitGroup
	for i := 0; i != 200000; i++ {
		wg.Add(1)
		go func() {
			err = WriteLog(tgLoggers, "test", "{\"test\": 1}")
			if err != nil {
				t.Error("Didn't expect error ", err)
			}
			wg.Add(-1)
		}()
	}
	wg.Wait()
}
