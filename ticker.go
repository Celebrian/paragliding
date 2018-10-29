package main

import (
	"encoding/json"
	"net/http"
	"path"
	"regexp"
	"strconv"
	"time"

	"github.com/globalsign/mgo/bson"
)

//Get the timestamp of the latest track in the database
func getLatest(w http.ResponseWriter) int64 {
	var latest IGCObject
	//Count how many tracks there are total
	dbSize, err := collection.Count()
	if err != nil {
		errStatus(w, http.StatusInternalServerError, err, "Error getting database size in ticker.go")
		return 0
	}

	//Find the latest one by skipping all but 1 and get the last one
	err = collection.Find(nil).Skip(dbSize - 1).One(&latest)
	if err != nil {
		errStatus(w, http.StatusInternalServerError, err, "Error getting latest track in ticker.go")
		return 0
	}
	//Return the timestamp from that track
	return latest.Timestamp
}

//Get the ticker(t_latest, t_start, t_stop, []tracks, processing) and return it to the user
func returnTicker(w http.ResponseWriter, r *http.Request) {
	//ReturnTicker struct for json return
	type ReturnTicker struct {
		TLatest    int64    `json:"t_latest"`
		TStart     int64    `json:"t_start"`
		TStop      int64    `json:"t_stop"`
		Tracks     []string `json:"tracks"`
		Processing int64    `json:"processing"`
	}
	//Header should be json
	w.Header().Set("Content-Type", "application/json")

	//Start the timer so we can calculate the processing time
	startFunction := time.Now()
	//Get the latest timestamp
	latest := getLatest(w)

	var start, stop int64
	var tracks []string
	//If the latest timestamp is 0, quit because there is no tracks in the database
	if latest == 0 {
		errStatus(w, http.StatusRequestedRangeNotSatisfiable, nil, "No latest timestamp in database(returnTicker)")
		return
	}
	//Remove trailing / and regex match the timestamp if there is one
	stamp := path.Clean(r.URL.Path)
	base, err := regexp.MatchString("^[0-9]+$", path.Base(stamp))
	if err != nil {
		errStatus(w, http.StatusBadRequest, err, "No timestamp in base url")
		return
	}
	//If there was not a timestamp, call the function with a timestamp of zero, get paging number of tracks
	if !base {
		start, stop, tracks = getTracks(w, 0, latest, pagingSize)
		//If there was a timestamp, extract that
	} else {
		//Clean the timestamp
		temp := path.Base(stamp)
		//Convert the timestamp to int
		ts, er := strconv.Atoi(temp)
		if er != nil {
			errStatus(w, http.StatusInternalServerError, er, "Could not convert url to int in returnTicker")
			return
		}
		//Call the function with the timestamp as parameter and pagingSize as limit
		start, stop, tracks = getTracks(w, int64(ts), latest, pagingSize)
	}

	//Calculate the processing time in minutes
	processing := int64(time.Now().Sub(startFunction).Seconds() * 1000)

	//Build the struct to encode and return
	temp := ReturnTicker{
		TLatest:    latest,
		TStart:     start,
		TStop:      stop,
		Tracks:     tracks,
		Processing: processing,
	}
	err = json.NewEncoder(w).Encode(temp)
	if err != nil {
		errStatus(w, http.StatusInternalServerError, err, "Failed json encoding for temp in returnTicker")
	}
}

//Get the tracks between <timestamp> and <latest>, max <amount>, return as <start>, <stop> and <[]trackIDs>
func getTracks(w http.ResponseWriter, timestamp int64, latest int64, amount int) (start int64, stop int64, tracks []string) {
	var allTracks []IGCObject
	//If amount is equal to pagingSize, we want to limit our search to that size
	if amount == pagingSize {
		//Find tracks with ID greater than timestamp, but limit search to paging size
		err := collection.Find(bson.M{"timestamp": bson.M{"$gt": timestamp}}).Limit(pagingSize).All(&allTracks)
		if err != nil {
			errStatus(w, http.StatusInternalServerError, err, "collection.indexes")
		}
		//Loop through all tracks gotten until the tracks array is as bit as pagingSize or the last one is the last one in the database
		for i := 0; len(tracks) != pagingSize && stop != latest; i++ {
			//Append the ID of the track to the array
			tracks = append(tracks, allTracks[i].ID.Hex())
			//Set stop as the last timestamp added to the slice, in case this is the last ID to be added
			stop = allTracks[i].Timestamp
		}
		//If we do not want to limit our search to the paging size do the same without limit
	} else {
		//Get all tracks after <timestamp>
		err := collection.Find(bson.M{"timestamp": bson.M{"$gt": timestamp}}).All(&allTracks)
		if err != nil {
			errStatus(w, http.StatusInternalServerError, err, "collection.indexes")
		}
		//Loop through all tracks gotten
		for i := 0; stop != latest; i++ {
			//Append the ID to the slice
			tracks = append(tracks, allTracks[i].ID.Hex())
			//Set stop to the last added timesstamp
			stop = allTracks[i].Timestamp
		}
	}
	//Set start to the first IDs timestamp
	start = allTracks[0].Timestamp

	return start, stop, tracks
}
