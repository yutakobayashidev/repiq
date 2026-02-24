package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunNoArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := Run([]string{}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for no args")
	}
	if !strings.Contains(stderr.String(), "Usage:") {
		t.Errorf("expected usage in stderr, got: %q", stderr.String())
	}
}

func TestRunVersion(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := Run([]string{"--version"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "repiq") {
		t.Errorf("expected version output, got: %q", stdout.String())
	}
}

func TestRunHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := Run([]string{"--help"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for --help")
	}
	if !strings.Contains(stderr.String(), "Usage:") {
		t.Errorf("expected usage in stderr, got: %q", stderr.String())
	}
}

func TestRunInvalidTarget(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := Run([]string{"nocolon"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for invalid target")
	}
}

func TestRunUnknownScheme(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := Run([]string{"unknown:thing"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for unknown scheme")
	}
}

func TestRunMultipleTargets(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := Run([]string{"unknown:a", "unknown:b"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for unknown schemes")
	}
}

func TestRunExclusiveFormatFlags(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := Run([]string{"--json", "--ndjson", "github:golang/go"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for multiple format flags")
	}
	if !strings.Contains(err.Error(), "specify only one") {
		t.Errorf("expected exclusive flag error, got: %v", err)
	}
}
