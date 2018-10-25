package main

import (
	"encoding/json"
	"net/http"
	"path"
	"strconv"

	"github.com/globalsign/mgo/bson"
)

func getOne(w http.ResponseWriter, r *http.Request) {
	id := path.Base(r.URL.Path)
	var track IGCObject

	//The response should have header type json
	w.Header().Set("Content-Type", "application/json")
	err := collection.FindId(bson.ObjectIdHex(id)).One(&track)
	if err != nil {
		errStatus(w, http.StatusBadRequest, err, "No track with that ID in the database")
		return
	}
	err = json.NewEncoder(w).Encode(track)
	if err != nil {
		errStatus(w, http.StatusInternalServerError, err, "Failed to encode return json payload")
		return
	}
}

//getField handles requests for a specific field given an ID
//nolint: gocyclo
func getField(w http.ResponseWriter, r *http.Request) {
	var track IGCObject

	//Clean path to remove trailing /
	temp := path.Clean(r.URL.Path)
	//Extract ID and convert to int
	temp = path.Dir(temp)
	id := path.Base(temp)
	//Check if ID exists in database, if not write 404 error
	err := collection.FindId(bson.ObjectIdHex(id)).One(&track)
	if err != nil {
		errStatus(w, http.StatusInternalServerError, err, "Could not get a single id from the database")
		return
	}

	//Switch based on the final part of the url, write the applicable field, handle possible errors
	switch path.Base(r.URL.Path) {
	case "pilot":
		_, err = w.Write([]byte(track.Pilot))
		if err != nil {
			errStatus(w, http.StatusInternalServerError, err, "Error writing Pilot in <field>")
			return
		}
	case "glider":
		_, err = w.Write([]byte(track.Glider))
		if err != nil {
			errStatus(w, http.StatusInternalServerError, err, "Error writing glider in <field>")
			return
		}
	case "glider_id":
		_, err = w.Write([]byte(track.GliderID))
		if err != nil {
			errStatus(w, http.StatusInternalServerError, err, "Error writing glider_id in <field>")
			return
		}
	case "track_length":
		_, err = w.Write([]byte(strconv.Itoa(int(track.TrackLength))))
		if err != nil {
			errStatus(w, http.StatusInternalServerError, err, "Error writing track_length in <field>")
			return
		}
	case "H_date":
		time, err := track.HDate.MarshalText()
		if err != nil {
			errStatus(w, http.StatusInternalServerError, err, "Error converting track date to []byte in <field>")
		}
		_, err = w.Write(time)
		if err != nil {
			errStatus(w, http.StatusInternalServerError, err, "Error writing H_date in <field>")
			return
		}
	default:
		//Should never happen
		panic("Should never happen regex/field handling has failed, the end is nigh")
	}
}
