package analytics

import (
	"context"
	"strings"
	"testing"
)

func TestResolveDSNEnv(t *testing.T) {
	resolver, err := NewDefaultResolver()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("UIFORGE_TEST_DSN", "root:@tcp(127.0.0.1:13307)/omniroadmap")

	dsn, err := ResolveDSN(context.Background(), resolver, "env://UIFORGE_TEST_DSN")
	if err != nil {
		t.Fatal(err)
	}
	if dsn != "root:@tcp(127.0.0.1:13307)/omniroadmap" {
		t.Fatalf("unexpected dsn %q", dsn)
	}
}

func TestResolveDSNErrors(t *testing.T) {
	resolver, err := NewDefaultResolver()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	tests := []struct {
		name    string
		ref     string
		wantSub string
	}{
		{"raw mysql dsn", "root:@tcp(127.0.0.1:13307)/db", "secret reference"},
		{"url-style raw dsn", "mysql://root:pass@127.0.0.1/db", "no secret provider registered"},
		{"missing env var", "env://UIFORGE_TEST_DSN_DOES_NOT_EXIST", "resolving"},
		{"empty value", "env://UIFORGE_TEST_EMPTY_DSN", "empty value"},
	}
	t.Setenv("UIFORGE_TEST_EMPTY_DSN", "   ")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ResolveDSN(ctx, resolver, tt.ref)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantSub)
			}
			if !strings.Contains(err.Error(), tt.wantSub) {
				t.Fatalf("expected error containing %q, got %q", tt.wantSub, err.Error())
			}
		})
	}

	if _, err := ResolveDSN(ctx, nil, "env://X"); err == nil {
		t.Fatal("expected error for nil resolver")
	}
}
