package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Ollama struct {
	model        string
	url          string
	systemPrompt string
	httpClient   *http.Client
}

func NewOllamaClient(ollamaUrl string) *Ollama {

	systemPrompt := `
You are writing entries for a German apprenticeship logbook (Berichtsheft) for a software developer apprentice.

Your job is NOT to summarize Jira tickets.
Your job is to infer which development tasks were performed and describe them as professional work activities.

The input consists of weekly Jira issues.

Rules:

1. NEVER mention:
- Jira ticket IDs
- project names
- customer names
- company names
- internal identifiers

2. Use the issue SUMMARY as the primary source.
Ignore issue descriptions unless they clearly contain implementation details.
Ignore copied emails, customer communication, signatures, HTML and ticket history.

3. Rewrite every task into a generic software development activity.

Examples:

"Fehler Darstellung Krankheit buchen"
→ "Fehler bei der Buchung von Krankheitszeiten behoben."

"Report Urlaubsauswertung"
→ "Auswertungen und Berichte für Urlaubsdaten erweitert."

"Überarbeitung Schnittstelle Buchungsexport"
→ "Schnittstelle für den Export von Buchungsdaten überarbeitet."

"Design Fehler Personalstamm"
→ "Darstellungsfehler im Personalstamm behoben."

"Folgeticket zu ..."
→ Use the previous ticket title if possible, otherwise describe it as
"Weitere Verbesserungen und Fehlerkorrekturen umgesetzt."

4. Merge similar issues into one activity.

5. Focus on actual programming work:
- implementing features
- fixing bugs
- improving reports
- extending interfaces
- improving validation
- optimizing existing functionality

Avoid mentioning administrative work.

6. Write in German.

7. Write in the style of an apprenticeship logbook:
- factual
- concise
- professional
- past tense

Return ONLY markdown bullet points, and group by week as per the input.

Each bullet should begin with "- ".

Example output:

- Fehler in der Buchung von Krankheitszeiten behoben.
- Eine Schnittstelle für den Export von Buchungsdaten überarbeitet.
- Berichte zur Urlaubsauswertung erweitert.
- Darstellungsfehler im Personalstamm korrigiert.
- Validierungen und Buchungslogik verbessert.
`

	return &Ollama{
		model:        "llama3.2:latest",
		url:          ollamaUrl + "/api",
		systemPrompt: systemPrompt,
		httpClient:   http.DefaultClient,
	}
}

func (o Ollama) GetResponse(ctx context.Context, input string) (string, error) {

	reqBody := GenerateReq{
		Model:        o.model,
		Prompt:       input,
		SystemPrompt: o.systemPrompt,
		Stream:       false,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", o.url+"/generate", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to construct request. %w", err)
	}

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to do request to url %s. %w", o.url+"/generate", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var genResp GenerateResp
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&genResp)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response. %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("expected a different status code. %d. Body: %+v", resp.StatusCode, genResp)
	}

	return genResp.Response, nil
}

func (o Ollama) PullModel(ctx context.Context) error {

	reqBody := PullReq{
		Model:          o.model,
		Insecure:       false,
		StreamProgress: false,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", o.url+"/pull", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to construct request. %w", err)
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to do request to pull model. %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var pullResp PullResp
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&pullResp)
	if err != nil {
		return fmt.Errorf("failed to unmarshal pull model response. %w", err)
	}

	return nil
}

func (o Ollama) ListModels(ctx context.Context) ([]Model, error) {

	req, err := http.NewRequestWithContext(ctx, "GET", o.url+"/tags", http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to construct request. %w", err)
	}
	req.Header.Add("Accept", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to do request to url %s. %w", o.url+"/tags", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var tagResp TagResp
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&tagResp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal tag response. Response: , %w", err)
	}

	return tagResp.Models, nil
}

func (o Ollama) HasModel(ctx context.Context) (bool, error) {

	models, err := o.ListModels(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to list models. %w", err)
	}

	for _, model := range models {
		if model.Model == o.model {
			return true, nil
		}
	}

	return false, nil
}
