package main

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"berichtsheft/internal/jira"
	"berichtsheft/internal/ollama"

	"charm.land/lipgloss/v2"
)

type HashSet map[string]struct{}

func getenv(key string) string {

	style := lipgloss.NewStyle().Foreground(lipgloss.Red)
	res := os.Getenv(key)
	if res == "" {
		fmt.Println(style.Render(fmt.Sprintf("environment variable `%s` not set.", key)))
		os.Exit(1)
	}

	return res
}

func main() {

	ctx := context.Background()

	jiraUrl := getenv("JIRA_URL")
	username := getenv("JIRA_USERNAME")
	password := getenv("JIRA_PASSWORD")
	ollama_host := getenv("OLLAMA_HOST")

	jira, err := jira.NewJiraClient(jiraUrl, username, password)
	if err != nil {
		fmt.Printf("failed to instantiate new jira client. %+v\n", err)
		return
	}

	activities, err := jira.GetActivities(ctx, 150)
	if err != nil {
		fmt.Printf("failed to get activities. %+v\n", err)
		return
	}

	tickets := make(HashSet)
	for _, entry := range activities.Entries {
		s := entry.Activity.Title

		regex := regexp.MustCompile("^[a-zA-Z]+-[0-9]{1,3}")
		if match := regex.FindString(s); match != "" {
			tickets[match] = struct{}{}
		}
	}

	var input strings.Builder
	for s := range tickets {
		fmt.Printf("Fetching %-10s ticket information", s)
		issue, err := jira.GetIssue(ctx, s)
		if err != nil {
			fmt.Printf("failed to get issue from Jira. %+v", err)
			return
		}
		time.Sleep(time.Millisecond * 10)
		fmt.Print("\r\033[2K")

		desc := fmt.Sprintf("Ticket: %s\nSummary: \"%s\"\nDescription: \"%s\"", issue.Key, issue.Fields.Summary, issue.Fields.Description)
		input.WriteString(desc)
	}

	o := ollama.NewOllamaClient(ollama_host)
	hasModel, err := o.HasModel(ctx)
	if err != nil {
		fmt.Printf("failed to list active models. %+v\n", err)
		return
	}

	if !hasModel {
		s := "pulling model..."
		fmt.Printf("\n\n%s", s)
		err := o.PullModel(ctx)
		if err != nil {
			fmt.Printf("failed to pull model. %+v\n", err)
			return
		}

		fmt.Printf("\r%s\r", strings.Repeat(" ", len(s)))
	}

	thinking := "thinking..."
	fmt.Printf("\n%s", thinking)
	resp, err := o.GetResponse(ctx, input.String())
	if err != nil {
		fmt.Printf("failed to generate response. %+v\n", err)
		return
	}
	fmt.Printf("\r%s\r", strings.Repeat(" ", len(thinking)))

	fmt.Printf("\n\n\n\nResult:\n%s\n", resp)
}
