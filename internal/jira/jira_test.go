package jira_test

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"berichtsheft/internal/jira"
)

func NewJiraMockServer() *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /activity", func(w http.ResponseWriter, req *http.Request) {
		encoder := xml.NewEncoder(w)

		err := encoder.Encode(jira.ActivityFeed{
			ID:      "aaa-bbb-ccc",
			Link:    jira.XMLLink{Href: "test_href", Rel: "test_rel"},
			Title:   "activity-feed",
			Updated: time.Now(),
			Entries: []jira.Entry{
				{ID: "xxx-yyy-zzz", Title: "Added an attachment image", Activity: jira.ActivityObject{ID: "mmm-nnn-lll", Link: jira.XMLLink{Href: "test_href", Rel: "test_rel"}, Title: "ISSUE-987", Summary: "Test Issue Summary", Type: "http://streams.atlassian.com/syndication/types/issue"}},
				{ID: "xyz-xyz-xyz", Title: "Added comment to Test Issue", Activity: jira.ActivityObject{ID: "abc-abc-abc", Link: jira.XMLLink{Href: "test_href", Rel: "test_rel"}, Title: "ISSUE-123", Summary: "Test Issue Summary", Type: "http://streams.atlassian.com/syndication/types/issue"}},
			},
		})

		if err != nil {
			fmt.Printf("error: %+v", err)
		}
	})

	mux.HandleFunc("GET /rest/api/2/issue/{key}", func(w http.ResponseWriter, req *http.Request) {
		key := req.PathValue("key")

		encoder := json.NewEncoder(w)
		_ = encoder.Encode(jira.Issue{
			ID:  "aaa-bbb-ccc",
			Key: key,
			Fields: jira.IssueFields{
				Summary:     "Test Summary",
				Description: "Test Description",
				Updated:     jira.JiraTime{time.Now()},
			},
		})
	})

	return httptest.NewServer(mux)
}

func TestJiraUnit(t *testing.T) {

	ctx := context.Background()
	server := NewJiraMockServer()

	j, err := jira.NewJiraClient(server.URL, "test_username", "password123")
	if err != nil {
		t.Fatalf("failed to instantiate new jira client. %+v", err)
	}

	activities, _, err := j.GetActivities(ctx, 1)
	if err != nil {
		t.Fatalf("failed to get activities. %+v", err)
	}

	_, _, err = j.ConstructWeeks(ctx, *activities)
	if err != nil {
		t.Fatalf("failed to construct weeks. %+v", err)
	}
}
