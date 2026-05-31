package ingest

import (
	"encoding/xml"
	"fmt"
	"strings"
)

type rssFeed struct {
	Channel struct {
		Items []rssItem `xml:"item"`
	} `xml:"channel"`
}

type rssItem struct {
	Title string `xml:"title"`
	Link  string `xml:"link"`
}

type atomFeed struct {
	Entries []struct {
		Title string `xml:"title"`
		Link  struct {
			Href string `xml:"href,attr"`
		} `xml:"link"`
	} `xml:"entry"`
}

type feedEntry struct {
	Title string
	Link  string
}

func parseFeedXML(data []byte) ([]feedEntry, error) {
	s := strings.TrimSpace(string(data))
	if strings.Contains(s, "<feed") {
		var af atomFeed
		if err := xml.Unmarshal(data, &af); err != nil {
			return nil, err
		}
		var out []feedEntry
		for _, e := range af.Entries {
			link := e.Link.Href
			if link == "" {
				continue
			}
			out = append(out, feedEntry{Title: strings.TrimSpace(e.Title), Link: link})
		}
		return out, nil
	}
	var rf rssFeed
	if err := xml.Unmarshal(data, &rf); err != nil {
		return nil, err
	}
	var out []feedEntry
	for _, it := range rf.Channel.Items {
		link := strings.TrimSpace(it.Link)
		if link == "" {
			continue
		}
		out = append(out, feedEntry{Title: strings.TrimSpace(it.Title), Link: link})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no feed entries")
	}
	return out, nil
}
