package tool

import (
	"context"
	"encoding/json"
	"math"
	"testing"
)

func TestEvaluate(t *testing.T) {
	value, err := evaluate("(12.5+7.5)/2")
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(value-10) > 1e-9 {
		t.Fatalf("value = %v", value)
	}
}

func TestCalculatorRejectsUnknownFields(t *testing.T) {
	registered := CalculatorTool()
	_, err := registered.Handler(context.Background(), Principal{TenantID: "u", UserID: "u"}, json.RawMessage(`{"expression":"1+1","code":"bad"}`))
	if err == nil {
		t.Fatal("expected error")
	}
}
