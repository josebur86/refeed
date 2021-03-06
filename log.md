05 Feb 2018
===========
### 20:27 ###
Heading back into this codebase because Feedly starting showing me ads that were mixed in with my
actual feed stories. I get why they have to but it annoys me enough that I think I could make
something basic that could be a good enough replacement.

So far this app is a simple CRUD app to manage RSS feeds. It does not parse RSS feeds or provide
the article text.

The current API is:

```
GET /feeds - returns all the feeds that we know about.
GET /feeds/{id} - return the feed with the id

PUT /feeds - Add a feed
```
OK I think the next thing that I want to do is make simple edit page that you get when the user goes
to /feeds/edit. It will have the fields needed for adding a new feed.

### 21:39 ###
OK now I want to parse and display the article titles.

### 21:54 ###
Well that was simplier than I thought it would be. I have the server returning the feed XML when the
URL is a valid one and exiting when the URL is not valid. Firefox does something cool and displays
the feed contents in a pretty way.

Polygon uses an Atom feed format so I'll work on parsing that format first. I seem to remember that
using Go's XML parsing can be kind of tricky and the easiest way to know what the parser was looking
for is to have it generate XML that is in the same structure as the XML you are trying to parse. The
Atom standard is RFC 4287 and has a simple one entry example that I will try to output first.

06 Feb 2018
===========
### 08:05 ###
I remembered right. Getting the XML just right is tricky and time consuming. Here's what I'm
outputing right now.

```
<feed xmlns="http://www.w3.org/2005/Atom">
     <title>Example Feed</title>
     <href xmlns="link,">http://example.org/</href>
</feed>
```

The link is all messed up. Off to work!

07 Feb 2018
===========
### 07:30 ###
I need to pay more attention to spacing when working with the XML tags on struct elements.

```
type Link struct {
    Href string `xml:"href, attr"`
}
```

is not the same as

```
type Link struct {
    Href string `xml:"href,attr"` // Note the lack of space between the comma and attr.
}
```

### 08:05 ###
Now I"m starting to get the hang of it. Current output is

```
<feed xmlns="http://www.w3.org/2005/Atom">
    <title>Example Feed</title>
    <link href="http://example.org/"></link>
    <updated>2003-12-13T18:30:02Z</updated>
    <author>
        <name>John Doe</name>
    </author>
    <id>urn:uuid:60a76c80-d399-11d9-b93C-0003939e0af6</id>
    <entry>
        <title>Atom-Powered Robots Run Amok</title>
        <link href="http://example.org/2003/12/13/atom03"></link>
        <id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
        <updated>2003-12-13T18:30:02Z</updated>
        <summary>Some Text.</summary>
    </entry>
</feed>
```

This is all of the example. Next, I want to see how this bare-bones version handles a real Atom
feed. But first, off to work!

08 Feb 2018
===========
### 08:08 ###
The current AtomFeed definition works with Polygon's feed. Now the single feed handler just returns
a bunch of text with basic feed and entry info.

