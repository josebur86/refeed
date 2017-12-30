package main

import (
    "encoding/json"
    "fmt"
    "net/http"
)

type Feed struct {
    ID int `json:"id"`
    Title string `json:"title"`
    URL string `json:"url"`
}

var GlobalFeeds []Feed

func main() {

    GlobalFeeds = append(GlobalFeeds, Feed{1, "Example", "http://www.Example.com/rss"})
    GlobalFeeds = append(GlobalFeeds, Feed{2, "WhoCares", "http://www.WhoCares.com/rss"})
    GlobalFeeds = append(GlobalFeeds, Feed{3, "Charlie.com", "http://www.Charlie.com/rss"})

    http.HandleFunc("/feeds", func(w http.ResponseWriter, r *http.Request) {
        response, err := json.Marshal(GlobalFeeds)
        if err != nil {
            fmt.Println("Error marshaling feeds: ", err)
        }

        w.Header().Set("Content-Type", "text/json")
        fmt.Fprintf(w, "%s", response)
    })

    http.ListenAndServe(":8080", nil)
}
