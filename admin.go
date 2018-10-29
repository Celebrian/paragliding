package main

import (
	"net/http"
	"strconv"
)

//Get the number of tracks in the database, and if the method is delete, delete all of them
func getAllTracks(w http.ResponseWriter, r *http.Request) {
	number, err := collection.Count()
	if err != nil {
		errStatus(w, http.StatusBadRequest, err, "Could not count all tracks for admin")
	}
	_, err = w.Write([]byte(strconv.Itoa(number)))
	if err != nil {
		errStatus(w, http.StatusInternalServerError, err, "Could not write number of tracks to page")
	}
	//If method=delete, remove all tracks
	if r.Method == http.MethodDelete {
		err = collection.DropAllIndexes()
		if err != nil {
			errStatus(w, http.StatusInternalServerError, err, "Could not remove all tracks from database")
		}
	}
}
