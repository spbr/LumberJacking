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

// Server logs a server response
import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/bitly/go-simplejson"
	"github.com/gorilla/mux"
)

// LogRequest exists to hold the request needed to log
type LogRequest struct {
	Message string
}

// SystemReturnError is a last resort response
var SystemReturnError = "\"result\": \"error\", \"message\": \"Fatal System Error\""

// Log takes the log request and logs it to the correct service
func Log(req *http.Request, writer http.ResponseWriter, logName string) string {

	gStats.IncRequests()
	bodytext, err := ioutil.ReadAll(req.Body)
	if err != nil {
		gStats.IncErrors()
		return SystemReturnError
	}

	var logReq LogRequest
	err = json.Unmarshal(bodytext, &logReq)
	if err != nil {
		log.Printf("Unable to marshal body: %v", err)
		gStats.IncErrors()
		return SystemReturnError
	}

	writer.Header().Set("Content-Type", "application/json")
	response := simplejson.New()

	vars := mux.Vars(req)
	if _, ok := vars["logname"]; ok {
		err = FindOrCreateLogEntry(vars["logname"])
		if err != nil {
			response.Set("result", "error")
			response.Set("message", err.Error())
		} else {
			err = SPiBLog(vars["logname"], logReq.Message)
			if err != nil {
				response.Set("result", "error")
				response.Set("message", err.Error())
				gStats.IncErrors()
			} else {
				gStats.IncLogsWritten()
				response.Set("result", "ok")
			}
		}
	} else {
		response.Set("result", "error")
		response.Set("message", "incorrect URL")
		gStats.IncErrors()
	}
	ret, err := response.MarshalJSON()
	if err != nil {
		gStats.IncErrors()
		return SystemReturnError
	}
	return string(ret)
}

// FetchStats returns the logging stats since the app was started
func FetchStats(req *http.Request, writer http.ResponseWriter) string {

	writer.Header().Set("Content-Type", "application/json")
	ret, err := gStats.ToJSONString()
	if err != nil {
		gStats.IncErrors()
		return SystemReturnError
	}
	return ret
}