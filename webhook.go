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
	type newWebhook struct {
		ID              bson.ObjectId `bson:"_id,omitempty"`
		WebhookURL      string        `json:"webhookURL" bson:"webhookURL"`
		MinTriggerValue int           `json:"minTriggerValue" bson:"minTriggerValue"`
		Latest          int64         `bson:"latest"`
	}

	var hook newWebhook
	err := json.NewDecoder(r.Body).Decode(&hook)
	if err != nil {
		errStatus(w, http.StatusInternalServerError, err, "Could not decode new webhook in newTrack")
		return
	}
	hook.ID = bson.NewObjectId()
	hook.Latest = getLatest(w)

	err = webhooks.Insert(&hook)
	if err != nil {
		if err != nil {
			errStatus(w, http.StatusInternalServerError, err, "Could not insert hook into database in newTrack")
			return
		}
	}

	_, err = w.Write([]byte(hook.ID.Hex()))
	if err != nil {
		errStatus(w, http.StatusInternalServerError, err, "Could not write ID in newTrack")
		return
	}

}

func invokeWebhooks(w http.ResponseWriter) {
	type Webhook struct {
		ID              bson.ObjectId `bson:"_id,omitempty"`
		WebhookURL      string        `json:"webhookURL" bson:"webhookURL"`
		MinTriggerValue int           `json:"minTriggerValue" bson:"minTriggerValue"`
		Latest          int64         `bson:"latest"`
	}
	type payloadStruct struct {
		Payload string `json:"content"`
	}

	var hooks []Webhook
	webhooks.Find(nil).All(&hooks)
	latest := getLatest(w)

	for _, hook := range hooks {
		timeStart := time.Now()
		_, _, allTracks := getTracks(w, hook.Latest, latest, 0)
		if len(allTracks) != 0 {
			returnString := "Latest timestamp: " + strconv.Itoa(int(latest)) + ", " + strconv.Itoa(len(allTracks)) + " new tracks are: "
			for _, track := range allTracks {
				returnString = fmt.Sprintf("%s, %s", returnString, track)
			}
			processing := int64(time.Now().Sub(timeStart).Seconds() * 1000)
			returnString = fmt.Sprintf("%s. (processing %dms)", returnString, processing)

			payload := payloadStruct{
				Payload: returnString,
			}

			jsonPayload, _ := json.Marshal(payload)
			_, err := http.Post(hook.WebhookURL, "application/json", bytes.NewBuffer(jsonPayload))
			if err != nil {
				errStatus(w, http.StatusInternalServerError, err, "Could not POST webhook")
			}
			webhooks.Update(bson.M{"_id": hook.ID}, bson.M{"$set": bson.M{"latest": latest}})
		}
	}
}

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
	hookID := path.Clean(r.URL.Path)
	hookID = path.Base(hookID)

	var hook WebhookBson

	webhooks.FindId(bson.ObjectIdHex(hookID)).One(&hook)
	if hook.ID == "" {
		errStatus(w, http.StatusBadRequest, nil, "No webhook with that ID")
		return
	}
	hookJSON := WebhookJSON{
		WebhookURL:      hook.WebhookURL,
		MinTriggerValue: hook.MinTriggerValue,
	}
	err := json.NewEncoder(w).Encode(hookJSON)
	if err != nil {
		errStatus(w, http.StatusBadRequest, err, "Could not encode that webhook to json")
	}

	if r.Method == http.MethodDelete {
		webhooks.DropIndex(hook.ID.Hex())
	}
}
