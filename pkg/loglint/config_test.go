package loglint

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestResolveConfigFromFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "loglint.json")
	content := []byte(`{
		"check_special_chars": false,
		"check_sensitive_data": false,
		"sensitive_keywords": ["session"]
	}`)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	analyzer := NewAnalyzer(nil)
	if err := analyzer.Flags.Set("config", path); err != nil {
		t.Fatalf("set flag: %v", err)
	}

	cfg, err := resolveConfig(analyzer, nil)
	if err != nil {
		t.Fatalf("resolve config: %v", err)
	}

	if cfg.CheckSpecialChars {
		t.Fatalf("expected check_special_chars to be disabled")
	}
	if cfg.CheckSensitiveData {
		t.Fatalf("expected check_sensitive_data to be disabled")
	}
	if !reflect.DeepEqual(cfg.SensitiveKeywords, []string{"session"}) {
		t.Fatalf("unexpected keywords: %#v", cfg.SensitiveKeywords)
	}
	if !cfg.CheckLowercaseStart || !cfg.CheckEnglishOnly {
		t.Fatalf("unrelated defaults should remain enabled: %#v", cfg)
	}
}

func TestConfigFromAnyMergesFileAndInline(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "loglint.json")
	content := []byte(`{
		"check_special_chars": false,
		"sensitive_keywords": ["session"]
	}`)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	raw := map[string]any{
		"config":               path,
		"check_sensitive_data": false,
		"sensitive_keywords":   []any{"token", "secret"},
	}

	cfg, err := ConfigFromAny(raw)
	if err != nil {
		t.Fatalf("parse config: %v", err)
	}

	if cfg == nil {
		t.Fatalf("expected config")
	}
	if cfg.CheckSpecialChars == nil || *cfg.CheckSpecialChars {
		t.Fatalf("expected file config to disable check_special_chars")
	}
	if cfg.CheckSensitiveData == nil || *cfg.CheckSensitiveData {
		t.Fatalf("expected inline config to disable check_sensitive_data")
	}
	if !reflect.DeepEqual(cfg.SensitiveKeywords, []string{"token", "secret"}) {
		t.Fatalf("unexpected keywords: %#v", cfg.SensitiveKeywords)
	}
}

func TestLowerFirstASCIILetter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		out  string
		ok   bool
	}{
		{name: "uppercase", in: "Starting server", out: "starting server", ok: true},
		{name: "leading spaces", in: "  Starting server", out: "  starting server", ok: true},
		{name: "already lowercase", in: "starting server", out: "", ok: false},
		{name: "non english", in: "Запуск сервера", out: "", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, ok := lowerFirstASCIILetter(tt.in)
			if ok != tt.ok {
				t.Fatalf("ok mismatch: got %v want %v", ok, tt.ok)
			}
			if out != tt.out {
				t.Fatalf("out mismatch: got %q want %q", out, tt.out)
			}
		})
	}
}
