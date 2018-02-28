package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
)

// TODO(joe): If I ever end up getting around to making a frontend for this thing, this type of
// thing will need to go there.
func EditFeedHandler(w http.ResponseWriter, r *http.Request) {
    // TODO(joe): We should really use an html text template here.
    fmt.Fprintf(w, "%s",
    `<html>
         <body>
            <form action="/feeds" method="post">
                Title: <input type="text" name="title"><br>
                URL: <input type="text" name="url"><br>
                <input type="submit" value="Add Feed">
            </form>
         </body>
     </html>`)
}

func GetAddFeedFromFormHandler(db *sql.DB) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        r.ParseForm()

        var f Feed
        f.Title = r.PostForm.Get("title")
        f.URL = r.PostForm.Get("url")

        log.Print(f)
        // FIXME(joe): Instead of directly calling the database, this should send a request through the
        // REST API.
        AddFeedToDatabase(f, db)

        w.Header().Set("Content-Type", "text/json; charset=utf-8")
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(f)
    })
}
