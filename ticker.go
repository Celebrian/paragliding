package main

import (
	"net/http"
	"strconv"
)

func getLatest(w http.ResponseWriter, r *http.Request) {
	var latest IGCObject
	dbSize, err := collection.Count()
	if err != nil {
		errStatus(w, http.StatusInternalServerError, err, "Error getting database size in ticker.go")
		return
	}

	err = collection.Find(nil).skip(dbSize - 1).One(&latest)
	if err != nil {
		errStatus(w, http.StatusInternalServerError, err, "Error getting latest track in ticker.go")
		return
	}
	_, err = w.Write([]byte(strconv.FormatInt(latest.Timestamp, 10)))
	if err != nil {
		errStatus(w, http.StatusInternalServerError, err, "Could not write timestamp")
	}
}
