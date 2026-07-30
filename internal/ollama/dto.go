package ollama

type Model struct {
	Name  string `json:"name"`
	Model string `json:"model"`
	Size  int    `json:"size"`
}

type TagResp struct {
	Models []Model `json:"models"`
}

type PullReq struct {
	Model          string `json:"model"`
	Insecure       bool   `json:"insecure"`
	StreamProgress bool   `json:"stream"`
}

type PullResp struct {
	Status string `json:"status"`
}

type GenerateReq struct {
	Model        string `json:"model"`
	Prompt       string `json:"prompt"`
	Format       string `json:"format"`
	SystemPrompt string `json:"system"`
	Stream       bool   `json:"stream"`
}

type GenerateResp struct {
	Model    string `json:"model"`
	Response string `json:"response"`
	Done     bool   `json:"done"`
	Error    string `json:"error"`
}
