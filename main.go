package main

/*
    TEST JSON
    {
	    "title": "Polygon",
	    "url": "https://www.polygon.com/rss/index.xml"
    }
*/

import (
	"database/sql"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "strconv"

    "github.com/gorilla/mux"
	_ "github.com/lib/pq"
)


type Feed struct {
    ID int `json:"id"`
    Title string `json:"title"`
    URL string `json:"url"` // TODO(joe): Use the URL type here?
}

var GlobalFeeds []Feed
var GlobalDB *sql.DB

func ConnectToDB() (*sql.DB) {
    db, err := sql.Open("postgres", getDatabaseConnectionString())
    if err != nil {
        log.Fatal(err)
    }

    rows, err := db.Query("SELECT version();")
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()

    if rows.Next() {
        var version string;
        rows.Scan(&version)
        log.Printf("Connected: %s", version)
    }

    return db
}

func main() {
    GlobalDB = ConnectToDB()

    GlobalFeeds = append(GlobalFeeds, Feed{1, "Example", "http://www.Example.com/rss"})
    GlobalFeeds = append(GlobalFeeds, Feed{2, "WhoCares", "http://www.WhoCares.com/rss"})
    GlobalFeeds = append(GlobalFeeds, Feed{3, "Charlie.com", "http://www.Charlie.com/rss"})

    router := mux.NewRouter() // TODO(joe): StrictSlash(true)??

    // GET
    router.HandleFunc("/feeds", AllFeedsHandler).Methods("GET");
    router.HandleFunc("/feeds/{id}", SingleFeedHandler).Methods("GET");

    // PUT
    router.HandleFunc("/feeds", AddFeedHandler).Methods("PUT");

    http.ListenAndServe(":8080", router)
}

// TODO(joe): For now, I'm going to assume that the db schema exists and fail or return now results
// if the schema do not exist. I'll need to either check and create or use a front end for this.
// STUDY(joe): Is there a rails like library for go?


func AllFeedsHandler(w http.ResponseWriter, r *http.Request) {
    rows, err := GlobalDB.Query("SELECT id, title, url from feeds;")
    if err != nil {
        log.Fatal("Error querying database: ", err)
    }
    defer rows.Close()

    feeds := []Feed{}
    for rows.Next() {
        var f Feed
        rows.Scan(&f.ID, &f.Title, &f.URL)
        feeds = append(feeds, f)
    }

    response, err := json.Marshal(feeds)
    if err != nil {
        log.Fatal("Error marshaling feeds: ", err)
    }

    w.Header().Set("Content-Type", "text/json; charset=utf-8")
    fmt.Fprintf(w, "%s", response)
}

func SingleFeedHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        log.Fatal("Error parsing feed id: ", err)
    }

    w.Header().Set("Content-Type", "text/json; charset=utf-8")
    for _, feed := range GlobalFeeds {
        if feed.ID == id {
            json.NewEncoder(w).Encode(feed)
            return
        }
    }

    fmt.Fprintf(w, "{}"); // TODO(joe): What's the more RESTy response to a request to a feed that doesn't exist.
}

func AddFeedHandler(w http.ResponseWriter, r *http.Request) {
    // Here's what I think I have to do:
    //  1. Get the data if there is any
    //  2. Assume the data is json and parse it.
    //  3. Add the feed to the GlobalFeeds list.
    //  4. Return the correct response.
    var feed Feed
    err := json.NewDecoder(r.Body).Decode(&feed)
    if err != nil {
        log.Fatal("Error parsing request body: ", err)
    }

    id := len(GlobalFeeds)+1
    feed.ID = id;

    GlobalFeeds = append(GlobalFeeds, feed)

    w.Header().Set("Content-Type", "text/json; charset=utf-8")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(feed)
}
