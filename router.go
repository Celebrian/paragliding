package main

import (
	"fmt"
	"net/http"
	"regexp"
)

//nolint: gocyclo
func router(w http.ResponseWriter, r *http.Request) {
	//Build regex expressions for the url and handle possible errors
	redirectAPI, err := regexp.Compile("^/paragliding/?$")
	if err != nil {
		errStatus(w, http.StatusInternalServerError, err, "Failed to compile redirect regex")
		return
	}
	apiHandler, err := regexp.Compile("^/paragliding/api/?$")
	if err != nil {
		errStatus(w, http.StatusInternalServerError, err, "Failed to compile api regex")
		return
	}
	apiTrackHandler, err := regexp.Compile("^/paragliding/api/track/?$")
	if err != nil {
		errStatus(w, http.StatusInternalServerError, err, "Failed to compile api/track regex")
		return
	}
	apiTrackIDHandler, err := regexp.Compile("^/paragliding/api/track/[a-f0-9]{24}/?$")
	if err != nil {
		errStatus(w, http.StatusInternalServerError, err, "Failed to compile api/igc/<id> regex")
		return
	}
	apiTrackIDFieldHandler, err := regexp.Compile("^/paragliding/api/track/[a-f0-9]{24}/(pilot|glider|glider_id|track_length|H_date)$")
	if err != nil {
		errStatus(w, http.StatusInternalServerError, err, "Failed to compile api/track/<id>/<field> regex")
		return
	}
	apiTicker, err := regexp.Compile("^/paragliding/api/ticker/?$")
	if err != nil {
		errStatus(w, http.StatusInternalServerError, err, "Failed to compile api/ticker regex")
		return
	}
	apiTickerLatest, err := regexp.Compile("^/paragliding/api/ticker/latest/?$")
	if err != nil {
		errStatus(w, http.StatusInternalServerError, err, "Failed to compile api/ticker/latest regex")
		return
	}
	apiTickerTimestamp, err := regexp.Compile("^/paragliding/api/ticker/latest/[0-9]+/?$")
	if err != nil {
		errStatus(w, http.StatusInternalServerError, err, "Failed to compile api/ticker/timestamp regex")
		return
	}
	apiWebhookNewTrack, err := regexp.Compile("^/paragliding/api/webhook/new_track/?$")
	if err != nil {
		errStatus(w, http.StatusInternalServerError, err, "Failed to compile api/webhook regex")
		return
	}
	apiWebhookID, err := regexp.Compile("^/paragliding/api/webhook/new_track/id/?$")
	if err != nil {
		errStatus(w, http.StatusInternalServerError, err, "Failed to compile api/webhook/id regex")
		return
	}
	adminAPITrackCount, err := regexp.Compile("^/paragliding/admin/api/tracks_count/?$")
	if err != nil {
		errStatus(w, http.StatusInternalServerError, err, "Failed to compile admin/api/track_count regex")
		return
	}
	adminAPITracks, err := regexp.Compile("^/paragliding/admin/tracks/?$")
	if err != nil {
		errStatus(w, http.StatusInternalServerError, err, "Failed to compile admin/api/tracks regex")
	}

	//Check if request is GET, POST or DELETE
	if r.Method == http.MethodGet || r.Method == http.MethodPost || r.Method == http.MethodDelete {
		//Switch on the request url path and select handler
		switch {
		case redirectAPI.MatchString(r.URL.Path):
			http.Redirect(w, r, "/paragliding/api", http.StatusPermanentRedirect)
		case apiHandler.MatchString(r.URL.Path):
			handleAPI(w)
		case apiTrackHandler.MatchString(r.URL.Path):
			//The response should have header type json no matter what
			w.Header().Set("Content-Type", "application/json")
			if r.Method == http.MethodPost {
				trackPOST(w, r)
			} else if r.Method == http.MethodGet {
				trackGET(w, r)
			}
		case apiTrackIDHandler.MatchString(r.URL.Path):
			getOne(w, r)
		case apiTrackIDFieldHandler.MatchString(r.URL.Path):
			getField(w, r)
		case apiTicker.MatchString(r.URL.Path):
		case apiTickerLatest.MatchString(r.URL.Path):
			getLatest(w, r)
		case apiTickerTimestamp.MatchString(r.URL.Path):
		case apiWebhookNewTrack.MatchString(r.URL.Path):
		case apiWebhookID.MatchString(r.URL.Path):
		case adminAPITrackCount.MatchString(r.URL.Path):
		case adminAPITracks.MatchString(r.URL.Path):
		default:
			errStatus(w, http.StatusNotFound, nil, "")
		}
	} else {
		errStatus(w, http.StatusNotImplemented, nil, "")
	}

}

//Write status header and body with status code, error if exist, and possible extra info
func errStatus(w http.ResponseWriter, status int, err error, extraInfo string) {
	if err != nil && extraInfo != "" {
		http.Error(w, fmt.Sprintf("%s\n%s\n%s", http.StatusText(status), err, extraInfo), status)
	} else if extraInfo != "" {
		http.Error(w, fmt.Sprintf("%s\n%s", http.StatusText(status), err), status)
	} else if err != nil {
		http.Error(w, fmt.Sprintf("%s\n%s", http.StatusText(status), extraInfo), status)
	} else {
		http.Error(w, http.StatusText(status), status)
	}
}
