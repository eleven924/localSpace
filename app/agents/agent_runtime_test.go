package agents

import (
	"context"
	"testing"

	"LocalSpace/app/models"
	"LocalSpace/app/tools"
)

type fakeRuntime struct {
	response *AgentRunResponse
	err      error
}

func (f *fakeRuntime) Run(ctx context.Context, req *AgentRunRequest) (*AgentRunResponse, error) {
	return f.response, f.err
}

type fakeTool struct{ name string }

func (f *fakeTool) Name() string { return f.name }
func (f *fakeTool) Description() string { return "fake" }
func (f *fakeTool) Execute(ctx context.Context, input string) (string, error) { return "ok", nil }

func TestEinoMetadataAgentAnalyze_ParsesStructuredOutput(t *testing.T) {
	agent := NewEinoMetadataAgent(&fakeRuntime{
		response: &AgentRunResponse{
			Output:        `{"tags":["video","action"],"description":"动作影片"}`,
			ToolsUsed:     []string{"web_search"},
			SearchQueries: []string{"movie mp4"},
		},
	})

	result, err := agent.Analyze(context.Background(), &MetadataGenerationInput{FileName: "movie.mp4", FileType: "video"}, &models.AIConfig{Enabled: true, EnableAgent: true}, []tools.Tool{&fakeTool{name: "web_search"}})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if len(result.Analysis.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(result.Analysis.Tags))
	}

	if result.Trace.ToolsUsed[0] != "web_search" {
		t.Fatalf("expected tool usage to be recorded, got %v", result.Trace.ToolsUsed)
	}
}

func TestEinoMetadataAgentAnalyze_InvalidOutput(t *testing.T) {
	agent := NewEinoMetadataAgent(&fakeRuntime{response: &AgentRunResponse{Output: "not-json"}})

	if _, err := agent.Analyze(context.Background(), &MetadataGenerationInput{FileName: "movie.mp4", FileType: "video"}, &models.AIConfig{Enabled: true, EnableAgent: true}, nil); err == nil {
		t.Fatal("expected invalid output to return error")
	}
}
