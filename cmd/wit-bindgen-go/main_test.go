// TODO: Remove the following go:build line when TinyGo supports Go 1.24.
//go:build !(tinygo && go1.24)

package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/kjanat/go-modules/internal/module"
)

// TestSimpleGenVerbosity ensures that a basic generation case honors the verbose flag
func TestSimpleGenVerbosity(t *testing.T) {
	inWIT := `package tests:test;`
	stdin := strings.NewReader(inWIT)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := Command
	cmd.Reader = stdin
	cmd.Writer = &stdout
	cmd.ErrWriter = &stderr

	args := []string{
		"wit-bindgen-go",
		"generate",
		"--world",
		"http-fetch-simple",
		"--dry-run",
		"-",
	}

	// Run the app without verbose
	cmd.Run(context.Background(), args)
	if stderr.Len() != 0 {
		t.Errorf("output was written to stderr despite lack of --verbose")
	}

	stdin.Reset(inWIT)
	stdout.Reset()
	stderr.Reset()

	cmd.Run(context.Background(), append(args, "-v"))
	if stderr.Len() == 0 {
		t.Errorf("no output was written to stderr when -v was used")
	}

	cmd.Run(context.Background(), append(args, "-vv"))
	if stderr.Len() == 0 {
		t.Errorf("no output was written to stderr when -vv was used")
	}

	cmd.Run(context.Background(), append(args, "--verbose"))
	if stderr.Len() == 0 {
		t.Errorf("no output was written to stderr when --verbose was used")
	}
}

func TestVersionText(t *testing.T) {
	saved := version
	t.Cleanup(func() { version = saved })
	for _, tt := range []struct{ set, want string }{
		{"", module.Version()},
		{"0.11.1", "v0.11.1"},
		{"v0.11.1", "v0.11.1"},
	} {
		version = tt.set
		if got := versionText(); got != tt.want {
			t.Errorf("version=%q: got %q, want %q", tt.set, got, tt.want)
		}
	}
}
