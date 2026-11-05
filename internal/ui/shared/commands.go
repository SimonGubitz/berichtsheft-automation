package shared

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"berichtsheft/internal/jira"
	"berichtsheft/internal/ollama"

	tea "charm.land/bubbletea/v2"
)

type ErrMsg error

type JiraActivityMsg struct {
	Activities *jira.ActivityFeed
	Cached     int
}
type JiraWeeksMsg struct {
	Input  string
	Cached int
}
type OllamaRespMsg struct {
	Response string
}

type State string

const (
	StateLoadingJiraAct  = "jiraact"
	StateLoadingJiraWeek = "jiraweek"
	StateLoadingLLM      = "ollama"
	StateFinished        = "finished"
)

func GetJiraActivityCmd(ctx context.Context, j *jira.JiraClient, from, to time.Time) tea.Cmd {
	return func() tea.Msg {
		activities, cached, err := getJiraInput(ctx, j, from, to)
		if err != nil {
			return ErrMsg(err)
		}

		cachePerc := 0
		if cached {
			cachePerc = 100
		}
		return JiraActivityMsg{Activities: activities, Cached: cachePerc}
	}
}

func getJiraInput(ctx context.Context, j *jira.JiraClient, from, to time.Time) (*jira.ActivityFeed, bool, error) {
	return j.GetActivities(ctx, from, to, jira.NoResultLimit)
}

func GetJiraWeekCmd(ctx context.Context, j *jira.JiraClient, activities *jira.ActivityFeed) tea.Cmd {
	return func() tea.Msg {

		input, cachePerc, err := constructJiraWeeks(ctx, j, activities)
		if err != nil {
			return ErrMsg(err)
		}

		return JiraWeeksMsg{Input: input, Cached: cachePerc}
	}
}

type LLMWeek struct {
	ISOWeek   int       `json:"week"`
	StartWeek time.Time `json:"startOfWeek"`
	EndWeek   time.Time `json:"endOfWeek"`
	Tasks     []string  `json:"tasks"`
}

func constructJiraWeeks(ctx context.Context, j *jira.JiraClient, activities *jira.ActivityFeed) (llmInput string, cachePercentage int, err error) {

	weeks, cached, err := j.ConstructWeeks(ctx, *activities)
	if err != nil {
		return "", cached, fmt.Errorf("failed to construct weeks. %w", err)
	}

	llmWeeks := make([]LLMWeek, 0, len(weeks))
	for _, week := range weeks {
		seen := make(map[string]struct{})
		tasks := make([]string, 0)

		for _, issue := range week.Issues {
			// Only deduplicate within this week
			if _, ok := seen[issue.ID]; ok {
				continue
			}
			seen[issue.ID] = struct{}{}

			summary := strings.TrimSpace(issue.Fields.Summary)
			if summary == "" {
				continue
			}

			tasks = append(tasks, summary)
		}

		if len(tasks) == 0 {
			continue
		}

		llmWeeks = append(llmWeeks, LLMWeek{
			ISOWeek:   week.IsoWeek,
			StartWeek: week.StartWeek,
			EndWeek:   week.EndWeek,
			Tasks:     tasks,
		})
	}

	inputBytes, err := json.MarshalIndent(llmWeeks, "", "  ")
	if err != nil {
		return "", cached, fmt.Errorf("failed to marshal model data. %w", err)
	}
	return string(inputBytes), cached, nil
}

func GetLLMSummaryCmd(ctx context.Context, ollamaHost, input string) tea.Cmd {
	return func() tea.Msg {
		resp, err := getLLMSummary(ctx, ollamaHost, input)
		if err != nil {
			return ErrMsg(err)
		}

		return OllamaRespMsg{Response: resp}
	}
}

func getLLMSummary(ctx context.Context, ollamaHost, input string) (string, error) {
	o := ollama.NewOllamaClient(ollamaHost)
	hasModel, err := o.HasModel(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list active models. %w", err)
	}

	if !hasModel {
		err := o.PullModel(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to pull model. %w", err)
		}
	}

	resp, err := o.GetResponse(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to generate response. %w", err)
	}

	return resp, nil
}
