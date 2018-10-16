package main

import (
	"net/http"
	"time"
)

//StartTime is service start time
var startTime = time.Now()

func main() {

	//Send all requests to the router
	http.HandleFunc("/", router)

	//Start web server
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
