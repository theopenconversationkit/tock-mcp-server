package config

import (
	"strings"
	"testing"
)

func TestServerConfig_ToolDescriptionFields(t *testing.T) {
	cfg, err := Load("../config.tock-rag.yaml")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Vérifier tool_name
	if cfg.Server.ToolName != "ask_tock_rag" {
		t.Fatalf("expected ToolName='ask_tock_rag', got %q", cfg.Server.ToolName)
	}

	// Vérifier tool_description (doit contenir du texte spécifique)
	if !strings.Contains(cfg.Server.ToolDescription, "RAG") {
		t.Fatalf("expected ToolDescription to contain 'RAG', got %q", cfg.Server.ToolDescription)
	}

	// Vérifier input_question_description
	if cfg.Server.InputQuestionDescription == "" {
		t.Fatalf("expected InputQuestionDescription to be non-empty")
	}
	if !strings.Contains(cfg.Server.InputQuestionDescription, "Question technique") {
		t.Fatalf("expected InputQuestionDescription to contain 'Question technique', got %q", cfg.Server.InputQuestionDescription)
	}
}
