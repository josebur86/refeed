package main

// TODO(joe): Do some real unit testing.
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
