package ollama_test

import (
	"context"
	"testing"

	"berichtsheft/internal/logging"
	"berichtsheft/internal/ollama"
)

func TestGetResponse(t *testing.T) {

	closer, err := logging.SetupDefault(logging.Test)
	if err != nil {
		t.Fatalf("failed to setup logging. %+v", err)
	}
	defer closer.Close()

	ctx := context.Background()

	o := ollama.NewOllamaClient("http://localhost:11434")
	hasModel, err := o.HasModel(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if !hasModel {
		err := o.PullModel(ctx)
		if err != nil {
			t.Fatalf("failed to pull model. %+v", err)
		}
	}

	resp, err := o.GetResponse(ctx, "Why is the sky blue?")
	if err != nil {
		t.Fatalf("received error on response. %+v", err)
	}
	t.Logf("response: %s", resp)

}
