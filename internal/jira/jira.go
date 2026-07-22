package jira

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"berichtsheft/internal/cache"
	"berichtsheft/internal/networking"
)

type JiraClient struct {
	username   string
	password   string
	baseURL    string
	httpClient *http.Client
	myCache    cache.FileCache
	log        *slog.Logger
}

func NewJiraClient(baseURL string, username, password string) (*JiraClient, error) {

	myCache, err := cache.NewFileCache("jira_cache")
	if err != nil {
		return nil, fmt.Errorf("failed to create file cache. %w", err)
	}

	return &JiraClient{
		username:   username,
		password:   password,
		httpClient: http.DefaultClient,
		myCache:    *myCache,
		baseURL:    baseURL,
		log:        slog.Default().With("layer", "GetActivities"),
	}, nil
}

func (j *JiraClient) GetActivities(ctx context.Context, maxResults int) (*ActivityFeed, error) {

	queryParams := make(url.Values)
	queryParams.Add("maxResults", strconv.Itoa(maxResults))
	queryParams.Add("streams", "user IS "+j.username)

	header := make(http.Header)
	reqUrl, err := url.JoinPath(j.baseURL, "/activity")
	if err != nil {
		return nil, fmt.Errorf("failed to construct request url. %w", err)
	}

	reqUrl += "?" + queryParams.Encode()
	r, err := networking.HTTPRequestWithCacheAndBasicAuth(ctx, j.httpClient, header, j.myCache, reqUrl, j.username, j.password)
	if err != nil {
		return nil, fmt.Errorf("failed to request activity feed. %w", err)
	}

	var res ActivityFeed
	decoder := xml.NewDecoder(r)
	err = decoder.Decode(&res)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal activity feed. %w", err)
	}

	slog.Debug("result", "object", res)

	return &res, nil
}

func (j *JiraClient) GetIssue(ctx context.Context, key string) (*Issue, error) {

	header := make(http.Header)
	reqUrl, err := url.JoinPath(j.baseURL, "/rest/api/2/issue/%s", key)
	if err != nil {
		return nil, fmt.Errorf("failed to construct request url. %w", err)
	}

	r, err := networking.HTTPRequestWithCacheAndBasicAuth(ctx, j.httpClient, header, j.myCache, reqUrl, j.username, j.password)
	if err != nil {
		return nil, fmt.Errorf("failed to do request to jira issue endpoint. %w", err)
	}

	var issue Issue
	decoder := json.NewDecoder(r)
	err = decoder.Decode(&issue)
	if err != nil {
		return nil, fmt.Errorf("failed to decode issue. %w", err)
	}

	return &issue, nil
}
