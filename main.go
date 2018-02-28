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

var DB *sql.DB

func ConnectToDB() (*sql.DB) {
    db, err := sql.Open("postgres", getDatabaseConnectionString())
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Database connected!")

    return db
}

func main() {
    DB = ConnectToDB()

    router := mux.NewRouter() // TODO(joe): StrictSlash(true)??

    // GET
    router.HandleFunc("/feeds", AllFeedsHandler).Methods("GET");
    router.HandleFunc("/feeds/edit", EditFeedHandler).Methods("GET");
    router.HandleFunc("/feeds/{id}", SingleFeedHandler).Methods("GET");

    // PUT
    router.HandleFunc("/feeds", AddFeedHandler).Methods("PUT");
    router.HandleFunc("/feeds", AddFeedFromFormHandler).Methods("POST");

    // DELETE
    router.HandleFunc("/feeds/{id}", DeleteFeedHandler).Methods("DELETE");

    log.Printf("Listening on port 8080")
    http.ListenAndServe(":8080", router)
}

// TODO(joe): For now, I'm going to assume that the db schema exists and fail or return now results
// if the schema do not exist. I'll need to either check and create or use a front end for this.
// STUDY(joe): Is there a rails like library for go?


func AllFeedsHandler(w http.ResponseWriter, r *http.Request) {
    rows, err := DB.Query("SELECT id, title, url from feeds;")
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
}

func GetFeedIDFromRequest(r *http.Request) int {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        log.Fatal("Error parsing feed id: ", err)
    }

    return id
}

func SingleFeedHandler(w http.ResponseWriter, r *http.Request) {
    id := GetFeedIDFromRequest(r)

    var f Feed
    err := DB.QueryRow("SELECT id, title, url FROM feeds WHERE id = $1;", id).Scan(&f.ID, &f.Title, &f.URL)
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
    }
}

func AddFeedHandler(w http.ResponseWriter, r *http.Request) {
    var f Feed
    err := json.NewDecoder(r.Body).Decode(&f)
    if err != nil {
        log.Fatal("Error parsing request body: ", err)
    }

    AddFeedToDatabase(f, w)
}

func DeleteFeedHandler(w http.ResponseWriter, r *http.Request) {
    id := GetFeedIDFromRequest(r)

    res, err := DB.Exec("DELETE FROM feeds WHERE id = $1;", id)
    if err != nil {
        log.Fatal(err)
    }

    rowCount, err := res.RowsAffected()
    if err != nil {
        log.Fatal(err)
    }

    log.Println("Deleted row count: ", rowCount)
    w.WriteHeader(http.StatusNoContent)
}

func AddFeedToDatabase(f Feed, w http.ResponseWriter) {
    log.Print(f)

    var id int
    err := DB.QueryRow("INSERT INTO feeds (title, url) VALUES ($1, $2) RETURNING id;", f.Title, f.URL).Scan(&id)
    if err != nil {
        log.Fatal("Error adding feed to the db: ", err)
    }

    err = DB.QueryRow("SELECT id, title, url from feeds where id = $1 limit 1;", id).Scan(&f.ID, &f.Title, &f.URL)
    if err != nil {
        log.Fatal("Error querying database: ", err)
    }

    w.Header().Set("Content-Type", "text/json; charset=utf-8")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(f)
}

