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
    "encoding/xml"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "strconv"
    "time"

    "github.com/gorilla/mux"
	_ "github.com/lib/pq"
)


type Feed struct {
    ID int `json:"id"`
    Title string `json:"title"`
    URL string `json:"url"` // TODO(joe): Use the URL type here?
}

type FeedEntry struct {
    ID int `json:"id"`
    Title string `json:"title"`
    URL string `json:"url"` // TODO(joe): Use the URL type here?
    FeedID int `json:"feedId"`
    Unread bool `json:"unread"`
}

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

    //OutputTestXML()

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
    rows, err := DB.Query("SELECT id, title, url FROM feeds WHERE id = $1;", id)
    if err != nil {
        log.Fatal("Error querying database: ", err)
    }
    defer rows.Close()

    if rows.Next() {
        var f Feed
        rows.Scan(&f.ID, &f.Title, &f.URL)

        resp, err := http.Get(f.URL)
        if err != nil {
            log.Fatal("Error fetching feed contents: ", err)
        }
        defer resp.Body.Close()

        contents, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            log.Fatal("Error consuming feed contents: ", err)
        }

        var atomFeed AtomFeed
        err = xml.Unmarshal(contents, &atomFeed)
        if err != nil {
            log.Fatal("Error parsing feed contents: ", err)
        }

        fmt.Fprintf(w, "%s\n", atomFeed.Title)
        fmt.Fprintf(w, "%s\n", atomFeed.Author.Name)
        fmt.Fprintf(w, "%s\n\n", atomFeed.Link.Href)
        for _, entry := range atomFeed.Entries {
            fmt.Fprintf(w, "  %s\n", entry.Title)
            fmt.Fprintf(w, "  %s\n", entry.Link.Href)
            fmt.Fprintf(w, "  %s\n\n", entry.Summary)
        }
    }
    if err = rows.Err(); err != nil {
        log.Printf("Error while iterating feed entries: ", err)
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

// TODO(joe): This section should be moved to where feeds are parsed and their entries are added to
// the unread list.

type AtomFeed struct {
    XMLName  xml.Name    `xml:"http://www.w3.org/2005/Atom feed"`
    Title    string      `xml:"title"`
    Link     Link        `xml:"link"`
    Updated  time.Time   `xml:"updated"`
    Author   Author      `xml:"author"`
    ID       string      `xml:"id"`
    Entries  []AtomEntry `xml:"entry"`
}
type AtomEntry struct {
    Title   string    `xml:"title"`
    Link    Link      `xml:"link"`
    ID      string    `xml:"id"`
    Updated time.Time `xml:"updated"`
    Summary string    `xml:"summary"`
}
type Link struct {
    Href string `xml:"href,attr"`
}
type Author struct {
    Name string `xml:"name"`
}
func OutputTestXML() {
    feed := AtomFeed {
        XMLName: xml.Name{"http://www.w3.org/2005/Atom", "feed"},
        Title: "Example Feed",
        Link: Link{ Href: "http://example.org/" },
        Updated: ParseTime("2003-12-13T18:30:02Z"),
        Author: Author{ Name: "John Doe" },
        ID: "urn:uuid:60a76c80-d399-11d9-b93C-0003939e0af6",
        Entries: []AtomEntry {
            {
                Title: "Atom-Powered Robots Run Amok",
                Link: Link { Href: "http://example.org/2003/12/13/atom03" },
                ID: "urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a",
                Updated: ParseTime("2003-12-13T18:30:02Z"),
                Summary: "Some Text.",
            },
        },
    }

    encoder := xml.NewEncoder(os.Stdout)
    encoder.Indent("  ", "    ")
    err := encoder.Encode(feed)
    if err != nil {
        log.Fatal("Unable to encode feed: ", err)
    }
    fmt.Printf("\n")
}
func ParseTime(s string) (time.Time) {
    time, err := time.Parse(time.RFC3339, s)
    if err != nil {
        log.Fatal("Unable to parse time: ", err)
    }

    return time
}
