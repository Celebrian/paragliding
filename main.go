package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/globalsign/mgo/bson"

	"github.com/globalsign/mgo"
)

//StartTime is service start time
var startTime = time.Now()

// IGCObject is the struct for track information we have deemed relevant
type IGCObject struct {
	ID          bson.ObjectId `json:"id" bson:"_id"`
	URL         string        `json:"url" bson:"track_scr_url"`
	Pilot       string        `json:"pilot" bson:"pilot"`
	HDate       time.Time     `json:"H_date" bson:"H_date"`
	Glider      string        `json:"glider" bson:"glider"`
	GliderID    string        `json:"glider_id" bson:"glider_id"`
	TrackLength float64       `json:"track_length" bson:"track_length"`
	Timestamp   int64         `json:"timestamp" bson:"timestamp"`
}

var db *mgo.Database

var collection *mgo.Collection

func main() {
	fmt.Println("MAIN STARTED")
	mongoDialInfo := &mgo.DialInfo{
		Addrs:    []string{os.Getenv("MONGO_HOST")},
		Timeout:  60 * time.Second,
		Database: os.Getenv("MONGO_DATABASE"),
		Username: os.Getenv("MONGO_USER"),
		Password: os.Getenv("MONGO_PASSWORD"),
	}

	//Database connection
	session, err := mgo.DialWithInfo(mongoDialInfo)
	if err != nil {
		panic(err)
	}

	db = session.DB(mongoDialInfo.Database)
	collection = db.C("tracks")

	//Send all requests to the router
	http.HandleFunc("/", router)

	//Start web server
	if err := http.ListenAndServe(":"+os.Getenv("PORT"), nil); err != nil {
		panic(err)
	}
}
