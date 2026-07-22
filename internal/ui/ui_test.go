package ui_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// start the mock server
func NewMockServer() *httptest.Server {

	server := http.NewServeMux()
	server.HandleFunc("/rest/api/2/issue/:key", func(w http.ResponseWriter, req *http.Request) {

	})

	return httptest.NewServer(server)
}

func Test(t *testing.T) {

	s, err := url.JoinPath("jira.example.com", "rest/api/2")
	if err != nil {
		t.Fatalf("%+v", err)
	}

	s = NormalizeUserURL(s)
}

func NormalizeUserURL(s string) string {
	return ""
}

func Validate(s string) error {

	u, err := url.ParseRequestURI(s)
	if err != nil {
		return fmt.Errorf("invalid URL")
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("URL must start with http:// or https://")
	}

	if u.Host == "" {
		return fmt.Errorf("missing host")
	}

	return nil
}
