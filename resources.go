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