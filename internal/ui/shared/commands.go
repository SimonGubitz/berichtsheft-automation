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

type InputMsg struct {
	Input string
}
type RespMsg struct {
	Response string
}

type State string

const (
	StateLoadingJira = "jira"
	StateLoadingLLM  = "ollama"
	StateFinished    = "finished"
)

func GetJiraInputCmd(ctx context.Context, jiraUrl, username, password string) tea.Cmd {
	return func() tea.Msg {
		input, err := getJiraInput(ctx, jiraUrl, username, password)
		if err != nil {
			return ErrMsg(err)
		}

		return InputMsg{Input: input}
	}
}

type LLMWeek struct {
	ISOWeek   int       `json:"week"`
	StartWeek time.Time `json:"startOfWeek"`
	EndWeek   time.Time `json:"endOfWeek"`
	Tasks     []string  `json:"tasks"`
}

func getJiraInput(ctx context.Context, jiraUrl, username, password string) (string, error) {
	j, err := jira.NewJiraClient(jiraUrl, username, password)
	if err != nil {
		return "", fmt.Errorf("failed to instantiate new jira client. %w", err)
	}

	activities, err := j.GetActivities(ctx, 150)
	if err != nil {
		return "", fmt.Errorf("failed to get activities. %w", err)
	}

	weeks, err := j.ConstructWeeks(ctx, *activities)
	if err != nil {
		return "", fmt.Errorf("failed to construct weeks. %w", err)
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
		return "", fmt.Errorf("failed to marshal model data. %w", err)
	}
	return string(inputBytes), nil
}

func GetLLMSummaryCmd(ctx context.Context, ollama_host, input string) tea.Cmd {
	return func() tea.Msg {
		resp, err := getLLMSummary(ctx, ollama_host, input)
		if err != nil {
			return ErrMsg(err)
		}

		return RespMsg{Response: resp}
	}
}

func getLLMSummary(ctx context.Context, ollama_host, input string) (string, error) {
	o := ollama.NewOllamaClient(ollama_host)
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
