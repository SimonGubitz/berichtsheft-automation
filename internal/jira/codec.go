package jira

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const jiraTimeLayout = "2006-01-02T15:04:05.000-0700"

type ActivityFeed struct {
	ID      string    `xml:"id"`
	Link    XMLLink   `xml:"link"`
	Title   string    `xml:"title"`
	Updated time.Time `xml:"updated"`
	Entries []Entry   `xml:"entry"`
}

type Entry struct {
	ID       string         `xml:"id"`
	Title    string         `xml:"title"`
	Activity ActivityObject `xml:"object"`
}

type ActivityObject struct {
	ID      string  `xml:"id"`
	Title   string  `xml:"title"`
	Summary string  `xml:"summary"`
	Link    XMLLink `xml:"link"`
	Type    string  `xml:"object-type"`
}

type XMLLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
}

type Issue struct {
	ID     string      `json:"id"`
	Key    string      `json:"key"`
	Fields IssueFields `json:"fields"`
}

type IssueFields struct {
	Summary     string          `json:"summary"`
	Description string          `json:"description"`
	Status      JiraStatus      `json:"status"`
	Assignee    JiraUser        `json:"assignee"`
	Creator     JiraUser        `json:"creator"`
	Reporter    JiraUser        `json:"reporter"`
	Comment     JiraCommentMeta `json:"comment"`
}

type JiraCommentMeta struct {
	MaxResults int           `json:"maxResults"`
	Total      int           `json:"total"`
	StartAt    int           `json:"startAt"`
	Comments   []JiraComment `json:"comments"`
}

type JiraComment struct {
	Self         string   `json:"self"`
	ID           string   `json:"id"`
	Author       JiraUser `json:"author"`
	UpdateAuthor JiraUser `json:"updateAuthor"`
	Body         string   `json:"body"`
	Created      JiraTime `json:"created"`
	Updated      JiraTime `json:"updated"`
}

type JiraStatus struct {
	Self        string `json:"self"`
	Description string `json:"description"`
	IconURL     string `json:"iconUrl"`
	Name        string `json:"name"`
	ID          string `json:"id"`
	// "statusCategory": {
	// 		"self": "https://jira.example.com/rest/api/2/statuscategory/2",
	// 		"id": 2,
	// 		"key": "new",
	// 		"colorName": "blue-gray",
	// 		"name": "Aufgaben"
	// }
}

type JiraUser struct {
	Self         string `json:"self"`
	Name         string `json:"name"`
	Key          string `json:"key"`
	EmailAddress string `json:"emailAddress"`
	// avatarURLs
	DiplayName string `json:"displayName"`
	Active     bool   `json:"active"`
	Timezone   string `json:"timeZone"`
	// Timezone time.Location
}

type JiraTime struct {
	time.Time
}

func (jt *JiraTime) UnmarshalJSON(data []byte) error {

	// Handle null values
	if string(data) == "null" {
		jt.Time = time.Time{}
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if strings.TrimSpace(s) == "" {
		jt.Time = time.Time{}
		return nil
	}
	t, err := time.Parse(jiraTimeLayout, s)
	if err != nil {
		return fmt.Errorf("failed to parse Jira time %q: %w", s, err)
	}
	jt.Time = t
	return nil

}

func (jt JiraTime) MarshalJSON() ([]byte, error) {

	if jt.Time.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(jt.Format(jiraTimeLayout))

}
