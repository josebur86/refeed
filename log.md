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

Polygon uses an Atom feed format so I'll work on parsing that format first.

