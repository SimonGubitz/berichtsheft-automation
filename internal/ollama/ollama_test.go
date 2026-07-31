package ollama_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"berichtsheft/internal/ollama"
)

func NewOllamaMockServer() *httptest.Server {
	mux := http.NewServeMux()

	var modelsLock sync.Mutex
	models := make([]ollama.Model, 0)

	mux.HandleFunc("POST /api/generate", func(w http.ResponseWriter, req *http.Request) {
		encoder := json.NewEncoder(w)

		var reqBody ollama.GenerateReq
		decoder := json.NewDecoder(req.Body)
		err := decoder.Decode(&reqBody)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = encoder.Encode(ollama.GenerateResp{
				Error: "failed to read request body",
			})
		}
		defer func() { _ = req.Body.Close() }()

		_ = encoder.Encode(ollama.GenerateResp{
			Model: reqBody.Model,
		})
	})

	// WARN: does not implement the `StreamProgress` option
	mux.HandleFunc("POST /api/pull", func(w http.ResponseWriter, req *http.Request) {
		encoder := json.NewEncoder(w)

		var reqBody ollama.PullReq
		decoder := json.NewDecoder(req.Body)
		err := decoder.Decode(&reqBody)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = encoder.Encode(ollama.PullResp{
				Status: "error",
			})
		}
		defer func() { _ = req.Body.Close() }()

		modelsLock.Lock()
		models = append(models, ollama.Model{Name: reqBody.Model, Model: reqBody.Model, Size: 123})
		modelsLock.Unlock()

		_ = encoder.Encode(ollama.PullResp{
			Status: "success",
		})
	})

	mux.HandleFunc("GET /api/tags", func(w http.ResponseWriter, req *http.Request) {
		encoder := json.NewEncoder(w)
		modelsLock.Lock()
		_ = encoder.Encode(ollama.TagResp{
			Models: models,
		})
		modelsLock.Unlock()
	})

	return httptest.NewServer(mux)
}

func TestOllamaUnit(t *testing.T) {

	ctx := context.Background()

	server := NewOllamaMockServer()
	o := ollama.NewOllamaClient(server.URL)

	hasModel, err := o.HasModel(ctx)
	if err != nil {
		t.Fatalf("failed to list active models. %+v", err)
	}

	if !hasModel {
		err := o.PullModel(ctx)
		if err != nil {
			t.Fatalf("failed to pull model. %+v", err)
		}
	}

	_, err = o.GetResponse(ctx, "test input")
	if err != nil {
		t.Fatalf("failed to generate response. %+v", err)
	}
}
