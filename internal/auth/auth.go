package auth

import (
	"os/exec"
	"strings"
)

// CmdRunner abstracts external command execution for testability.
type CmdRunner interface {
	Run(name string, args ...string) (string, error)
}

// ExecRunner executes real OS commands.
type ExecRunner struct{}

func (ExecRunner) Run(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).Output()
	return string(out), err
}

// Resolver resolves a GitHub authentication token.
type Resolver struct {
	Cmd    CmdRunner
	Getenv func(string) string
}

// ResolveToken returns a token using the priority: gh auth token > GITHUB_TOKEN > empty.
func (r *Resolver) ResolveToken() string {
	if out, err := r.Cmd.Run("gh", "auth", "token"); err == nil {
		if tok := strings.TrimSpace(out); tok != "" {
			return tok
		}
	}
	if tok := r.Getenv("GITHUB_TOKEN"); tok != "" {
		return tok
	}
	return ""
}
