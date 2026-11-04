package jira

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"berichtsheft/internal/cache"
	"berichtsheft/internal/networking"
	"berichtsheft/internal/percentage"

	"github.com/snabb/isoweek"
)

type ResultAmount int

const (
	NoResultLimit = -1
)

type JiraClient struct {
	username   string
	password   string
	baseURL    string
	httpClient *http.Client
	reqCache   *cache.FileCache
	log        *slog.Logger
}

func NewJiraClient(baseURL, username, password string) (*JiraClient, error) {

	jiraHost := normalizeUserURL(baseURL)
	if err := validateURL(jiraHost); err != nil {
		return nil, fmt.Errorf("invalid user email. %w", err)
	}

	reqCache, err := cache.NewFileCache("jira_cache", 336*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("failed to create file cache. %w", err)
	}

	return &JiraClient{
		username:   username,
		password:   password,
		httpClient: http.DefaultClient,
		reqCache:   reqCache,
		baseURL:    jiraHost,
		log:        slog.Default().With("layer", "jira"),
	}, nil
}

func normalizeUserURL(s string) string {
	if !strings.HasPrefix(s, "https://") && !strings.HasPrefix(s, "http://") {
		return "https://" + s
	}
	return s
}

func validateURL(s string) error {

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

func (j *JiraClient) GetActivities(ctx context.Context, from, to time.Time, maxResults ResultAmount) (*ActivityFeed, bool, error) {

	header := make(http.Header)

	queryParams := make(url.Values)
	if maxResults != NoResultLimit {
		queryParams.Add("maxResults", strconv.Itoa(int(maxResults)))
	}
	queryParams.Add("streams", "user IS "+j.username)
	queryParams.Add("streams", fmt.Sprintf("update-date BETWEEN %d %d", from.UnixMilli(), to.UnixMilli()))

	reqUrl, err := url.JoinPath(j.baseURL, "/activity")
	if err != nil {
		return nil, false, fmt.Errorf("failed to construct request url. baseURL: %s, %w", j.baseURL, err)
	}

	reqUrl += "?" + queryParams.Encode()
	r, cached, err := networking.HTTPRequestWithCacheAndBasicAuth(ctx, j.httpClient, header, nil, j.reqCache, reqUrl, j.username, j.password)
	if err != nil {
		return nil, false, fmt.Errorf("failed to request activity feed. ULR: %s. %w", reqUrl, err)
	}

	var res ActivityFeed
	decoder := xml.NewDecoder(r)
	err = decoder.Decode(&res)
	if err != nil {
		return nil, cached, fmt.Errorf("failed to unmarshal activity feed. ULR: %s. %w", reqUrl, err)
	}

	slog.Debug("fetched activities", "activities", res)

	return &res, cached, nil
}

type WeekSummary struct {
	IsoWeek   int
	StartWeek time.Time
	EndWeek   time.Time
	Issues    []Issue
}

func (j *JiraClient) ConstructWeeks(ctx context.Context, activityFeed ActivityFeed) (w []WeekSummary, cachePercentage int, err error) {

	p := percentage.NewBoolPercentage()

	// key is athe iso week
	weekSummaries := make(map[int]WeekSummary)

	for _, entry := range activityFeed.Entries {

		regex := regexp.MustCompile(`^[a-zA-Z]+-\d{1,3}$`)
		issueKey := regex.FindString(entry.Activity.Title)
		if issueKey == "" {
			continue
		}

		issue, cached, err := j.GetIssue(ctx, issueKey)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to fetch issue. %w", err)
		}
		p.Inc(cached)

		var weekSummary WeekSummary
		y, w := issue.Fields.Updated.ISOWeek()
		startWeek := isoweek.StartTime(y, w, time.Local)
		endWeek := isoweek.StartTime(y, w, time.Local).Add(6 * 24 * time.Hour)
		j.log.Debug("isoweek", "date", issue.Fields.Updated, "week", w, "start", startWeek, "end", endWeek)

		if s, ok := weekSummaries[w]; ok {
			weekSummary = s
		} else {
			weekSummary = WeekSummary{
				IsoWeek:   w,
				StartWeek: startWeek,
				EndWeek:   endWeek,
				Issues:    make([]Issue, 0),
			}
		}
		weekSummary.Issues = append(weekSummary.Issues, *issue)

		weekSummaries[w] = weekSummary
	}

	res := make([]WeekSummary, len(weekSummaries))
	i := 0
	for _, v := range weekSummaries {
		res[i] = v
		i++
	}

	return res, p.Calculate(), nil
}

func (j *JiraClient) SearchIssues(ctx context.Context, jql string) ([]Issue, error) {

	header := make(http.Header)

	queryParams := make(url.Values)
	queryParams.Add("", "")

	reqUrl, err := url.JoinPath(j.baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to construct request url. baseURL: %s, %w", j.baseURL, err)
	}

	reqUrl += "?" + queryParams.Encode()
	r, _, err := networking.HTTPRequestWithCacheAndBasicAuth(ctx, j.httpClient, header, nil, j.reqCache, reqUrl, j.username, j.password)
	if err != nil {
		return nil, fmt.Errorf("failed to request search. %w", err)
	}

	var resp any
	decoder := json.NewDecoder(r)
	err = decoder.Decode(&resp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response. %w", err)
	}

	return nil, nil
}

func (j *JiraClient) GetIssue(ctx context.Context, key string) (*Issue, bool, error) {

	header := make(http.Header)
	reqUrl, err := url.JoinPath(j.baseURL, fmt.Sprintf("/rest/api/2/issue/%s", key))
	if err != nil {
		return nil, false, fmt.Errorf("failed to construct request url. baseURL: %s, %w", j.baseURL, err)
	}

	r, cached, err := networking.HTTPRequestWithCacheAndBasicAuth(ctx, j.httpClient, header, nil, j.reqCache, reqUrl, j.username, j.password)
	if err != nil {
		return nil, cached, fmt.Errorf("failed to do request to jira issue endpoint. %w", err)
	}

	var issue Issue
	decoder := json.NewDecoder(r)
	err = decoder.Decode(&issue)
	if err != nil {
		return nil, cached, fmt.Errorf("failed to decode issue. %w", err)
	}

	return &issue, cached, nil
}
