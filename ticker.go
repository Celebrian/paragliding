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

func getLatest(w http.ResponseWriter) int64 {
	var latest IGCObject
	dbSize, err := collection.Count()
	if err != nil {
		errStatus(w, http.StatusInternalServerError, err, "Error getting database size in ticker.go")
		return 0
	}

	err = collection.Find(nil).Skip(dbSize - 1).One(&latest)
	if err != nil {
		errStatus(w, http.StatusInternalServerError, err, "Error getting latest track in ticker.go")
		return 0
	}
	return latest.Timestamp
}

func returnTicker(w http.ResponseWriter, r *http.Request) {
	//ReturnTicker struct for json return
	type ReturnTicker struct {
		TLatest    int64    `json:"t_latest"`
		TStart     int64    `json:"t_start"`
		TStop      int64    `json:"t_stop"`
		Tracks     []string `json:"tracks"`
		Processing int64    `json:"processing"`
	}
	w.Header().Set("Content-Type", "application/json")

	startFunction := time.Now()
	latest := getLatest(w)
	var start, stop int64
	var tracks []string
	if latest == 0 {
		errStatus(w, http.StatusRequestedRangeNotSatisfiable, nil, "No latest timestamp in database(returnTicker)")
		return
	}
	base, err := regexp.MatchString("^[0-9]+$", path.Base(r.URL.Path))
	if err != nil {
		errStatus(w, http.StatusBadRequest, err, "No timestamp in base url")
		return
	}
	if !base {
		start, stop, tracks = getTracks(w, 0, latest, pagingSize)

	} else {
		temp := path.Clean(r.URL.Path)
		temp = path.Base(temp)
		ts, er := strconv.Atoi(temp)
		if er != nil {
			errStatus(w, http.StatusInternalServerError, er, "Could not convert url to int in returnTicker")
			return
		}
		start, stop, tracks = getTracks(w, int64(ts), latest, pagingSize)
	}

	processing := int64(time.Now().Sub(startFunction).Seconds() * 1000)

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

func getTracks(w http.ResponseWriter, timestamp int64, latest int64, amount int) (start int64, stop int64, tracks []string) {
	var allTracks []IGCObject
	if amount == pagingSize {
		err := collection.Find(bson.M{"timestamp": bson.M{"$gt": timestamp}}).Limit(pagingSize).All(&allTracks)
		if err != nil {
			errStatus(w, http.StatusInternalServerError, err, "collection.indexes")
		}
		for i := 0; len(tracks) != pagingSize && stop != latest; i++ {
			tracks = append(tracks, allTracks[i].ID.Hex())
			stop = allTracks[i].Timestamp
		}
	} else {
		err := collection.Find(bson.M{"timestamp": bson.M{"$gt": timestamp}}).All(&allTracks)
		if err != nil {
			errStatus(w, http.StatusInternalServerError, err, "collection.indexes")
		}
		for i := 0; stop != latest; i++ {
			tracks = append(tracks, allTracks[i].ID.Hex())
			stop = allTracks[i].Timestamp
		}
	}

	start = allTracks[0].Timestamp

	return start, stop, tracks
}
