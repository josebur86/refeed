package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"

    "github.com/gorilla/mux"
)

type Feed struct {
    ID int `json:"id"`
    Title string `json:"title"`
    URL string `json:"url"` // TODO(joe): Use the URL type here?
}

var GlobalFeeds []Feed

func main() {

    GlobalFeeds = append(GlobalFeeds, Feed{1, "Example", "http://www.Example.com/rss"})
    GlobalFeeds = append(GlobalFeeds, Feed{2, "WhoCares", "http://www.WhoCares.com/rss"})
    GlobalFeeds = append(GlobalFeeds, Feed{3, "Charlie.com", "http://www.Charlie.com/rss"})

    router := mux.NewRouter() // TODO(joe): StrictSlash(true)??
    router.HandleFunc("/feeds", AllFeedsHandler)
    router.HandleFunc("/feeds/{id}", SingleFeedHandler)

    http.ListenAndServe(":8080", router)
}

func AllFeedsHandler(w http.ResponseWriter, r *http.Request) {
    response, err := json.Marshal(GlobalFeeds)
    if err != nil {
        fmt.Println("Error marshaling feeds: ", err)
        return
    }

    w.Header().Set("Content-Type", "text/json; charset=utf-8")
    fmt.Fprintf(w, "%s", response)
}

func SingleFeedHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        fmt.Println("Error parsing feed id: ", err)
        return
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
