package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/globalsign/mgo/bson"

	"github.com/marni/goigc"
)

func trackPOST(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	//ReturnID is the format for the returned id after an inserted track
	type ReturnID struct {
		ID string `json:"id"`
	}

	//Extract request body
	postBody := r.Body
	defer r.Body.Close()
	var temp IGCObject

	//Extract and decode json payload containing URL, handle possible error
	err := json.NewDecoder(postBody).Decode(&temp)
	if err != nil {
		errStatus(w, http.StatusBadRequest, err, "Decoding of Request body failed")
		return
	}

	//Check URL, first with url.parse, then regex
	_, err = url.Parse(temp.URL)
	if err != nil {
		errStatus(w, http.StatusBadRequest, err, "Not a valid URL")
	}
	works, err := regexp.MatchString("^http.+\\.igc$", temp.URL)
	if err != nil {
		errStatus(w, http.StatusBadRequest, err, "Regex matchString failed for url")
		return
	}
	if !works {
		errStatus(w, http.StatusBadRequest, nil, "Not a valid URL(regex)")
		return
	}

	//Extract URL and try to parse it
	uri := temp.URL
	track, err := igc.ParseLocation(uri)
	if err != nil {
		errStatus(w, http.StatusBadRequest, err, "Failed to parse the URL")
		return
	}

	//Create and fill the document for insertion into the database
	temp = IGCObject{
		ID:          bson.NewObjectId(),
		Timestamp:   time.Now().UnixNano() / int64(time.Millisecond),
		URL:         uri,
		Pilot:       track.Pilot,
		HDate:       track.Header.Date,
		Glider:      track.GliderType,
		GliderID:    track.GliderID,
		TrackLength: lengthCalc(track),
	}

	//Insert into database
	err = collection.Insert(&temp)
	if err != nil {
		errStatus(w, http.StatusInternalServerError, err, "Failed to insert track into collection")
	}
	//Create and return document id to user
	retID := ReturnID{temp.ID.Hex()}
	err = json.NewEncoder(w).Encode(retID)
	if err != nil {
		errStatus(w, http.StatusInternalServerError, err, "Failed to encode return json payload")
		return
	}
	invokeWebhooks(w)
}

func trackGET(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	var allTracks []IGCObject
	var allIDs []string
	err := collection.Find(nil).All(&allTracks)
	if err != nil {
		errStatus(w, http.StatusInternalServerError, err, "collection.indexes")
	}
	for i := 0; i < len(allTracks); i++ {
		allIDs = append(allIDs, allTracks[i].ID.Hex())
	}
	err = json.NewEncoder(w).Encode(allIDs)
	if err != nil {
		errStatus(w, http.StatusInternalServerError, err, "Failed json encoding for allIDs in trackGET")
	}
}

//lengthCalc calculates track length based on example found here:
//https://github.com/marni/goigc/blob/master/doc_test.go
func lengthCalc(track igc.Track) float64 {
	totalDistance := 0.0
	for i := 0; i < len(track.Points)-1; i++ {
		totalDistance += track.Points[i].Distance(track.Points[i+1])
	}
	return totalDistance
}
