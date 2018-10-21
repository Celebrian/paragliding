# Assignment 2: IGC track viewer extended


## About

Similar to Assignment 1, we will develop an online service that will allow users to browse information about IGC files. IGC is an international file format for soaring track files that are used by paragliders and gliders. The program will store IGC files metadata in a NoSQL Database (persistent storage). The system will generate events and it will monitor for new events happening from the outside services. The project will make use of Heroku, OpenStack, and AWS Cloud functions. 

The system must be deployed on Heroku, local SkyHigh OpenStack infrastructure, and as a Cloud Function with AWS. The Go source code must be available for inspection by the teaching staff (read-only access is sufficient).

You can re-use Assignment 1 codebase, and substitute the internal in-memory storage with proper DB query subsystem to request information about the IGC tracks. YOU DO NOT NEED to store IGC files in the Database. In fact, you should NOT store them in a Database. All you need to store is the meta information about the IGC track that has been "uploaded". The file itself, after processing, can be discarded. You will keep the associated URL that has been used to upload the track with the track metadata. 


## Dependencies

   * You will use **Go modules** for your dependencies and for your own module. 
   * For the development of the IGC processing, you will use an open source IGC library for Go: **[goigc](https://github.com/marni/goigc)**
   * You will use **MongoDB**. The recommended driver for MongoDB is the [official MongoDB Go driver](https://github.com/mongodb/mongo-go-driver) or the mgo successor [globalsign/mgo](https://github.com/globalsign/mgo). It is up to the student to decide which one to use. Motivate your choice. Each of them might be slightly different to the old mgo that we've used last year, however, the core concepts are the same and the usage patterns should feel familiar in all of them.


## Deployment

   * The core logic of the system will stay as in Assignment 1 on Heroku. 
   * The database (MongoDB) will be stored in mlab.com MongoLabs Sandbox.
   * The ticker end-point will be implemented as a Cloud Function [OPTIONAL] or as normal end-point within your Heroku deployment.
   * The "clock" trigger will be implemented in Go as independent executable deployed on OpenStack.

### The following will need to be added as environment variables for the program to run:

    PORT=<port>
    MONGO_HOST=<host>
    MONGO_USER=<user>
    MONGO_PASSWORD=<password>
    MONGO_DATABASE=<dbName>

## General rules

The project should be named **paragliding** and this should be the root of the URL. The server should respond with 404 when asked about rubbish URLs. The API should be mounted on the **api** path. All the REST verbs will be subsequently attached to the /paragliding/api/* root.

```
http://localhost:8080/paragliding/api            -> root URL, responds with meta info about API, see GET /api
http://localhost:8080/paragliding/               -> redirect to /paragliding/api
http://localhost:8080/paragliding/<rubbish>      -> 404
http://localhost:8080/<rubbish>                  -> 404
http://localhost:8080/paragliding/api/<rubbish>  -> 404
```

**Note:** the use of `http://localhost:8080` serves only a demonstration purposes. You will have your own URL from the provider, such as Heroku. `<rubbish>` represents any sequence of letters and digits that are not described in this specification.


### Track timestamps

Each track, ie. each IGC file uploaded to the system, will have a timestamp represented as a LONG number, that must be unique and monotonic. This can be achieved by storing a milisecond of the upload time to the server, although, you have to plan how to make it thread safe and scalable. This is relevant for the ticker API. Hint - you could use mongoDB IDs that are monotonic, but then you would have some security considerations - think which ones. 



### Ticker API

Your system allows multiple people to upload tracks. To facilitate sharing track information, the API allows people to search through the entire collection of tracks, by track ID, to obtain the details about a given track. To notify dependent systems about new tracks uploaded to the system, there is a ticker API. The purpose of the ticker API is to notify dependant applications (such as complex IGC visualisation webapps) about the new tracks being available. For example: imagine that another webapp needs to know what new tracks have been uploaded from the last sync that the app done with your system. They will keep track of the last timestamp of the last track that they know about. If they ask your API about the last timestamp, and it is different to the one they have, that means that your system has more tracks now that they know about. I.e. new tracks have been uploaded since they have queried the track information. So that they can request to get IDs of all the new tracks that the system now has. The ticker API provide this facility, of obtaining information about updates. It provides simple paging functionality. 


### Webhook API

Your system will allow subscribing a webhook such that it can notify subscribers about the events in your system. With the minimal system as ours, the really only exciting event that we have is adding (registering) new track via the `POST /api/track`. Thus, your system will notify all interested subscribers to the webhook API, with the notification about new track being added. For details, see the endpoints for registering and removing the webhook.



## Core API Specification

### GET /api

* What: meta information about the API
* Response type: application/json
* Response code: 200
* Body template

```
{
  "uptime": <uptime>
  "info": "Service for Paragliding tracks."
  "version": "v1"
}
```

* where: `<uptime>` is the current uptime of the service formatted according to [Duration format as specified by ISO 8601](https://en.wikipedia.org/wiki/ISO_8601#Durations). 




### POST /api/track

* What: track registration
* Response type: application/json
* Response code: 200 if everything is OK, appropriate error code otherwise, eg. when provided body content, is malformed or URL does not point to a proper IGC file, etc. Handle all errors gracefully. 
* Request body template

```
{
  "url": "<url>"
}
```

* Response body template

```
{
  "id": "<id>"
}
```

* where: `<url>` represents a normal URL, that would work in a browser, eg: `http://skypolaris.org/wp-content/uploads/IGS%20Files/Madrid%20to%20Jerez.igc` and `<id>` represents an ID of the track, according to your internal management system. You can choose what format <id> should be in your system. The only restriction is that it needs to be easily used in URLs and it must be unique. It is used in subsequent API calls to uniquely identify a track, see below.


### GET /api/track

* What: returns the array of all tracks ids
* Response type: application/json
* Response code: 200 if everything is OK, appropriate error code otherwise. 
* Response: the array of IDs, or an empty array if no tracks have been stored yet.

```
[<id1>, <id2>, ...]
```

### GET /api/track/`<id>`

* What: returns the meta information about a given track with the provided `<id>`, or NOT FOUND response code with an empty body.
* Response type: application/json
* Response code: 200 if everything is OK, appropriate error code otherwise. 
* Response: 

```
{
"H_date": <date from File Header, H-record>,
"pilot": <pilot>,
"glider": <glider>,
"glider_id": <glider_id>,
"track_length": <calculated total track length>
"track_src_url": <the original URL used to upload the track, ie. the URL used with POST>
}
```

### GET /api/track/`<id>`/`<field>`

* What: returns the single detailed meta information about a given track with the provided `<id>`, or NOT FOUND response code with an empty body. The response should always be a string, with the exception of the calculated track length, that should be a number.
* Response type: text/plain
* Response code: 200 if everything is OK, appropriate error code otherwise. 
* Response
   * `<pilot>` for `pilot`
   * `<glider>` for `glider`
   * `<glider_id>` for `glider_id`
   * `<calculated total track length>` for `track_length`
   * `<H_date>` for `H_date`
   * `<track_src_url>` for `track_src_url`



### GET /api/ticker/latest

* What: returns the timestamp of the latest added track
* Response type: text/plain
* Response code: 200 if everything is OK, appropriate error code otherwise. 
* Response: `<timestamp>` for the latest added track


### GET /api/ticker/

* What: returns the JSON struct representing the ticker for the IGC tracks. The first track returned should be the oldest. The array of track ids returned should be capped at 5, to emulate "paging" of the responses. The cap (5) should be a configuration parameter of the application (ie. easy to change by the administrator).
* Response type: application/json
* Response code: 200 if everything is OK, appropriate error code otherwise. 
* Response

```
{
"t_latest": <latest added timestamp>,
"t_start": <the first timestamp of the added track>, this will be the oldest track recorded
"t_stop": <the last timestamp of the added track>, this might equal to t_latest if there are no more tracks left
"tracks": [<id1>, <id2>, ...]
"processing": <time in ms of how long it took to process the request>
}
```

### GET /api/ticker/`<timestamp>`

* What: returns the JSON struct representing the ticker for the IGC tracks. The first returned track should have the timestamp HIGHER than the one provided in the query. The array of track IDs returned should be capped at 5, to emulate "paging" of the responses. The cap (5) should be a configuration parameter of the application (ie. easy to change by the administrator).
* Response type: application/json
* Response code: 200 if everything is OK, appropriate error code otherwise. 
* Response:

```
{
   "t_latest": <latest added timestamp of the entire collection>,
   "t_start": <the first timestamp of the added track>, this must be higher than the parameter provided in the query
   "t_stop": <the last timestamp of the added track>, this might equal to t_latest if there are no more tracks left
   "tracks": [<id1>, <id2>, ...]
   "processing": <time in ms of how long it took to process the request>
}
```



## Webhooks API

### POST /api/webhook/new_track/

* What: Registration of new webhook for notifications about tracks being added to the system. Returns the details about the registration. The `webhookURL` is required parameter of the request. The `minTriggerValue` is optional integer, that defaults to 1 if ommited. It indicated the frequency of updates - after how many new tracks the webhook should be called. 
* Response type: application/json
* Response code: 200 or 201 if everything is OK, appropriate error code otherwise. 
* **Request**

```
{
    "webhookURL": {
      "type": "string"
    },
    "minTriggerValue": {
      "type": "number"
    }
}
```

Example, that registers a webhook that should be trigger for every two new tracks added to the system. 

```
{
    "webhookURL": "http://remoteUrl:8080/randomWebhookPath",
    "minTriggerValue": 2, 
}
```

* **Response**

The response body should contain the id of the created resource (aka webhook registration), as string. Note, the response body will contain only the created id, as string, not the entire path; no json encoding. Response code upon success should be 200 or 201.


### Invoking a registered webhook

When invoking a registered webhook, use POST with the webhookURL and the following payload specification:
```
{
   "t_latest": <latest added timestamp of the entire collection>,
   "tracks": [<id1>, <id2>, ...]
   "processing": <time in ms of how long it took to process the request>
}
```

Note: the body should only include the NEW tracks ids. Not the entire collection.


### GET /api/webhook/new_track/`<webhook_id>`

* What: Accessing registered webhooks. Registered webhooks should be accessible using the GET method and the webhook id generated during registration.
* Response type: application/json
* Response code: 200 or 201 if everything is OK, appropriate error code otherwise. 
* **Response body**

```
{
    "webhookURL": {
      "type": "string"
    },
    "minTriggerValue": {
      "type": "number"
    }
}
```

### DELETE /api/webhook/new_track/`<webhook_id>`

* What: Deleting registered webhooks. Registered webhooks can further be deleted using the DELETE method and the webhook id.
* Response type: application/json
* Response code: 200 or 201 if everything is OK, appropriate error code otherwise. 
* Response body:

```
{
    "webhookURL": {
      "type": "string"
    },
    "minTriggerValue": {
      "type": "number"
    }
}
```


## Clock trigger

The idea behind the clock is to have a task that happens on regular basis without user interventions. In our case, you will implement a task, that checks every 10min if the number of tracks differs from the previous check, and if it does, it will notify a predefined Slack webhook. The actual webhook can be hardcoded in the system, or configured via some environmental variables - think which solution is better and why. 



## Admin API 

*Note*: The endpoints below should be either not exposed at all, or should be exposed to ADMIN users only. Best practice is to keep them in a completely different API root, prefixed with something unique, or keep the URL different to the publicly exposed API. Here, we are making it extremely simplistic exclusively for testing purposes.


### GET /admin/api/tracks_count

* What: returns the current count of all tracks in the DB
* Response type: text/plain
* Response code: 200 if everything is OK, appropriate error code otherwise. 
* Response: current count of the DB records


### DELETE /admin/api/tracks

* What: deletes all tracks in the DB
* Response type: text/plain
* Response code: 200 if everything is OK, appropriate error code otherwise. 
* Response: count of the DB records removed from DB



## Resources

* [Go IGC library](https://github.com/marni/goigc)
* successor of mgo driver: [globalsign/mgo](https://github.com/globalsign/mgo)
