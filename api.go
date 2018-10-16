package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

//Handle requests to /api/, respond with json encoded struct
func handleAPI(w http.ResponseWriter) {

	// APIInfo is a struct for /api/ call
	type APIInfo struct {
		Uptime  string `json:"uptime"`
		Info    string `json:"info"`
		Version string `json:"version"`
	}
	currentTime := time.Now()
	//Create return struct with calculated uptime
	api := APIInfo{uptimeFunc(startTime, currentTime), "Service for paragliding tracks", "v1"}
	w.Header().Set("Content-Type", "application/json")
	//Encode the struct and send to response writer, else handle error
	err := json.NewEncoder(w).Encode(api)
	if err != nil {
		errStatus(w, http.StatusInternalServerError, err, "Could not encode APIInfo to json")
		return
	}
}

//Most of this function was found at:
//https://stackoverflow.com/questions/36530251/golang-time-since-with-months-and-years
//and adapted to this assignment.
//Calculate uptime based on start time and current time
//nolint: gocyclo
func uptimeFunc(start, current time.Time) string {
	//Check if both times are in the same location, if not, then set current time to starts location
	if start.Location() != current.Location() {
		current = current.In(start.Location())
	}
	//With different locations it can happen that start time is after current time. If so, swap them
	if start.After(current) {
		start, current = current, start
	}

	//Extract date
	year1, month1, day1 := start.Date()
	year2, month2, day2 := current.Date()

	//Extract clock
	hour1, minute1, second1 := start.Clock()
	hour2, minute2, second2 := current.Clock()

	//Calculate difference between current- and start-time
	year := int(year2 - year1)
	month := int(month2 - month1)
	day := int(day2 - day1)
	hour := int(hour2 - hour1)
	minute := int(minute2 - minute1)
	second := int(second2 - second1)

	//Normalize negative values
	if second < 0 {
		second += 60
		minute--
	}
	if minute < 0 {
		minute += 60
		hour--
	}
	if hour < 0 {
		hour += 24
		day--
	}
	if day < 0 {
		//Calculate number of days dependent on what month the service was started
		t := time.Date(year1, month1, 32, 0, 0, 0, 0, time.UTC)
		day += 32 - t.Day()
		month--
	}
	if month < 0 {
		month += 12
		year--
	}

	//Construct ISO8601 string
	isoString := "P"
	if year != 0 {
		isoString = fmt.Sprintf("%s%sY", isoString, strconv.Itoa(year))
	}
	if month != 0 {
		isoString = fmt.Sprintf("%s%sM", isoString, strconv.Itoa(month))
	}
	if day != 0 {
		isoString = fmt.Sprintf("%s%sD", isoString, strconv.Itoa(day))
	}
	if hour != 0 || minute != 0 || second != 0 {
		isoString = fmt.Sprintf("%sT", isoString)
	}
	if hour != 0 {
		isoString = fmt.Sprintf("%s%sH", isoString, strconv.Itoa(hour))
	}
	if minute != 0 {
		isoString = fmt.Sprintf("%s%sM", isoString, strconv.Itoa(minute))
	}
	if second != 0 {
		isoString = fmt.Sprintf("%s%sS", isoString, strconv.Itoa(second))
	}
	//If time somehow is 0, return PT0S, not just P
	if isoString == "P" {
		isoString = "PT0S"
	}
	return isoString
}
