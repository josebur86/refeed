package main

import (
    "encoding/json"
    "fmt"
    "net/http"

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
    router.HandleFunc("/feeds", apiHandler)

    http.ListenAndServe(":8080", router)
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
    response, err := json.Marshal(GlobalFeeds)
    if err != nil {
        fmt.Println("Error marshaling feeds: ", err)
    }

    w.Header().Set("Content-Type", "text/json; charset=utf-8")
    fmt.Fprintf(w, "%s", response)
}
