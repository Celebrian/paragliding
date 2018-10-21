package main

import (
	"github.com/globalsign/mgo"
	"net/http"
	"time"
)

//StartTime is service start time
var startTime = time.Now()

const (
	MongoDBHosts = "ds137643.mlab.com:37643"
	MongoDatabase = "paragliding"
	MongoUser	= "databaseuser"
	MongoPassword = "databasepassword1"

)

type IGCObject struct {
	URL			string	`json:"url"`
	HDate		string	`json:"H_date"`
	Glider		string	`json:"glider"`
	GliderID	string	`json:"glider_id"`
	TrackLength	float64	`json:"track_length"`
}

var db *mgo.Database

func main() {
	mongoDialInfo := &mgo.DialInfo{
		Addrs:    []string{MongoDBHosts},
		Timeout:  60 * time.Second,
		Database: MongoDatabase,
		Username: MongoUser,
		Password: MongoPassword,
	}

	//Database connection
	session, err := mgo.DialWithInfo(mongoDialInfo)
	if err != nil {
		panic(err)
	}
	session.Close()

	db = session.DB(MongoDatabase)

	//Send all requests to the router
	http.HandleFunc("/", router)

	//Start web server
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
