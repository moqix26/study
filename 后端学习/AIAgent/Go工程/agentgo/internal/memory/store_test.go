package memory

import (
	"testing"

	"study.local/agentgo/internal/llm"
)

func TestWindowStoreIsolationAndLimit(t *testing.T) {
	store := NewWindowStore(2)
	if err := store.Append("tenant", "u1", "c1", llm.UserText("one")); err != nil {
		t.Fatal(err)
	}
	if err := store.Append("tenant", "u1", "c1", llm.AssistantText("two")); err != nil {
		t.Fatal(err)
	}
	if err := store.Append("tenant", "u1", "c1", llm.UserText("three")); err != nil {
		t.Fatal(err)
	}
	if err := store.Append("tenant", "u2", "c1", llm.UserText("private")); err != nil {
		t.Fatal(err)
	}

	history, err := store.History("tenant", "u1", "c1")
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 2 || history[0]["content"] != "two" || history[1]["content"] != "three" {
		t.Fatalf("history = %#v", history)
	}
	other, _ := store.History("tenant", "u2", "c1")
	if len(other) != 1 || other[0]["content"] != "private" {
		t.Fatalf("other = %#v", other)
	}
}
