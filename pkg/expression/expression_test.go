package expression

import (
	"testing"
)

func TestEvaluateSimplePath(t *testing.T) {
	ctx := map[string]any{
		"context": map[string]any{"id": "abc-123"},
	}
	got, err := Evaluate("${context.id}", ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got != "abc-123" {
		t.Errorf("got %v, want abc-123", got)
	}
}

func TestEvaluateNestedPath(t *testing.T) {
	ctx := map[string]any{
		"state": map[string]any{
			"filters": map[string]any{"month": "July"},
		},
	}
	got, err := Evaluate("${state.filters.month}", ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got != "July" {
		t.Errorf("got %v, want July", got)
	}
}

func TestEvaluateStringInterpolation(t *testing.T) {
	ctx := map[string]any{
		"user": map[string]any{"name": "Alice"},
	}
	got, err := Evaluate("Hello ${user.name}!", ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got != "Hello Alice!" {
		t.Errorf("got %v, want 'Hello Alice!'", got)
	}
}

func TestEvaluateNoExpressions(t *testing.T) {
	got, err := Evaluate("plain text", nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != "plain text" {
		t.Errorf("got %v, want 'plain text'", got)
	}
}

func TestEvaluateMissingPath(t *testing.T) {
	ctx := map[string]any{}
	_, err := Evaluate("${missing.key}", ctx)
	if err == nil {
		t.Fatal("expected error for missing path")
	}
}

func TestEvaluatePreservesType(t *testing.T) {
	ctx := map[string]any{"count": 42}
	got, err := Evaluate("${count}", ctx)
	if err != nil {
		t.Fatal(err)
	}
	num, ok := got.(int)
	if !ok {
		t.Fatalf("expected int, got %T", got)
	}
	if num != 42 {
		t.Errorf("got %d, want 42", num)
	}
}

func TestEvaluatePreservesTypeBool(t *testing.T) {
	ctx := map[string]any{"enabled": true}
	got, err := Evaluate("${enabled}", ctx)
	if err != nil {
		t.Fatal(err)
	}
	b, ok := got.(bool)
	if !ok {
		t.Fatalf("expected bool, got %T", got)
	}
	if !b {
		t.Error("got false, want true")
	}
}

func TestEvaluateArrayIndex(t *testing.T) {
	ctx := map[string]any{
		"items": []any{
			map[string]any{"name": "first"},
			map[string]any{"name": "second"},
		},
	}
	got, err := Evaluate("${items.0.name}", ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got != "first" {
		t.Errorf("got %v, want 'first'", got)
	}
}

func TestEvaluateArrayIndexOutOfRange(t *testing.T) {
	ctx := map[string]any{
		"items": []any{"a"},
	}
	_, err := Evaluate("${items.5}", ctx)
	if err == nil {
		t.Fatal("expected error for out-of-range index")
	}
}

func TestEvaluateMultipleExpressions(t *testing.T) {
	ctx := map[string]any{
		"first": "John",
		"last":  "Doe",
	}
	got, err := Evaluate("${first} ${last}", ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got != "John Doe" {
		t.Errorf("got %v, want 'John Doe'", got)
	}
}

func TestContainsExpression(t *testing.T) {
	if !ContainsExpression("${foo}") {
		t.Error("expected true for ${foo}")
	}
	if !ContainsExpression("prefix ${foo} suffix") {
		t.Error("expected true for mixed string")
	}
	if ContainsExpression("no expressions here") {
		t.Error("expected false for plain text")
	}
	if ContainsExpression("$notbrace") {
		t.Error("expected false for $ without braces")
	}
}

func TestExtractPaths(t *testing.T) {
	paths := ExtractPaths("${user.name} said ${message.text}")
	if len(paths) != 2 {
		t.Fatalf("expected 2 paths, got %d", len(paths))
	}
	if paths[0] != "user.name" {
		t.Errorf("paths[0] = %q, want user.name", paths[0])
	}
	if paths[1] != "message.text" {
		t.Errorf("paths[1] = %q, want message.text", paths[1])
	}
}

func TestExtractPathsNone(t *testing.T) {
	paths := ExtractPaths("plain text")
	if len(paths) != 0 {
		t.Errorf("expected 0 paths, got %d", len(paths))
	}
}

func TestEvaluateWhitespace(t *testing.T) {
	ctx := map[string]any{"x": "val"}
	got, err := Evaluate("${ x }", ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got != "val" {
		t.Errorf("got %v, want val", got)
	}
}
