# LumberJacking
LumberJacking is a network logging server written in Go.  It's meant to be used in situations where it's expected the
 log files will be processed by some external process.  Multiple kinds of log files can be created.  For example, you
  can have an "info" log file as well as a "debug" log file.

## Running
 
 All configuration values are passed in via a JSON-based config file
 
    ./TheLumberjack --config ./example.cfg
    
 
 
## Config values

* LogHome    - Where the log files should be written
* IpAddress  - Which address the server should listen on
* Port       - Which port the server should listen on 
* MaxLogs    - Maximum number of log files the server should keep track of 
* MaxMinutes - Log files will be rotated according to this setting.  The value cannot be greater than 60 and must 
	divide evenly into 60`
	
## Endpoints

### Write Log
    POST /log/{logname}
    Content Type: "application/json"
    
### Fetch Logging Stats
    GET /stats