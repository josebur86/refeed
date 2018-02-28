package main

import (
    "time"
    "database/sql"
    "encoding/xml"
)

type Feed struct {
    ID int        `json:"id"`
    Title string  `json:"title"`
    URL string    `json:"url"` // TODO(joe): Use the URL type here?
}

func (f *Feed) ParseEntries(contents []byte) ([]FeedEntry, error) {
    // TODO(joe): We can only parse Atom Feeds. Implement RSS as well.

    var atomFeed AtomFeed
    if err := xml.Unmarshal(contents, &atomFeed); err != nil {
        return nil, err
    }

    var entries []FeedEntry
    for _, entry := range atomFeed.Entries {
        entries = append(entries, FeedEntry {
            ID: -1,
            Title: entry.Title,
            URL: entry.Link.Href,
            FeedID: f.ID,
            Unread: true,
        })
    }

    return entries, nil
}

type FeedEntry struct {
    ID int        `json:"id"`
    Title string  `json:"title"`
    URL string    `json:"url"` // TODO(joe): Use the URL type here?
    FeedID int    `json:"feedId"`
    Unread bool   `json:"unread"`
}

func (e *FeedEntry) Save(db *sql.DB) error {
    var id int
    err := db.QueryRow("INSERT INTO entries (title, url, feed_id, unread) VALUES ($1, $2, $3, $4) RETURNING id;",
        e.Title, e.URL, e.FeedID, e.Unread).Scan(&id)
    if err != nil {
        // TODO(joe): Wrap in my own error?
        return err
    }

    e.ID = id

    return nil
}

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
