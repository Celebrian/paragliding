package main

import (
	"net/http"
	"strconv"
)

func getAllTracks(w http.ResponseWriter, r *http.Request) {
	number, err := collection.Count()
	if err != nil {
		errStatus(w, http.StatusBadRequest, err, "Could not count all tracks for admin")
	}
	_, err = w.Write([]byte(strconv.Itoa(number)))
	if err != nil {
		errStatus(w, http.StatusInternalServerError, err, "Could not write number of tracks to page")
	}

	if r.Method == http.MethodDelete {
		err = collection.DropAllIndexes()
		if err != nil {
			errStatus(w, http.StatusInternalServerError, err, "Could not remove all tracks from database")
		}
	}
}
