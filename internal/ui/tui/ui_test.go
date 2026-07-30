package tui_test

import (
	"berichtsheft/internal/jira"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func NewMockServer() *httptest.Server {

	server := http.NewServeMux()
	server.HandleFunc("/rest/api/2/issue/{key}", func(w http.ResponseWriter, req *http.Request) {
		key := req.PathValue("key")

		_ = json.NewEncoder(w).Encode(jira.Issue{
			ID:  "uuid",
			Key: key,
			Fields: jira.IssueFields{
				Summary:     "Test Summary",
				Description: "Test Description",
			},
		})
	})

	return httptest.NewServer(server)
}

func TestUI(t *testing.T) {

	s, err := url.JoinPath("", "rest/api/2/issue/mykey")
	if err != nil {
		t.Fatalf("%+v", err)
	}

	s = NormalizeUserURL(s)
	t.Log(s)
	t.Log(ValidateURL(s))
}

func NormalizeUserURL(s string) string {
	if !strings.HasPrefix(s, "https://") && !strings.HasPrefix(s, "http://") {
		return "https://" + s
	}
	return s
}

func ValidateURL(s string) error {

	u, err := url.ParseRequestURI(s)
	if err != nil {
		return fmt.Errorf("invalid URL. %w", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("URL must start with http:// or https://")
	}

	if u.Host == "" {
		return fmt.Errorf("missing host")
	}

	return nil
}
