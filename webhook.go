package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/globalsign/mgo/bson"
)

func newTrack(w http.ResponseWriter, r *http.Request) {
	//Struct with info about a webhook
	type newWebhook struct {
		ID              bson.ObjectId `bson:"_id,omitempty"`
		WebhookURL      string        `json:"webhookURL" bson:"webhookURL"`
		MinTriggerValue int           `json:"minTriggerValue" bson:"minTriggerValue"`
		Latest          int64         `bson:"latest"`
	}
	//Decode json payload into a variable
	var hook newWebhook
	err := json.NewDecoder(r.Body).Decode(&hook)
	if err != nil {
		errStatus(w, http.StatusInternalServerError, err, "Could not decode new webhook in newTrack")
		return
	}
	//Create and set the ID for the webhook and get the latest timestamp
	hook.ID = bson.NewObjectId()
	hook.Latest = getLatest(w)

	//Insert hook into the database
	err = webhooks.Insert(&hook)
	if err != nil {
		if err != nil {
			errStatus(w, http.StatusInternalServerError, err, "Could not insert hook into database in newTrack")
			return
		}
	}

	//Write the ID back to the user
	_, err = w.Write([]byte(hook.ID.Hex()))
	if err != nil {
		errStatus(w, http.StatusInternalServerError, err, "Could not write ID in newTrack")
		return
	}

}

//Function for invoking webhooks each time a new track is added
func invokeWebhooks(w http.ResponseWriter) {
	//Struct for extracting the webhook from database
	type Webhook struct {
		ID              bson.ObjectId `bson:"_id,omitempty"`
		WebhookURL      string        `json:"webhookURL" bson:"webhookURL"`
		MinTriggerValue int           `json:"minTriggerValue" bson:"minTriggerValue"`
		Latest          int64         `bson:"latest"`
	}
	//Payload for posting webhook
	type payloadStruct struct {
		Payload string `json:"content"`
	}

	//Extract all webhooks from the database and get the latest timestamp
	var hooks []Webhook
	webhooks.Find(nil).All(&hooks)
	latest := getLatest(w)

	//Loop through all hooks
	for _, hook := range hooks {
		//Start timer for hook processing time
		timeStart := time.Now()
		//Get all tracks that happened after the hooks last update
		_, _, allTracks := getTracks(w, hook.Latest, latest, 0)
		//If the number of tracks are not zero and it is equal or more than the hooks trigger value
		if len(allTracks) != 0 && len(allTracks) >= hook.MinTriggerValue {
			//Construct the payload to send to the webhook
			returnString := "Latest timestamp: " + strconv.Itoa(int(latest)) + ", " + strconv.Itoa(len(allTracks)) + " new tracks are: "
			for _, track := range allTracks {
				returnString = fmt.Sprintf("%s, %s", returnString, track)
			}
			//Calculate how long in ms it has taken to do this and add it to the payload
			processing := int64(time.Now().Sub(timeStart).Seconds() * 1000)
			returnString = fmt.Sprintf("%s. (processing %dms)", returnString, processing)

			//Create the payload
			payload := payloadStruct{
				Payload: returnString,
			}

			//Marshal and post the payload
			jsonPayload, _ := json.Marshal(payload)
			_, err := http.Post(hook.WebhookURL, "application/json", bytes.NewBuffer(jsonPayload))
			if err != nil {
				errStatus(w, http.StatusInternalServerError, err, "Could not POST webhook")
			}
			//Update the webhook with what is now the latest timestamp
			webhooks.Update(bson.M{"_id": hook.ID}, bson.M{"$set": bson.M{"latest": latest}})
		}
	}
}

//Get and show/delete a webhook
func getWebhook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	type WebhookBson struct {
		ID              bson.ObjectId `bson:"_id,omitempty"`
		WebhookURL      string        `json:"webhookURL" bson:"webhookURL"`
		MinTriggerValue int           `json:"minTriggerValue" bson:"minTriggerValue"`
	}
	type WebhookJSON struct {
		WebhookURL      string `json:"webhookURL" bson:"webhookURL"`
		MinTriggerValue int    `json:"minTriggerValue" bson:"minTriggerValue"`
	}
	//Clean and separate the webhook ID that is in the URL
	hookID := path.Clean(r.URL.Path)
	hookID = path.Base(hookID)

	var hook WebhookBson

	//Try to get the webhook from the database. If there is not a webhook with that ID, show error and return
	webhooks.FindId(bson.ObjectIdHex(hookID)).One(&hook)
	if hook.ID == "" {
		errStatus(w, http.StatusBadRequest, nil, "No webhook with that ID")
		return
	}
	//Use the webhook info to get the info to show back to the user
	hookJSON := WebhookJSON{
		WebhookURL:      hook.WebhookURL,
		MinTriggerValue: hook.MinTriggerValue,
	}
	//Encode and post as a json object
	err := json.NewEncoder(w).Encode(hookJSON)
	if err != nil {
		errStatus(w, http.StatusBadRequest, err, "Could not encode that webhook to json")
	}

	//If the method was delete, also drop the webhook from the database
	if r.Method == http.MethodDelete {
		webhooks.DropIndex(hook.ID.Hex())
	}
}
