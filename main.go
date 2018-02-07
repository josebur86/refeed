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

var GlobalDB *sql.DB

func ConnectToDB() (*sql.DB) {
    db, err := sql.Open("postgres", getDatabaseConnectionString())
    if err != nil {
        log.Fatal(err)
    }

    // TODO(joe): Use QueryRow here?
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

    OutputTestXML()

    GlobalDB = ConnectToDB()

    router := mux.NewRouter() // TODO(joe): StrictSlash(true)??

    // GET
    router.HandleFunc("/feeds", AllFeedsHandler).Methods("GET");
    router.HandleFunc("/feeds/edit", EditFeedHandler).Methods("GET");
    router.HandleFunc("/feeds/{id}", SingleFeedHandler).Methods("GET");

    // PUT
    router.HandleFunc("/feeds", AddFeedHandler).Methods("PUT");
    router.HandleFunc("/feeds", AddFeedFromFormHandler).Methods("POST");


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

    rows, err := GlobalDB.Query("SELECT id, title, url from feeds where id = $1;", id)
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

        w.Header().Set("Content-Type", "text/xml; charset=utf-8")
        fmt.Fprintf(w, "%s", contents)
    }

    /*
    w.Header().Set("Content-Type", "text/json; charset=utf-8")
    if rows.Next() {
        var f Feed
        rows.Scan(&f.ID, &f.Title, &f.URL)
        json.NewEncoder(w).Encode(f)
    } else {
        fmt.Fprintf(w, "{}"); // TODO(joe): What's the more RESTy response to a request to a feed that doesn't exist.
    }
    */
}

func AddFeedHandler(w http.ResponseWriter, r *http.Request) {
    var f Feed
    err := json.NewDecoder(r.Body).Decode(&f)
    if err != nil {
        log.Fatal("Error parsing request body: ", err)
    }

    AddFeedToDatabase(f, w)
}

func AddFeedFromFormHandler(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()

    var f Feed
    f.Title = r.PostForm.Get("title")
    f.URL = r.PostForm.Get("url")

    log.Print(f)
    AddFeedToDatabase(f, w)
}

func AddFeedToDatabase(f Feed, w http.ResponseWriter) {
    log.Print(f)
    row := GlobalDB.QueryRow("INSERT INTO feeds (title, url) VALUES ($1, $2) RETURNING id;", f.Title, f.URL)

    var id int
    err := row.Scan(&id)
    if err != nil {
        log.Fatal("Error adding feed to the db: ", err)
    }

    rows, err := GlobalDB.Query("SELECT id, title, url from feeds where id = $1;", id)
    if err != nil {
        log.Fatal("Error querying database: ", err)
    }
    defer rows.Close()

    w.Header().Set("Content-Type", "text/json; charset=utf-8")
    if rows.Next() {
        var f Feed
        rows.Scan(&f.ID, &f.Title, &f.URL)
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(f)
    } else {
        fmt.Fprintf(w, "{}"); // TODO(joe): What's the more RESTy response to a request to a feed that doesn't exist.
    }
}

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

type AtomFeed struct {
    XMLName  xml.Name  `xml:"http://www.w3.org/2005/Atom feed"`
    Title    string    `xml:"title"`
    Link     Link      `xml:"link"`
    Updated  time.Time `xml:"updated"`
    Author   Author    `xml:"author"`
    ID       string    `xml:"id"`
    Entries  []Entry   `xml:"entry"`
}
type Entry struct {
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
    feed := AtomFeed{
        XMLName: xml.Name{"http://www.w3.org/2005/Atom", "feed"},
        Title: "Example Feed",
        Link: Link{ Href: "http://example.org/" },
        Updated: ParseTime("2003-12-13T18:30:02Z"),
        Author: Author{ Name: "John Doe" },
        ID: "urn:uuid:60a76c80-d399-11d9-b93C-0003939e0af6",
        Entries: []Entry {
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
