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
    "io/ioutil"
    "log"
    "net/http"
    "strconv"

    "github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func ConnectToDB() (*sql.DB) {
    db, err := sql.Open("postgres", getDatabaseConnectionString())
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Database connected!")

    return db
}

func main() {
    db := ConnectToDB()

    router := mux.NewRouter() // TODO(joe): StrictSlash(true)??

    // GET
    router.Handle("/feeds", getAllFeedsHandler(db)).Methods("GET");
    router.HandleFunc("/feeds/edit", editFeedHandler).Methods("GET");
    router.Handle("/feeds/{id}", getSingleFeedHandler(db)).Methods("GET");

    // PUT
    router.Handle("/feeds", getAddFeedHandler(db)).Methods("PUT");
    router.Handle("/feeds", getAddFeedFromFormHandler(db)).Methods("POST");

    // DELETE
    router.Handle("/feeds/{id}", getDeleteFeedHandler(db)).Methods("DELETE");

    log.Printf("Listening on port 8080")
    http.ListenAndServe(":8080", router)
}

// TODO(joe): For now, I'm going to assume that the db schema exists and fail or return now results
// if the schema do not exist. I'll need to either check and create or use a front end for this.
// STUDY(joe): Is there a rails like library for go?


func getAllFeedsHandler(db *sql.DB) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        rows, err := db.Query("SELECT id, title, url from feeds;")
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
        if err = rows.Err(); err != nil {
            log.Printf("Error while iterating feeds: ", err)
        }

        response, err := json.Marshal(feeds)
        if err != nil {
            log.Fatal("Error marshaling feeds: ", err)
        }

        w.Header().Set("Content-Type", "text/json; charset=utf-8")
        fmt.Fprintf(w, "%s", response)
    })
}

func getSingleFeedHandler(db *sql.DB) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        id := getFeedIDFromRequest(r)

        var f Feed
        err := db.QueryRow("SELECT id, title, url FROM feeds WHERE id = $1;", id).Scan(&f.ID, &f.Title, &f.URL)
        if err != nil {
            log.Fatal("Error querying database: ", err)
        }

        resp, err := http.Get(f.URL)
        if err != nil {
            log.Fatal("Error fetching feed contents: ", err)
        }
        defer resp.Body.Close()

        contents, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            log.Fatal("Error consuming feed contents: ", err)
        }

        entries, err := f.ParseEntries(contents)
        if err != nil {
            log.Fatal("Error parsing feed contents: ", err)
        }

        fmt.Fprintf(w, "%s\n", f.Title)
        fmt.Fprintf(w, "%s\n\n", f.URL)
        for _, entry := range entries {
            fmt.Fprintf(w, "  %s\n", entry.Title)
            fmt.Fprintf(w, "  %s\n", entry.URL)

            entry.Save(db)
        }
    })
}

func getAddFeedHandler(db *sql.DB) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        var f Feed
        err := json.NewDecoder(r.Body).Decode(&f)
        if err != nil {
            log.Fatal("Error parsing request body: ", err)
        }

        addFeedToDatabase(f, db)

        w.Header().Set("Content-Type", "text/json; charset=utf-8")
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(f)
    })
}

func getDeleteFeedHandler(db *sql.DB) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        id := getFeedIDFromRequest(r)

        res, err := db.Exec("DELETE FROM feeds WHERE id = $1;", id)
        if err != nil {
            log.Fatal(err)
        }

        rowCount, err := res.RowsAffected()
        if err != nil {
            log.Fatal(err)
        }

        log.Println("Deleted row count: ", rowCount)
        w.WriteHeader(http.StatusNoContent)
    })
}

func getFeedIDFromRequest(r *http.Request) int {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        log.Fatal("Error parsing feed id: ", err)
    }

    return id
}

func addFeedToDatabase(f Feed, db *sql.DB) {
    var id int
    err := db.QueryRow("INSERT INTO feeds (title, url) VALUES ($1, $2) RETURNING id;", f.Title, f.URL).Scan(&id)
    if err != nil {
        log.Fatal("Error adding feed to the db: ", err)
    }

    err = db.QueryRow("SELECT id, title, url from feeds where id = $1 limit 1;", id).Scan(&f.ID, &f.Title, &f.URL)
    if err != nil {
        log.Fatal("Error querying database: ", err)
    }

}

