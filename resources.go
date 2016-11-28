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

	bodytext, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return SystemReturnError
	}

	var logReq LogRequest
	err = json.Unmarshal(bodytext, &logReq)
	if err != nil {
		log.Printf("Unable to marshal body: %v", err)
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
			TWLog(vars["logname"], logReq.Message)
			response.Set("result", "ok")
		}
	} else {
		response.Set("result", "error")
		response.Set("message", "incorrect URL")
	}
	ret, err := response.MarshalJSON()
	if err != nil {
		return SystemReturnError
	}
	return string(ret)
}
